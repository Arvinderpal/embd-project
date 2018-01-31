package networkhookapi

import (
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/pipeline"
)

type NetworkHookConf interface {
	NewNetworkHook() (NetworkHook, error)
	GetType() string
	GetContainerID() string
}

func NewNetworkHook(config NetworkHookConf) (NetworkHook, error) {
	return config.NewNetworkHook()
}

type NetworkHook interface {
	StartPipeline([]programapi.ProgramConf) (string, error)
	StopPipeline(string) error
	LookupPipeline(pipelineID string) *pipeline.Pipeline
	GetPipelines() []*pipeline.Pipeline
	GetConf() NetworkHookConf
	DetachNetworkHook() error
	Copy() NetworkHook
	String() string
}

// NetworkHookConfEnvelope is used primarly to unmarshal the config file
// provided by the user. This is necessary because we can have various
// types of hooks. During the unmarshalling phase, the Type field is used
// to determine which actual hook to create.
type NetworkHookConfEnvelope struct {
	Type        string      `json:"type"`
	ContainerID string      `json:"container-id"`
	Conf        interface{} `json:"conf"`
}

// List of Hooks
const (
	NetworkHookType_Mock           = "mock"
	NetworkHookType_PacketMMAP     = "packet_mmap"
	NetworkHookType_NetfilterQueue = "netfilter_queue"
	NetworkHookType_BPF            = "bpf"
)
