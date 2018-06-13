package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/procwatcher"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

// MPIConf implements ControllerConf interface
type MPIConf struct {
	//////////////////////////////////////////////////////
	// All controller confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID      string   `json:"machine-id"`
	ID             string   `json:"id"` // IMPORTANT: ID should be unique for across all nodes in the RF24Network
	ControllerType string   `json:"controller-type"`
	Subscriptions  []string `json:"subscriptions"` // Message Type Subscriptions.

	///////////////////////////////////////////////
	// The fields below are controller specific. //
	///////////////////////////////////////////////
	PluginName string      `json:"plugin-name"` // Name of the plugin.
	PluginConf interface{} `json:"plugin-conf"` // Plugin spcecific conf.
	KeepAlive  bool        `json:"keepalive"`   // Should the process be kept alive

	GRPCConf controllerapi.GRPCConf `json:"grpc-conf"` // GRPC server conf
}

func (c MPIConf) ValidateConf() error {
	if c.ControllerType == "" {
		return fmt.Errorf("no controller type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.ControllerType != Controller_MPI {
		return fmt.Errorf("invalid controller type specified, expected %s, but got %s", Controller_MPI, c.ControllerType)
	}
	if c.PluginName == "" {
		return fmt.Errorf("no plugin specified")
	}
	if c.GRPCConf.HostAddress == "" {
		return fmt.Errorf("no hostname specified in configuration")
	}
	if c.GRPCConf.Port < 0 || c.GRPCConf.Port > 65535 {
		return fmt.Errorf("invalid port %d specified in configuration", c.GRPCConf.Port)
	}
	if c.GRPCConf.TLSEnabled {
		if c.GRPCConf.CertFile == "" || c.GRPCConf.KeyFile == "" {
			return fmt.Errorf("must specifiy both a cert file and a key file when enabling TLS")
		}
	}
	return nil
}

func (c MPIConf) GetType() string {
	return c.ControllerType
}

func (c MPIConf) GetID() string {
	return c.ID
}

func (c MPIConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c MPIConf) NewController(rcvQ *message.Queue, sndQ *message.Queue) (controllerapi.Controller, error) {

	drv := MPI{
		State: &MPIInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}

	return &drv, nil
}

// MPI implements the Controller interface
type MPI struct {
	mu    sync.RWMutex
	State *MPIInternal
}

type MPIInternal struct {
	Conf     MPIConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}

	procMaster *procwatcher.ProcMaster   // Pointer to the one and only master.
	proc       procwatcher.ProcContainer // proc is the plugin process.

	grpcServer *grpc.Server
}

// Start: starts the controller logic.
func (d *MPI) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// We will start the grpc server and process messages from plugin
	go d.work()

	// Write the plugin conf to a file; the file will be passed in via stdin
	// pluginConfMap := d.State.Conf.PluginConf.(map[string]interface{})
	confBytes, err := json.Marshal(d.State.Conf.PluginConf)
	if err != nil {
		return err
	}
	// We will also write the grpc conf on a separate line
	grpcConfBytes, err := json.Marshal(d.State.Conf.GRPCConf)
	if err != nil {
		return err
	}
	inFilePath := buildPluginInFilePath(d.State.Conf.MachineID, d.State.Conf.ID)
	f, err := os.Create(inFilePath)
	// err = common.WriteFile(inFilePath, confBytes)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(confBytes)
	if err != nil {
		return err
	}
	_, _ = f.WriteString("\n")
	_, err = f.Write(grpcConfBytes)
	if err != nil {
		return err
	}
	f.Sync()

	d.State.procMaster = procwatcher.GetProcMaster()
	d.State.proc = &procwatcher.Proc{
		Name: buildPluginProcessName(d.State.Conf.MachineID, d.State.Conf.ID), // must be unique
		Cmd:  filepath.Join(common.SeguePluginsPath, d.State.Conf.PluginName),
		// Args:      args,
		Pidfile:   buildPluginPidFilePath(d.State.Conf.MachineID, d.State.Conf.ID),
		Infile:    inFilePath,
		Outfile:   buildPluginOutFilePath(d.State.Conf.MachineID, d.State.Conf.ID),
		Errfile:   buildPluginErrFilePath(d.State.Conf.MachineID, d.State.Conf.ID),
		KeepAlive: d.State.Conf.KeepAlive,
		Status:    &procwatcher.ProcStatus{},
		Sys:       &syscall.SysProcAttr{Setpgid: true}, // See note below on setpgid
	}

	var pid int
	// If a pid file exists, we'll try to restore
	pid, err = readPidFromFile(buildPluginPidFilePath(d.State.Conf.MachineID, d.State.Conf.ID))
	if pid > 0 && common.GetProcess(pid) != nil {
		// Matra restarted, try to recover Plugin proc.
		p := common.GetProcess(pid)
		if p == nil {
			logger.Errorf("Could not recover pid %d in MPI controller", pid)
			return fmt.Errorf("Could not recover pid %d in MPI controller", pid)
		}
		d.State.proc.Restore(p)
		d.State.procMaster.Restore(d.State.proc)
		logger.Infof("Restored plugin %s with pid %d", d.State.Conf.PluginName, d.State.proc.GetPid())
	} else {
		// Start new plugin process.
		logger.Debugf("Starting plugin process...")
		err = d.State.procMaster.StartProcess(d.State.proc)
		if err != nil {
			return err
		}

		logger.Infof("Started plugin with pid %d", d.State.proc.GetPid())
	}

	return nil
}

