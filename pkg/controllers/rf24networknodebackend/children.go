package rf24networknodebackend

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/Arvinderpal/RF24Network"
	"github.com/Arvinderpal/embd-project/common/message"
)

// Implements RF24NetworkNodeBackend interface
type RF24NetworkNodeChild struct {
	mu                sync.RWMutex
	network           RF24Network.RF24Network
	killChan          chan struct{}
	heartbeatInterval time.Duration
	pollInterval      time.Duration
}

func NewRF24NetworkNodeChild(n RF24Network.RF24Network, pollInterval, hbinterval int) *RF24NetworkNodeChild {

	child := &RF24NetworkNodeChild{
		network:           n,
		killChan:          make(chan struct{}),
		pollInterval:      time.Duration(pollInterval),
		heartbeatInterval: time.Duration(hbinterval),
	}
	return child
}

func (r *RF24NetworkNodeChild) Stop() error {
	close(r.killChan)
	return nil
}

func (r *RF24NetworkNodeChild) Run() error {

	fmt.Printf("RF24NetworkNodeMaster: running... %v", msg)
	go r.heartbeat()

	// Main listener loop for RF24Network:
	for {
		r.mu.Lock()
		r.network.Update()
		r.mu.Unlock()
	INNER_LOOP:
		for {
			select {
			case <-r.killChan:
				return nil
			default:
				r.mu.Lock()
				if r.network.Available() { // Is there anything ready for us?
					var payload int64
					ptr := (uintptr)(unsafe.Pointer(&payload))
					header := RF24Network.NewRF24NetworkHeader()
					r.network.Read(header, uintptr(ptr), uint16(8))
					fmt.Printf("received payload %x\n", payload)
				} else {
					r.mu.Unlock()
					break INNER_LOOP
				}
				r.mu.Unlock()
			}
			time.Sleep(r.pollInterval * time.Millisecond)
		}
	}
	return nil
}

func (r *RF24NetworkNodeChild) Send(msg message.Message) error {
	fmt.Printf("RF24NetworkNodeChild: sending message %v", msg)
	return nil
}

// Send periodic heartbeats to master.
func (r *RF24NetworkNodeChild) heartbeat() {

	master_node := uint16(Master_Node_Address)
	tickChan := time.NewTicker(r.heartbeatInterval * time.Second).C
	for {
		select {
		case <-r.killChan:
			return
		case <-tickChan:
			fmt.Printf(".\n")
			r.mu.Lock()
			r.network.Update()
			var payload int64
			payload = time.Now().UnixNano()
			ptr := (uintptr)(unsafe.Pointer(&payload))
			header := RF24Network.NewRF24NetworkHeader(master_node)
			ok := r.network.Write(header, uintptr(ptr), uint16(8))
			if !ok {
				fmt.Printf("write failed.\n")
			}
			r.mu.Unlock()
		}
	}
}
