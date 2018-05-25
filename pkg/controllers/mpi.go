package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/procwatcher"
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

}

// Start: starts the controller logic.
func (d *MPI) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Write the plugin conf to a file; the file will be passed in via stdin
	// pluginConfMap := d.State.Conf.PluginConf.(map[string]interface{})
	confBytes, err := json.Marshal(d.State.Conf.PluginConf)
	if err != nil {
		return err
	}

	inFilePath := buildPluginInFilePath(d.State.Conf.MachineID, d.State.Conf.ID)
	err = common.WriteFile(inFilePath, confBytes)
	if err != nil {
		return err
	}

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

	// os.Remove(buildpluginSockPath(containerID)) // remove old sock if it exists
	// s.sockConn, err = createpluginSocket(containerID)
	// if err != nil {
	// 	return nil, err
	// }

	// go d.work()

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
		return err
	}
	return nil
}

// work: processes messages to be sent out and listens for incoming messages on RF.
func (d *MPI) work() {

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
