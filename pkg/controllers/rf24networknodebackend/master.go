package rf24networknodebackend

import (
	"fmt"
	"sync"
	"time"

	"github.com/Arvinderpal/RF24Network"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/dim13/cobs"
	"github.com/gogo/protobuf/proto"
)

// Implements RF24NetworkNodeBackend interface
type RF24NetworkNodeMaster struct {
	mu             sync.RWMutex
	id             string
	address        uint16
	subscriptions  []string
	network        RF24Network.RF24Network
	killChan       chan struct{}
	pollInterval   time.Duration
	controllerSndQ *message.Queue // messages received on rf24network are placed here for consumption by drivers/controllers on this node.
	controllerRcvQ *message.Queue // messages are read from this queue and sent out over the rf24network

	messageChan chan message.Message // received messages are put here
	rfhook      *RF24NetworkHook
}

func NewRF24NetworkNodeMaster(id string, address uint16, subs []string, n RF24Network.RF24Network, pollInterval int, sndQ, rcvQ *message.Queue) *RF24NetworkNodeMaster {

	master := &RF24NetworkNodeMaster{
		id:             id,
		address:        address,
		subscriptions:  subs,
		network:        n,
		killChan:       make(chan struct{}),
		pollInterval:   time.Duration(pollInterval),
		rfhook:         NewRF24NetworkHook(n),
		messageChan:    make(chan message.Message),
		controllerSndQ: sndQ,
		controllerRcvQ: rcvQ,
	}
	return master
}

func (r *RF24NetworkNodeMaster) Stop() error {
	close(r.killChan)
	return nil
}

func (r *RF24NetworkNodeMaster) Run() error {
	fmt.Printf("RF24NetworkNodeMaster: running...\n")

	go r.processFrames()
	go r.messageHandler()
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
func (r *RF24NetworkNodeMaster) sender() {
	for {
		select {
		case <-r.killChan:
			return
		default:
			iMsg, shutdown := r.controllerRcvQ.Get()
			if shutdown {
				fmt.Printf("stopping sender\n")
				return
			}

			fmt.Printf("sender: sending message %v\n", iMsg)

			err := r.rfhook.rf24NetworkSend(iMsg)
			if err != nil {
				fmt.Errorf("error in rf24 send routine: %v", err)
			}
			r.controllerRcvQ.Done(iMsg)
		}
	}
}

func (r *RF24NetworkNodeMaster) processFrames() {
	for {
		select {
		case <-r.killChan:
			return
		case frame := <-r.rfhook.frameChan:
			fmt.Printf("received frame of length %d: %x\n", len(frame), frame)
			// Decode the frame in cobs framing.
			rawData := cobs.Decode(frame)
			eMsg := &seguepb.Message{}
			err := proto.Unmarshal(rawData, eMsg)
			if err != nil {
				fmt.Printf(fmt.Sprintf("error unmarshaling to sequepb.message: %s", err))
			}
			iMsg, err := message.ConvertToInternalFormat(eMsg)
			if err != nil {
				fmt.Printf(fmt.Sprintf("error converting to internal message format: %s", err))
			}
			r.messageChan <- iMsg
		}
	}
}

// messageHandler: messages received on rf24network are p
func (r *RF24NetworkNodeMaster) messageHandler() {
	for {
		select {
		case <-r.killChan:
			return
		case iMsg := <-r.messageChan:
			if iMsg.ID.Type == seguepb.MessageType_RF24NetworkNodeHeartbeat {
				r.heartbeatHandler(iMsg)
			} else {
				fmt.Printf("Got Message: %v\n", iMsg)
				r.controllerSndQ.Add(iMsg)
			}
		}
	}
}

func (r *RF24NetworkNodeMaster) heartbeatHandler(msg message.Message) {

	fmt.Printf("Got hearbeat: %v\n", msg)
}
