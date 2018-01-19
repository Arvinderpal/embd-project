package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

// GRPCMessageServerConf implements programapi.ProgramConf interface
type GRPCMessageServerConf struct {
	//////////////////////////////////////////////////////
	// All controller confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID      string   `json:"machine-id"`
	ID             string   `json:"id"`
	ControllerType string   `json:"controller-type"`
	Subscriptions  []string `json:"subscriptions"` // Message Type Subscriptions.

	////////////////////////////////////////////
	// The fields below are controller specific. //
	////////////////////////////////////////////
	LogFilePathname string `json:"log-file-path-name"` // logs will be wirten to this file.

	HostAddress string `json:"host-address"` // IP address or just localhost
	Port        int    `json:"port"`         // Port to listen on

	TLSEnabled bool   `json:"tls-enabled"` // Will enable TLS on server
	CertFile   string `json:"cert-file"`   // TLS certfile (optional)
	KeyFile    string `json:"key-file"`    // TLS Key (optional)
}

func (c GRPCMessageServerConf) ValidateConf() error {
	if c.ControllerType == "" {
		return fmt.Errorf("no controller type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.ControllerType != Controller_GRPCMessageServer {
		return fmt.Errorf("invalid controller type specified, expected %s, but got %s", Controller_GRPCMessageServer, c.ControllerType)
	}
	if c.HostAddress == "" {
		return fmt.Errorf("no hostname specified in configuration")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d specified in configuration", c.Port)
	}
	if c.TLSEnabled {
		if c.CertFile == "" || c.KeyFile == "" {
			return fmt.Errorf("must specifiy both a cert file and a key file when enabling TLS")
		}
	}
	return nil
}

func (c GRPCMessageServerConf) GetType() string {
	return c.ControllerType
}

func (c GRPCMessageServerConf) GetID() string {
	return c.ID
}

func (c GRPCMessageServerConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c GRPCMessageServerConf) NewController(rcvQ *message.Queue, sndQ *message.Queue) (controllerapi.Controller, error) {

	drv := GRPCMessageServer{
		State: &grpcMessageServerInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}
	return &drv, nil
}

// GRPCMessageServer implements the Controller interface
type GRPCMessageServer struct {
	mu    sync.RWMutex
	State *grpcMessageServerInternal
}

type grpcMessageServerInternal struct {
	Conf     GRPCMessageServerConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}

	grpcServer *grpc.Server
}

// Start: starts the controller logic.
func (d *GRPCMessageServer) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.work()
	return nil
}

// Stop: stops the controller logic.
func (d *GRPCMessageServer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)
	if d.State.grpcServer != nil {
		d.State.grpcServer.Stop()
	}
	return nil
}

// work: starts a grpc server and processes requests and responses.
func (d *GRPCMessageServer) work() {

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", d.State.Conf.HostAddress, d.State.Conf.Port))
	if err != nil {
		logger.Errorf("failed to listen (%s:%d): %v", d.State.Conf.HostAddress, d.State.Conf.Port, err)
		return
	}
	var opts []grpc.ServerOption
	if d.State.Conf.TLSEnabled {
		logger.Infof("grpc-message-server: starting tls server")
		creds, err := credentials.NewServerTLSFromFile(d.State.Conf.CertFile, d.State.Conf.KeyFile)
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

func (d *GRPCMessageServer) GetConf() controllerapi.ControllerConf {
	return d.State.Conf
}

func (d *GRPCMessageServer) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *GRPCMessageServer) Copy() controllerapi.Controller {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &GRPCMessageServer{
		State: &grpcMessageServerInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *GRPCMessageServer) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *GRPCMessageServer) UnmarshalJSON(data []byte) error {
	d.State = &grpcMessageServerInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}

func (d *GRPCMessageServer) Messenger(stream seguepb.Messenger_MessengerServer) error {
	logger.Debugf("grpc-message-server: messenger started")
	for {
		msgPBEnv, err := stream.Recv()
		if err == io.EOF {
			logger.Debugf("grpc-message-server: exiting Messanger() EOF recieved")
			return nil
		}
		if err != nil {
			logger.Errorf("grpc-message-server: error: %s", err)
			return err
		}
		for _, msgPB := range msgPBEnv.Messages {
			// We go through the messages in the envelope, converting them into
			// the internal message format and sending them to the internal
			// routing system.
			msg, err := message.ConvertToInternalFormat(msgPB)
			if err != nil {
				logger.Errorf("messenger: messge convertion error: %s", err)
				// continue
			} else {
				d.State.sndQ.Add(msg)
			}
		}

		// TODO: if desired by the client, we should send messages that this controller has subscribed to.
		// for _, note := range rn {
		// 	if err := stream.Send(note); err != nil {
		// 		return err
		// 	}
		// }
	}
}
