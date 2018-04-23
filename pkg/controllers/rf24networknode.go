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
	ID             string   `json:"id"`
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

	backend  rf24networknodebackend.RF24NetworkNodeBackend
	radio    RF24.RF24
	network  RF24Network.RF24Network
	streamQs []*message.Queue // per stream queue for outgoing messages
	sQCount  int              // used to identify each stream queue
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
		d.State.backend = rf24networknodebackend.NewRF24NetworkNodeMaster(d.State.network, d.State.Conf.PollInterval)
	} else {
		logger.Infof("creating child node for rf24 network...\n")
		d.State.backend = rf24networknodebackend.NewRF24NetworkNodeChild(d.State.network, d.State.Conf.PollInterval, d.State.Conf.HeartbeatInterval)
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
	return nil
}

// work: processes messages to be sent out and listens for incoming messages on RF.
func (d *RF24NetworkNode) work() {

	// Get internal messages from rcvQ and send them to all subscribers
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
				d.State.backend.Send(msg)
				d.State.rcvQ.Done(msg)
			}
		}
	}()

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

// func (d *RF24NetworkNode) Messenger(stream seguepb.Messenger_MessengerServer) error {
// 	// Add a new queue on which we will receive messages from the work() func; these messages will be then be sent on the stream.
// 	d.mu.Lock()
// 	id := fmt.Sprintf("%d", d.State.sQCount)
// 	d.State.sQCount += 1
// 	sq := message.NewQueue(id)
// 	d.State.streamQs = append(d.State.streamQs, sq)
// 	d.mu.Unlock()
// 	defer func() {
// 		logger.Debugf("grpc: stoping messenger")
// 		sq.ShutDown()
// 		d.mu.Lock()
// 		for i, q := range d.State.streamQs {
// 			if q.ID() == id {
// 				d.State.streamQs = append(d.State.streamQs[:i], d.State.streamQs[i+1:]...)
// 				break
// 			}
// 		}
// 		d.mu.Unlock()
// 	}()
// 	// Send routine
// 	go func() {
// 		defer func() {
// 			logger.Debugf("grpc: stoping messenger-send routine")
// 		}()

// 		logger.Debugf("grpc: starting messenger-send routine")
// 		// Read all the messages in the streams send queue every X
// 		// milliseconds. All the messages will be put in the envelope.
// 		// The idea being that we'll minimize network stack overhead.
// 		// However, this approach introduces latency; alternatives include
// 		// reducing the interval or just sending a message as soon as it arrives.
// 		tickChan := time.NewTicker(time.Millisecond * 100).C
// 		for {
// 			select {
// 			case <-d.State.killChan:
// 				return
// 			case <-tickChan:
// 				if sq.IsShuttingDown() {
// 					return
// 				}
// 				var eMsgs []*seguepb.Message
// 				// NOTE: this assumes only a single consumer of the queue.
// 				sqLen := sq.Len() // TODO: may want to cap this to say 25 msgs
// 				if sqLen == 0 {
// 					continue
// 				}
// 				logger.Debugf("grpc: sending %d messages", sqLen)
// 				for i := 0; i < sqLen; i++ {
// 					iMsg, shutdown := sq.Get()
// 					if shutdown {
// 						logger.Debugf("stopping sender routine for grpc-stream-send-queue-%s", id)
// 						return
// 					}
// 					eMsg, err := message.ConvertToExternalFormat(iMsg)
// 					if err != nil {
// 						logger.Errorf(fmt.Sprintf("error marshaling data in routine for grpc-stream-send-queue-%s: %s", id, err))
// 					} else {
// 						eMsgs = append(eMsgs, eMsg)
// 					}
// 					sq.Done(iMsg)
// 				}
// 				msgEnv := &seguepb.MessageEnvelope{eMsgs}
// 				if err := stream.Send(msgEnv); err != nil {
// 					logger.Errorf("error while sending on grpc-stream-send-queue-%s (exiting send routine): %s", id, err)
// 					return
// 				}
// 			}
// 		}
// 	}()
// 	logger.Debugf("grpc: starting messenger-receive routine")
// 	for {
// 		select {
// 		case <-d.State.killChan:
// 			return nil
// 		default:
// 			msgPBEnv, err := stream.Recv()
// 			if err == io.EOF {
// 				logger.Debugf("grpc: exiting Messanger() EOF recieved")
// 				return nil
// 			}
// 			if err != nil {
// 				logger.Errorf("grpc: error: %s", err)
// 				return err
// 			}
// 			for _, msgPB := range msgPBEnv.Messages {
// 				// We go through the messages in the envelope, converting them into
// 				// the internal message format and sending them to the internal
// 				// routing system.
// 				msg, err := message.ConvertToInternalFormat(msgPB)
// 				if err != nil {
// 					logger.Errorf("grpc: messge convertion error: %s", err)
// 				} else {
// 					logger.Debugf("grpc: received msg %v", msg)
// 					d.State.sndQ.Add(msg)
// 				}
// 			}
// 		}
// 	}
// }
