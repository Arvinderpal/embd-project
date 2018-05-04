package controllers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Arvinderpal/RF24"
	"github.com/Arvinderpal/RF24Network"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/pkg/controllers/rf24networknodebackend"
)

// RF24NetworkNodeConf implements ControllerConf interface
type RF24NetworkNodeConf struct {
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
	CEPin             uint16 `json:"ce-pin"`             // CE pin on device
	CSPin             uint16 `json:"cs-pin"`             // CS pin on device
	Channel           byte   `json:"channel"`            // RF Channel to use [0...255]
	Address           uint16 `json:"address"`            // Node address in Octal
	Master            bool   `json:"master"`             // Is this a master node?
	PollInterval      int    `json:"poll-interval"`      // We poll RF module every poll interval [units: Millisecond]
	HeartbeatInterval int    `json:"heartbeat-interval"` // Child nodes send periodic heartbeats
}

func (c RF24NetworkNodeConf) ValidateConf() error {
	if c.ControllerType == "" {
		return fmt.Errorf("no controller type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.ControllerType != Controller_RF24NetworkNode {
		return fmt.Errorf("invalid controller type specified, expected %s, but got %s", Controller_RF24NetworkNode, c.ControllerType)
	}
	if c.Channel < 0 || c.Channel > 255 {
		return fmt.Errorf("invalid channel %d specified in configuration", c.Channel)
	}
	if c.Master {
		if c.Address != 00 {
			return fmt.Errorf("master node must have address 00 (octal), but found %#0", c.Address)
		}
	}
	// These checks are based on http://tmrh20.github.io/RF24NetworkNode/
	// The largest node address is 05555, so 3,125 nodes are allowed on a single channel. I believe this comes from the fact that nRF24L01(+) radio can listen actively on up to 6 radios at once. Thus, the max number of children a node can have is capped from 1...5. That's 5 children and 1 parent.
	if (c.Address&00007) > 00005 || (c.Address&00070) > 00050 || (c.Address&00700) > 00500 || (c.Address&07000) > 05000 {
		return fmt.Errorf("invalid address %#0. address must be less than 05555 (octal)", c.Address)
	}

	return nil
}

func (c RF24NetworkNodeConf) GetType() string {
	return c.ControllerType
}

func (c RF24NetworkNodeConf) GetID() string {
	return c.ID
}

func (c RF24NetworkNodeConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c RF24NetworkNodeConf) NewController(rcvQ *message.Queue, sndQ *message.Queue) (controllerapi.Controller, error) {

	drv := RF24NetworkNode{
		State: &RF24NetworkNodeInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}
	return &drv, nil
}

// RF24NetworkNode implements the Controller interface
type RF24NetworkNode struct {
	mu    sync.RWMutex
	State *RF24NetworkNodeInternal
}

type RF24NetworkNodeInternal struct {
	Conf     RF24NetworkNodeConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}

	backend rf24networknodebackend.RF24NetworkNodeBackend
	radio   RF24.RF24
	network RF24Network.RF24Network
}

// Start: starts the controller logic.
func (d *RF24NetworkNode) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.State.radio = RF24.NewRF24(d.State.Conf.CEPin, d.State.Conf.CSPin)
	d.State.network = RF24Network.NewRF24Network(d.State.radio)

	logger.Infof("starting rf24 radio...\n")

	ok := d.State.radio.Begin() // Setup and configure rf radio
	if !ok {
		return fmt.Errorf("error in radio.Begin() - likely cause is no response from module")
	}

	time.Sleep(5 * time.Millisecond)
	// Dump the confguration of the rf unit for debugging
	d.State.radio.PrintDetails() // TODO: we should send these to logger.

	logger.Infof("starting rf24 network...\n")
	d.State.network.Begin(d.State.Conf.Channel, d.State.Conf.Address)

	if d.State.Conf.Master {
		logger.Infof("creating master node for rf24 network...\n")
		d.State.backend = rf24networknodebackend.NewRF24NetworkNodeMaster(d.State.Conf.ID, d.State.Conf.Address, d.State.Conf.Subscriptions, d.State.network, d.State.Conf.PollInterval, d.State.sndQ, d.State.rcvQ)
	} else {
		logger.Infof("creating child node for rf24 network...\n")
		d.State.backend = rf24networknodebackend.NewRF24NetworkNodeChild(d.State.Conf.ID, d.State.Conf.Address, d.State.Conf.Subscriptions, d.State.network, d.State.Conf.PollInterval, d.State.Conf.HeartbeatInterval, d.State.sndQ, d.State.rcvQ)
	}

	go d.work()
	return nil
}

// Stop: stops the controller logic.
func (d *RF24NetworkNode) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)
	d.State.backend.Stop()

	// We need to cleanup RF24 and RF24Network objects
	RF24Network.DeleteRF24Network(d.State.network)
	d.State.radio.PowerDown() // not sure if ths necessary
	RF24.DeleteRF24(d.State.radio)
	return nil
}

// work: processes messages to be sent out and listens for incoming messages on RF.
func (d *RF24NetworkNode) work() {

	err := d.State.backend.Run()
	if err != nil {
		logger.Errorf("Listening stopped with an error: %s", err)
	}
}

func (d *RF24NetworkNode) GetConf() controllerapi.ControllerConf {
	return d.State.Conf
}

func (d *RF24NetworkNode) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *RF24NetworkNode) Copy() controllerapi.Controller {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &RF24NetworkNode{
		State: &RF24NetworkNodeInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *RF24NetworkNode) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *RF24NetworkNode) UnmarshalJSON(data []byte) error {
	d.State = &RF24NetworkNodeInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
