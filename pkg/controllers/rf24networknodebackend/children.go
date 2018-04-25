package rf24networknodebackend

import (
	"time"

	"github.com/Arvinderpal/RF24Network"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

// Implements RF24NetworkNodeBackend interface
type RF24NetworkNodeChild struct {
	id            string
	address       uint16
	subscriptions []string
	// network           RF24Network.RF24Network
	killChan          chan struct{}
	heartbeatInterval time.Duration
	pollInterval      time.Duration

	controllerSndQ *message.Queue // messages received on rf24network are placed here for consumption by drivers/controllers on this node.
	controllerRcvQ *message.Queue // messages are read from this queue and sent out over the rf24network

	rfhook *RF24NetworkHook
}

func NewRF24NetworkNodeChild(id string, address uint16, subs []string, n RF24Network.RF24Network, pollInterval, hbinterval int, sndQ, rcvQ *message.Queue) *RF24NetworkNodeChild {

	child := &RF24NetworkNodeChild{
		id:            id,
		address:       address,
		subscriptions: subs,
		// network:           n,
		killChan:          make(chan struct{}),
		pollInterval:      time.Duration(pollInterval),
		heartbeatInterval: time.Duration(hbinterval),
		controllerSndQ:    sndQ,
		controllerRcvQ:    rcvQ,
	}

	child.rfhook = NewRF24NetworkHook(n, child.killChan)

	return child
}

func (r *RF24NetworkNodeChild) Stop() error {
	close(r.killChan)
	return nil
}

func (r *RF24NetworkNodeChild) Run() error {

	go r.rfhook.processFrames()
	go r.messageHandler()
	go r.heartbeat()
	go r.sender()

	tickChan := time.NewTicker(r.pollInterval * time.Millisecond).C
	// Main listener loop for RF24Network:
	for {
		select {
		case <-r.killChan:
			return nil
		case <-tickChan:
			r.rfhook.Receive()
		}
	}
	return nil
}

// TODO: What if the queue gets too long? We should probably drop stale messages that have been sitting too long.
// Queue build up could occur when:
// 1. incoming > outgoing rate.
// 2. remote is unavailable.
func (r *RF24NetworkNodeChild) sender() {
	for {
		select {
		case <-r.killChan:
			return
		default:
			iMsg, shutdown := r.controllerRcvQ.Get()
			if shutdown {
				logger.Debugf("stopping sender\n")
				return
			}

			logger.Debugf("sender: sending message %v\n", iMsg)

			err := r.rfhook.rf24NetworkSend(iMsg)
			if err != nil {
				logger.Errorf("error in rf24 send routine: %v", err)
			}
			r.controllerRcvQ.Done(iMsg)
		}
	}
}

// Send periodic heartbeats to master.
func (r *RF24NetworkNodeChild) heartbeat() {

	var version uint64
	tickChan := time.NewTicker(r.heartbeatInterval * time.Second).C
	hbData := &seguepb.RF24NetworkNodeHeartbeatData{
		Id:                "child-node",
		Address:           uint32(r.address),
		Heartbeatinterval: uint64(r.heartbeatInterval),
		Subscriptions:     r.subscriptions,
	}
	for {
		select {
		case <-r.killChan:
			return
		case <-tickChan:
			logger.Debugf("hb.\n")
			msg := message.Message{
				ID: seguepb.Message_MessageID{
					Type:      seguepb.MessageType_RF24NetworkNodeHeartbeat,
					SubType:   "na",
					Version:   version,
					Qualifier: "",
				},
				Data: hbData,
			}
			version += 1
			err := r.rfhook.rf24NetworkSend(msg)
			if err != nil {
				logger.Errorf("error in rf24 send routine: %v\n", err)
			}
		}
	}
}

// messageHandler: messages received from rf24network
func (r *RF24NetworkNodeChild) messageHandler() {
	for {
		select {
		case <-r.killChan:
			return
		case iMsg := <-r.rfhook.messageChan:
			logger.Debugf("Got Message: %v\n", iMsg)
			r.controllerSndQ.Add(iMsg)
		}
	}
}
