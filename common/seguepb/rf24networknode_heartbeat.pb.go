// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rf24networknode_heartbeat.proto

package seguepb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type RF24NetworkNodeHeartbeatData struct {
	Id                string   `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Address           uint32   `protobuf:"varint,2,opt,name=address" json:"address,omitempty"`
	Heartbeatinterval uint64   `protobuf:"varint,3,opt,name=heartbeatinterval" json:"heartbeatinterval,omitempty"`
	Subscriptions     []string `protobuf:"bytes,4,rep,name=subscriptions" json:"subscriptions,omitempty"`
}

func (m *RF24NetworkNodeHeartbeatData) Reset()                    { *m = RF24NetworkNodeHeartbeatData{} }
func (m *RF24NetworkNodeHeartbeatData) String() string            { return proto.CompactTextString(m) }
func (*RF24NetworkNodeHeartbeatData) ProtoMessage()               {}
func (*RF24NetworkNodeHeartbeatData) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *RF24NetworkNodeHeartbeatData) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *RF24NetworkNodeHeartbeatData) GetAddress() uint32 {
	if m != nil {
		return m.Address
	}
	return 0
}

func (m *RF24NetworkNodeHeartbeatData) GetHeartbeatinterval() uint64 {
	if m != nil {
		return m.Heartbeatinterval
	}
	return 0
}

func (m *RF24NetworkNodeHeartbeatData) GetSubscriptions() []string {
	if m != nil {
		return m.Subscriptions
	}
	return nil
}

func init() {
	proto.RegisterType((*RF24NetworkNodeHeartbeatData)(nil), "seguepb.RF24NetworkNodeHeartbeatData")
}

func init() { proto.RegisterFile("rf24networknode_heartbeat.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 179 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2f, 0x4a, 0x33, 0x32,
	0xc9, 0x4b, 0x2d, 0x29, 0xcf, 0x2f, 0xca, 0xce, 0xcb, 0x4f, 0x49, 0x8d, 0xcf, 0x48, 0x4d, 0x2c,
	0x2a, 0x49, 0x4a, 0x4d, 0x2c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2f, 0x4e, 0x4d,
	0x2f, 0x4d, 0x2d, 0x48, 0x52, 0x9a, 0xc3, 0xc8, 0x25, 0x13, 0xe4, 0x66, 0x64, 0xe2, 0x07, 0x51,
	0xec, 0x97, 0x9f, 0x92, 0xea, 0x01, 0x53, 0xeb, 0x92, 0x58, 0x92, 0x28, 0xc4, 0xc7, 0xc5, 0x94,
	0x99, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x19, 0xc4, 0x94, 0x99, 0x22, 0x24, 0xc1, 0xc5, 0x9e,
	0x98, 0x92, 0x52, 0x94, 0x5a, 0x5c, 0x2c, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x1b, 0x04, 0xe3, 0x0a,
	0xe9, 0x70, 0x09, 0xc2, 0xad, 0xc9, 0xcc, 0x2b, 0x49, 0x2d, 0x2a, 0x4b, 0xcc, 0x91, 0x60, 0x56,
	0x60, 0xd4, 0x60, 0x09, 0xc2, 0x94, 0x10, 0x52, 0xe1, 0xe2, 0x2d, 0x2e, 0x4d, 0x2a, 0x4e, 0x2e,
	0xca, 0x2c, 0x28, 0xc9, 0xcc, 0xcf, 0x2b, 0x96, 0x60, 0x51, 0x60, 0xd6, 0xe0, 0x0c, 0x42, 0x15,
	0x4c, 0x62, 0x03, 0x3b, 0xd7, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x55, 0x9e, 0xa3, 0x2a, 0xd1,
	0x00, 0x00, 0x00,
}