// Stop: stops the controller logic.
func (d *MPI) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)

	err := d.State.procMaster.StopProcess(d.State.proc.Identifier())
	if err != nil {
		logger.Errorf("Error stopping plugin process (%d): %s", d.State.proc.GetPid(), err)
		// We log the error but keep executing Stop()
	}

	if d.State.grpcServer != nil {
		d.State.grpcServer.Stop()
	}
	return nil
}

// work: processes messages to be sent out and listens for incoming messages.
func (d *MPI) work() {

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", d.State.Conf.GRPCConf.HostAddress, d.State.Conf.GRPCConf.Port))
	if err != nil {
		logger.Errorf("failed to listen (%s:%d): %v", d.State.Conf.GRPCConf.HostAddress, d.State.Conf.GRPCConf.Port, err)
		return
	}
	var opts []grpc.ServerOption
	if d.State.Conf.GRPCConf.TLSEnabled {
		logger.Infof("grpc: starting tls server")
		creds, err := credentials.NewServerTLSFromFile(d.State.Conf.GRPCConf.CertFile, d.State.Conf.GRPCConf.KeyFile)
		if err != nil {
			logger.Errorf("Failed to generate credentials %v", err)
			return
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	d.State.grpcServer = grpc.NewServer(opts...)
	seguepb.RegisterMessengerServer(d.State.grpcServer, d)
	d.State.grpcServer.Serve(lis) // Stop will ensure this go routine exists

}

func (d *MPI) GetConf() controllerapi.ControllerConf {
	return d.State.Conf
}

func (d *MPI) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *MPI) Copy() controllerapi.Controller {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &MPI{
		State: &MPIInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *MPI) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *MPI) UnmarshalJSON(data []byte) error {
	d.State = &MPIInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}

func buildPluginProcessName(machineID, id string) string {
	return "segue-plugin-" + machineID + "-" + id
}

func buildPluginPidFilePath(machineID, id string) string {
	return filepath.Join(common.SeguePath, machineID, id+"-pid.file")
}

func buildPluginInFilePath(machineID, id string) string {
	return filepath.Join(common.SeguePath, machineID, id+"-in.file")
}

func buildPluginOutFilePath(machineID, id string) string {
	return filepath.Join(common.SeguePath, machineID, id+"-out.file")
}

func buildPluginErrFilePath(machineID, id string) string {
	return filepath.Join(common.SeguePath, machineID, id+"-err.file")
}

func readPidFromFile(filePath string) (int, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return -1, err
	}
	var pid int
	_, err = fmt.Fscanf(fd, "%d\n", &pid)
	if err != nil {
		if err == io.EOF {
			return pid, nil
		}
		return -1, err
	}
	return pid, nil
}

func (d *MPI) Messenger(stream seguepb.Messenger_MessengerServer) error {

	// Send routine
	go func() {
		defer func() {
			logger.Debugf("mpi-grpc: stoping messenger-send routine")
		}()

		logger.Debugf("mpi-grpc: starting messenger-send routine")
		// Read all the messages in the stream send queue every X
		// milliseconds. All the messages will be put in the envelope.
		// The idea being that we'll minimize network stack overhead.
		// However, this approach introduces latency; alternatives include
		// reducing the interval or just sending a message as soon as it arrives.
		tickChan := time.NewTicker(time.Millisecond * 100).C
		for {
			select {
			case <-d.State.killChan:
				return
			case <-tickChan:
				if d.State.rcvQ.IsShuttingDown() {
					return
				}
				var eMsgs []*seguepb.Message
				// NOTE: this assumes only a single consumer of the queue.
				sqLen := d.State.rcvQ.Len() // TODO: may want to cap this to say 25 msgs
				if sqLen == 0 {
					continue
				}
				logger.Debugf("mpi-grpc: sending %d messages", sqLen)
				for i := 0; i < sqLen; i++ {
					iMsg, shutdown := d.State.rcvQ.Get()
					if shutdown {
						logger.Debugf("stopping sender routine for mpi-grpc-stream-send-queue-%s", d.State.Conf.ID)
						return
					}
					eMsg, err := message.ConvertToExternalFormat(iMsg)
					if err != nil {
						logger.Errorf(fmt.Sprintf("error marshaling data in routine for mpi-grpc-stream-send-queue-%s: %s", d.State.Conf.ID, err))
					} else {
						eMsgs = append(eMsgs, eMsg)
					}
					d.State.rcvQ.Done(iMsg)
				}
				msgEnv := &seguepb.MessageEnvelope{eMsgs}
				if err := stream.Send(msgEnv); err != nil {
					logger.Errorf("error while sending on mpi-grpc-stream-send-queue-%s (exiting send routine): %s", d.State.Conf.ID, err)
					return
				}
			}
		}
	}()
	logger.Debugf("mpi-grpc: starting messenger-receive routine")
	for {
		select {
		case <-d.State.killChan:
			return nil
		default:
			msgPBEnv, err := stream.Recv()
			if err == io.EOF {
				logger.Debugf("mpi-grpc: exiting Messanger() EOF recieved")
				return nil
			}
			if err != nil {
				logger.Errorf("mpi-grpc: error: %s", err)
				return err
			}
			for _, msgPB := range msgPBEnv.Messages {
				// We go through the messages in the envelope, converting them into
				// the internal message format and sending them to the internal
				// routing system.
				msg, err := message.ConvertToInternalFormat(msgPB)
				if err != nil {
					logger.Errorf("mpi-grpc: messge convertion error: %s", err)
				} else {
					logger.Debugf("mpi-grpc: received msg %v", msg)
					d.State.sndQ.Add(msg)
				}
			}
		}
	}
}
