package rf24networknodebackend

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/Arvinderpal/RF24Network"
	"github.com/Arvinderpal/embd-project/common/message"
)

// Implements RF24NetworkNodeBackend interface
type RF24NetworkNodeMaster struct {
	network      RF24Network.RF24Network
	killChan     chan struct{}
	pollInterval time.Duration
}

func NewRF24NetworkNodeMaster(n RF24Network.RF24Network, pollInterval int) *RF24NetworkNodeMaster {

	master := &RF24NetworkNodeMaster{
		network:      n,
		killChan:     make(chan struct{}),
		pollInterval: time.Duration(pollInterval),
	}
	return master
}

func (r *RF24NetworkNodeMaster) Stop() error {
	close(r.killChan)
	return nil
}

func (r *RF24NetworkNodeMaster) Run() error {
	fmt.Printf("RF24NetworkNodeMaster: running... %v", msg)

	// Main listener loop for RF24Network:
	for {
		r.network.Update()
	INNER_LOOP:
		for {
			select {
			case <-r.killChan:
				return nil
			default:
				if r.network.Available() { // Is there anything ready for us?
					var payload int64
					ptr := (uintptr)(unsafe.Pointer(&payload))
					header := RF24Network.NewRF24NetworkHeader()
					r.network.Read(header, uintptr(ptr), uint16(8))
					fmt.Printf("received payload %x\n", payload)
				} else {
					break INNER_LOOP
				}
			}
			time.Sleep(r.pollInterval * time.Millisecond)
		}
	}
	return nil
}

func (r *RF24NetworkNodeMaster) Send(msg message.Message) error {
	fmt.Printf("RF24NetworkNodeMaster: sending message %v", msg)
	return nil
}
