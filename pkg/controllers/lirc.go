package controllers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/Arvinderpal/embd-project/pkg/controllers/lircbackend"
)

// README:
// It's important that the lirc module is installed and running before
// starting this controller. Aditionally, the remote's conf file needs to be
// placed in /etc/lirc/lircd.conf.d and loaded by lirc. The conf file
// should define the codes. For example, here is the snippet of the conf
// file for the Sunter Fan Remote:
// begin codes
//     BTN_START                0x02FD
//     BTN_LEFT                 0x22DD
//     BTN_RIGHT                0xC23D
//     BTN_TOP                  0x629D
//     BTN_BACK                 0xA857
// end codes

// LIRCConf implements ControllerConf interface
type LIRCConf struct {
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
	LIRCSocketPath string `json:"lirc-socket"` // Default: /var/run/lirc/lircd
}

func (c LIRCConf) ValidateConf() error {
	if c.ControllerType == "" {
		return fmt.Errorf("no controller type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.ControllerType != Controller_LIRC {
		return fmt.Errorf("invalid controller type specified, expected %s, but got %s", Controller_LIRC, c.ControllerType)
	}
	return nil
}

func (c LIRCConf) GetType() string {
	return c.ControllerType
}

func (c LIRCConf) GetID() string {
	return c.ID
}

func (c LIRCConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c LIRCConf) NewController(rcvQ *message.Queue, sndQ *message.Queue) (controllerapi.Controller, error) {

	if c.LIRCSocketPath == "" {
		c.LIRCSocketPath = "/var/run/lirc/lircd"
	}
	drv := LIRC{
		State: &lircInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}
	return &drv, nil
}

// LIRC implements the Controller interface
type LIRC struct {
	mu    sync.RWMutex
	State *lircInternal
}

type lircInternal struct {
	Conf     LIRCConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}

	irRouter   *lircbackend.Router
	eventCount uint64
}

// Start: starts the controller logic.
func (d *LIRC) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	var err error
	// Initialize with path to lirc socket
	d.State.irRouter, err = lircbackend.Init(d.State.Conf.LIRCSocketPath)
	if err != nil {
		return fmt.Errorf("failed to init lirc socket (%s): %s", d.State.Conf.LIRCSocketPath, err)
	}

	d.State.irRouter.SetDefaultHandle(d.lircReceiveEventHandler)
	// d.State.irRouter.Handle("", "", d.lircReceiveEventHandler)

	go d.work()
	return nil
}

// Stop: stops the controller logic.
func (d *LIRC) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)
	return nil
}

// work: starts listening on the lirc socket and sends IR commands.
func (d *LIRC) work() {

	// Get internal messages from rcvQ and send them via LIRC
	go func() {
		for {
			select {
			case <-d.State.killChan:
				return
			default:
				msg, shutdown := d.State.rcvQ.Get()
				if shutdown {
					logger.Debugf("stopping send routine in worker on controller %s", d.State.Conf.GetID())
					return
				}

				d.mu.RLock()
				logger.Debugf("lirc: (internal) message received: %v", msg)

				//TODO: add send()

				d.State.rcvQ.Done(msg)
				d.mu.RUnlock()
			}
		}
	}()

	// run the receive service
	d.State.irRouter.Run(d.State.killChan)

}

func (d *LIRC) lircReceiveEventHandler(event lircbackend.Event) {
	logger.Infof("%v\n", event)
	lircMsg := message.Message{
		ID: seguepb.Message_MessageID{
			Type:      seguepb.MessageType_LIRCEvent,
			SubType:   "",
			Version:   d.State.eventCount,
			Qualifier: "",
		},
		Data: &seguepb.LIRCEventData{
			Code:   event.Code,
			Repeat: int64(event.Repeat),
			Button: event.Button,
			Remote: event.Remote,
		},
	}
	d.mu.Lock()
	d.State.eventCount += 1
	d.State.sndQ.Add(lircMsg)
	d.mu.Unlock()
	return
}

func (d *LIRC) GetConf() controllerapi.ControllerConf {
	return d.State.Conf
}

func (d *LIRC) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *LIRC) Copy() controllerapi.Controller {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &LIRC{
		State: &lircInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *LIRC) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *LIRC) UnmarshalJSON(data []byte) error {
	d.State = &lircInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
