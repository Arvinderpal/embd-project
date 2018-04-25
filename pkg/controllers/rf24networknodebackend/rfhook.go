package rf24networknodebackend

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/Arvinderpal/RF24Network"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/dim13/cobs"
	"github.com/gogo/protobuf/proto"
)

type RF24NetworkHook struct {
	mu      sync.RWMutex
	network RF24Network.RF24Network
	// receive side data structs
	frameChan chan []byte // frames off the wire are written to this chan
	buf       []byte      // data of the wire is accumulated into this buffer
	tbuf      []byte      // temp buffer for wire immediately off the wire

}

func NewRF24NetworkHook(n RF24Network.RF24Network) *RF24NetworkHook {

	hook := &RF24NetworkHook{
		network:   n,
		frameChan: make(chan []byte, frameChanCapacity),
		tbuf:      make([]byte, rf24NetworkReadBufferrSize),
	}
	return hook
}

func (r *RF24NetworkHook) Receive() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.network.Update()
	for {
		if r.network.Available() { // Is there anything ready for us?
			fmt.Printf("RF24NetworkHook: data available...\n")
			r.rf24NetworkReceive()
		} else {
			break
		}
	}
	return nil
}

func (r *RF24NetworkHook) rf24NetworkReceive() {

	ptr := (uintptr)(unsafe.Pointer(&r.tbuf[0]))
	header := RF24Network.NewRF24NetworkHeader()
	tbufLen := r.network.Read(header, uintptr(ptr), uint16(rf24NetworkReadBufferrSize))
	// fmt.Printf("received payload: \n%x\n", r.tbuf[:tbufLen])
	if tbufLen == 0 {
		// hmmm. no data in payload
		return
	}
	prevEnd := len(r.buf)
	r.buf = append(r.buf, r.tbuf[:tbufLen]...) // append new data to buf
	curFrameStart := 0                         // offset in buf of where the current frame starts
	// Extract frames (if any)
	for i := prevEnd; i < len(r.buf); i++ {
		if r.buf[i] == 0x00 {
			// found frame end
			r.frameChan <- r.buf[curFrameStart : i+1]
			curFrameStart = i + 1
		}
	}
	if curFrameStart >= len(r.buf) {
		// frame ends right at the end of buf
		r.buf = nil
	} else {
		// part of an incomplete frame is present in this buf
		// save it for next read().
		r.buf = r.buf[curFrameStart:]
	}
}

func (r *RF24NetworkHook) rf24NetworkSend(iMsg message.Message) error {

	eMsg, err := message.ConvertToExternalFormat(iMsg)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("error converting to external message format: %s", err))
	}

	// Let's marshal entire message into []byte
	rawpayload, err := proto.Marshal(eMsg)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("error marshalling message: %s", err))
	}

	// Encode the outgoing payload in cobs framing.
	payload := cobs.Encode(rawpayload)

	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Printf("writing payload of len %d: \n %x\n", len(payload), payload)

	r.network.Update() // FIXME: how often to call this method?
	ptr := (uintptr)(unsafe.Pointer(&payload[0]))
	header := RF24Network.NewRF24NetworkHeader(uint16(Master_Node_Address))
	ok := r.network.Write(header, uintptr(ptr), uint16(len(payload)))
	if !ok {
		return fmt.Errorf("write failed.\n")
	}
	return nil
}
