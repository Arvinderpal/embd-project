// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package bpf

import "fmt"

// Must be synchronized with <bpf/lib/dbg.h>
const (
	DBG_CAPTURE_UNSPEC = iota
	DBG_CAPTURE_FROM_NETDEV
	DBG_CAPTURE_DELIVERY
)

// Must be synchronized with <bpf/lib/dbg.h>
const (
	DBG_UNSPEC = iota
	DBG_GENERIC
	DBG_LOCAL_DELIVERY
	DBG_ERROR_RET
	DBG_TO_HOST
	DBG_TO_STACK
	DBG_PKT_HASH
)

type DebugMsg struct {
	Type    uint8
	SubType uint8
	Source  uint16
	Hash    uint32
	Arg1    uint32
	Arg2    uint32
}

func (n *DebugMsg) Dump(data []byte, prefix string) {
	fmt.Printf("%s MARK %#x FROM %d DEBUG: ", prefix, n.Hash, n.Source)
	switch n.SubType {
	case DBG_GENERIC:
		fmt.Printf("No message, arg1=%d (%#x) arg2=%d (%#x)\n", n.Arg1, n.Arg1, n.Arg2, n.Arg2)
	case DBG_LOCAL_DELIVERY:
		fmt.Printf("Attempting local delivery for container id %d from seclabel %d\n", n.Arg1, n.Arg2)
	case DBG_ERROR_RET:
		fmt.Printf("BPF function %d returned error %d\n", n.Arg1, n.Arg2)
	case DBG_TO_HOST:
		fmt.Printf("Going to host, policy-skip=%d\n", n.Arg1)
	case DBG_TO_STACK:
		fmt.Printf("Going to the stack, policy-skip=%d\n", n.Arg1)
	case DBG_PKT_HASH:
		fmt.Printf("Packet hash=%d (%#x), selected_service=%d\n", n.Arg1, n.Arg1, n.Arg2)
	default:
		fmt.Printf("Unknown message type=%d arg1=%d arg2=%d\n", n.SubType, n.Arg1, n.Arg2)
	}
}

const (
	DebugCaptureLen = 20
)

type DebugCapture struct {
	Type    uint8
	SubType uint8
	Source  uint16
	Hash    uint32
	Len     uint32
	OrigLen uint32
	Arg1    uint32
	// data
}

func (n *DebugCapture) Dump(dissect bool, data []byte, prefix string) {
	fmt.Printf("%s MARK %#x FROM %d DEBUG: %d bytes ", prefix, n.Hash, n.Source, n.Len)
	switch n.SubType {
	case DBG_CAPTURE_FROM_NETDEV:
		fmt.Printf("Incoming packet from netdev ifindex %d\n", n.Arg1)
	case DBG_CAPTURE_DELIVERY:
		fmt.Printf("Delivery to ifindex %d\n", n.Arg1)
	default:
		fmt.Printf("Unknown message type=%d arg1=%d\n", n.SubType, n.Arg1)
	}

	if n.Len > 0 && len(data) > DebugCaptureLen {
		Dissect(dissect, data[DebugCaptureLen:])
	}
}
