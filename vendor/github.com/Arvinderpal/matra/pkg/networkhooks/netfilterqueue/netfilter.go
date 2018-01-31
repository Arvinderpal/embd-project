/*
Go bindings for libnetfilter_queue

This library provides access to packets in the IPTables netfilter queue (NFQUEUE).
The libnetfilter_queue library is part of the http://netfilter.org/projects/libnetfilter_queue/ project.
*/
package netfilterqueue

/*
#cgo pkg-config: libnetfilter_queue
#cgo CFLAGS: -Wall -I/usr/include
#cgo LDFLAGS: -L/usr/lib64/

#include "netfilter.h"
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/Arvinderpal/matra/common"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type NFQueueConf struct {
	// IFace is the specific interface to bind to.
	Iface             string            `json:"iface"`
	QueueId           int               `json:"queue-id"`
	MaxPacketsInQueue int               `json:"max-packets-in-queue"`
	PacketSize        int               `json:"packet-size"`
	PacketFilter      *PacketFilterConf `json:"packet-filter,omitempty"`
}

type PacketFilterConf struct {
	IptablesChain string            `json:"chain"` // default is both INPUT and OUTPUT
	ProtocolPort  *ProtocolPortConf `json:"protocol-port,omitempty"`
	// TODO: make this a slice of ProtocolPortConf
}

type ProtocolPortConf struct {
	Protocol string `json:"protocol"`       // default is all protocols
	Port     *int   `json:"port,omitempty"` // default is all ports
	Source   bool   `json:"source"`         // src or dst (default)
}

func (c NFQueueConf) ValidateConf() error {
	// if c.Iface == "" {
	// 	return fmt.Errorf("no network interface specified in configuration")
	// }
	if c.QueueId < 0 {
		return fmt.Errorf("invalid queue id specified in conf: %v must be >= 0", c.QueueId)
	}
	if c.MaxPacketsInQueue < 0 {
		return fmt.Errorf("invalid max packets in queue specified in conf: %v > 0", c.MaxPacketsInQueue)
	}
	if c.PacketSize < 64 {
		return fmt.Errorf("invalid packet size specified in conf: %v must be >= 64", c.PacketSize)
	}
	if c.PacketFilter != nil {
		if err := c.PacketFilter.ValidateConf(); err != nil {
			return err
		}
	}
	return nil
}

func (c PacketFilterConf) ValidateConf() error {
	// TODO: update the iptables chains.
	chain := strings.ToUpper(c.IptablesChain)
	if chain != "" {
		if chain != "OUTPUT" && chain != "INPUT" {
			//|| chain != "FORWARD"
			return fmt.Errorf("invalid iptables chain specified in conf: %s", c.IptablesChain)
		}
	}
	if c.ProtocolPort != nil {
		proto := strings.ToUpper(c.ProtocolPort.Protocol)
		// TODO: update the protocol list. What about IP and dst/src addresses?
		if proto != "TCP" && proto != "UDP" && proto != "ICMP" {
			return fmt.Errorf("invalid protocol specified in conf: %s", proto)
		}
		if c.ProtocolPort.Port != nil {
			if *c.ProtocolPort.Port < 0 || *c.ProtocolPort.Port >= (1<<16) {
				return fmt.Errorf("invalid port specified in conf: %v must be >= 0 and < %d", *c.ProtocolPort.Port, (1 << 16))
			}
		}
	}
	return nil
}

type NFPacket struct {
	Packet         gopacket.Packet
	VerdictChannel chan uint
}

//Set the verdict for the packet
func (p *NFPacket) SetVerdict(v uint) {
	p.VerdictChannel <- v
}

//Set the verdict for the packet
func (p *NFPacket) SetRequeueVerdict(newQueueId uint16) {
	v := uint(common.NF_QUEUE)
	q := (uint(newQueueId) << 16)
	v = v | q
	p.VerdictChannel <- v
}

type NFQueue struct {
	h       *C.struct_nfq_handle
	qh      *C.struct_nfq_q_handle
	fd      C.int
	packets chan NFPacket
	idx     uint32
}

//Verdict for a packet
// type Verdict C.uint

var theTable = make(map[uint32]*chan NFPacket, 0)
var theTabeLock sync.RWMutex

//Create and bind to queue specified by QueueId
func NewNFQueue(conf NFQueueConf) (*NFQueue, error) {
	var nfq = NFQueue{}
	var err error
	var ret C.int

	if nfq.h, err = C.nfq_open(); err != nil {
		return nil, fmt.Errorf("Error opening NFQueue handle: %v\n", err)
	}

	if ret, err = C.nfq_unbind_pf(nfq.h, common.AF_INET); err != nil || ret < 0 {
		return nil, fmt.Errorf("Error unbinding existing NFQ handler from AF_INET protocol family: %v\n", err)
	}

	if ret, err := C.nfq_bind_pf(nfq.h, common.AF_INET); err != nil || ret < 0 {
		return nil, fmt.Errorf("Error binding to AF_INET protocol family: %v\n", err)
	}

	nfq.packets = make(chan NFPacket)
	nfq.idx = uint32(time.Now().UnixNano())
	theTabeLock.Lock()
	theTable[nfq.idx] = &nfq.packets
	theTabeLock.Unlock()
	if nfq.qh, err = C.CreateQueue(nfq.h, C.u_int16_t(conf.QueueId), C.u_int32_t(nfq.idx)); err != nil || nfq.qh == nil {
		C.nfq_close(nfq.h)
		return nil, fmt.Errorf("Error binding to queue: %v\n", err)
	}

	if ret, err = C.nfq_set_queue_maxlen(nfq.qh, C.u_int32_t(conf.MaxPacketsInQueue)); err != nil || ret < 0 {
		C.nfq_destroy_queue(nfq.qh)
		C.nfq_close(nfq.h)
		return nil, fmt.Errorf("Unable to set max packets in queue: %v\n", err)
	}

	if C.nfq_set_mode(nfq.qh, C.u_int8_t(2), C.uint(conf.PacketSize)) < 0 {
		C.nfq_destroy_queue(nfq.qh)
		C.nfq_close(nfq.h)
		return nil, fmt.Errorf("Unable to set packets copy mode: %v\n", err)
	}

	if nfq.fd, err = C.nfq_fd(nfq.h); err != nil {
		C.nfq_destroy_queue(nfq.qh)
		C.nfq_close(nfq.h)
		return nil, fmt.Errorf("Unable to get queue file-descriptor. %v", err)
	}

	go nfq.run()

	return &nfq, nil
}

//Unbind and close the queue
func (nfq *NFQueue) Close() {
	C.nfq_destroy_queue(nfq.qh)
	C.nfq_close(nfq.h)
	theTabeLock.Lock()
	delete(theTable, nfq.idx)
	theTabeLock.Unlock()
}

//Get the channel for packets
func (nfq *NFQueue) GetPackets() <-chan NFPacket {
	return nfq.packets
}

func (nfq *NFQueue) run() {
	C.Run(nfq.h, nfq.fd)
}

//export go_callback
func go_callback(queueId C.int, data *C.uchar, len C.int, idx uint32) C.uint {
	xdata := C.GoBytes(unsafe.Pointer(data), len)
	packet := gopacket.NewPacket(xdata, layers.LayerTypeIPv4, gopacket.DecodeOptions{Lazy: true, NoCopy: true})
	// TODO(awander): we should also populate CaptureInfo in Packet. See afpacket.go ZeroCopyReadPacketData() as an example.
	p := NFPacket{VerdictChannel: make(chan uint), Packet: packet}
	theTabeLock.RLock()
	cb, ok := theTable[idx]
	theTabeLock.RUnlock()
	if !ok {
		fmt.Fprintf(os.Stderr, "Dropping, unexpectedly due to bad idx=%d\n", idx)
		return C.uint(common.NF_DROP)
	}
	select {
	case (*cb) <- p:
		v := <-p.VerdictChannel
		return C.uint(v)
	default:
		fmt.Fprintf(os.Stderr, "Dropping, unexpectedly due to no recv, idx=%d\n", idx)
		return C.uint(common.NF_DROP)
	}
}
