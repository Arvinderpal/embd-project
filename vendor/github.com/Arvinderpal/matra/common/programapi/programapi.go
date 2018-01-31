package programapi

import (
	"fmt"
	"net"

	pb "github.com/Arvinderpal/matra/common/matrapb"
	"github.com/Arvinderpal/matra/pkg/mac"
	"github.com/google/gopacket"
	"github.com/vishvananda/netlink"
)

type MatraEvent struct {
	Data           []byte
	CI             gopacket.CaptureInfo
	VerdictChannel chan uint // used by the netfilter_queue hook
	Msgs           []*pb.Message
}

type ProgramConf interface {
	NewProgram() (Program, error)
	ValidateConf() error
	GetContainerID() string
	GetHookType() string
	GetProgramType() string
	GetName() string
	GetPipelineID() string
}

func NewProgram(config ProgramConf) (Program, error) {
	return config.NewProgram()
}

type Program interface {
	Start(in, out chan MatraEvent, done chan MatraEvent, lastInPipeline bool) error
	Stop() error
	GetConf() ProgramConf
	Copy() Program
	String() string
}

// ProgramConfEnvelope is used primarly for easy marshalling/unmarshalling
// of various ProgramConf.
type ProgramConfEnvelope struct {
	Type string
	Conf interface{}
}

// EndpointNetworkInfo contains useful network information about the node and interface with which the ep is associated. This info object is passed down from the ep -> hook -> pipe -> program -> bpf.
type EndpointNetworkInfo struct {
	// IfName must be specified by the user during endpoint create.
	IfName string `json:"iface"` // Container's interface name.

	// These are populated by the EndPoint object:
	HostSideIfIndex int     `json:"host-side-index"` // Host's interface index.
	HostSideMAC     mac.MAC `json:"host-side-mac"`   // Host side veth MAC address.

	// Container internal info:
	ContainerMode bool    `json:"container-mode"` // If interface is inside a container. If false, MAC=HostSideMAC and IPv4 is IfName IP.
	MAC           mac.MAC `json:"mac"`            // Container MAC address.
	IPv4          net.IP  `json:"ipv4"`           // Container IPv4 address.
	// Info of the node on which the ep is running
	NodeMAC mac.MAC `json:"node-mac"` // Node MAC address.
	NodeIP  net.IP  `json:"node-ip"`  // Node IPv4/6 address.
}

func (e EndpointNetworkInfo) Validate() error {
	if e.IfName == "" {
		return fmt.Errorf("no interface specified")
	}
	_, err := netlink.LinkByName(e.IfName)
	if err != nil {
		return fmt.Errorf("Error while fetching Link for interface %s: %s", e.IfName, err)
	}
	// TODO: check remaining fields.
	return nil
}

func (e *EndpointNetworkInfo) DeepCopy() *EndpointNetworkInfo {
	cpy := &EndpointNetworkInfo{
		IfName:          e.IfName,
		HostSideIfIndex: e.HostSideIfIndex,
		HostSideMAC:     make(mac.MAC, len(e.HostSideMAC)),
		MAC:             make(mac.MAC, len(e.MAC)),
		IPv4:            make(net.IP, len(e.IPv4)),
		NodeMAC:         make(mac.MAC, len(e.NodeMAC)),
		NodeIP:          make(net.IP, len(e.NodeIP)),
	}
	copy(cpy.HostSideMAC, e.HostSideMAC)
	copy(cpy.MAC, e.MAC)
	copy(cpy.IPv4, e.IPv4)
	copy(cpy.NodeMAC, e.NodeMAC)
	copy(cpy.NodeIP, e.NodeIP)
	return cpy
}
