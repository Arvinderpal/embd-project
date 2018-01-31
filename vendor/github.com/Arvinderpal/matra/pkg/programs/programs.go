package programs

import (
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/common/types"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("programs")
)

// List of all programs
const (
	Program_UnitTest    = "unittest_program"
	Program_Sample      = "sample_program"
	Program_Transporter = "transporter_program"
	Program_HTTPFilter  = "httpfilter_program"
	Program_DNSFilter   = "dnsfilter_program"
	// BPFs
	Program_BPFSample = "bpf_sample_program"
)

// NewConf is a util method used by pipeline.unmarshal() to get
// a program conf of a particular type. The idea is to localize
// program specific code to this (programs) package.
func NewProgramConf(progType, hookType, pipeID, containerID string, epNInfo *programapi.EndpointNetworkInfo) (programapi.ProgramConf, error) {
	switch progType {
	case Program_UnitTest:
		return &UnitTestProgramConf{
			PipelineID:  pipeID,
			HookType:    hookType,
			ContainerID: containerID,
			ProgramType: progType,
		}, nil
	case Program_Sample:
		return &SampleProgramConf{
			PipelineID:  pipeID,
			HookType:    hookType,
			ContainerID: containerID,
			ProgramType: progType,
		}, nil
	case Program_Transporter:
		return &TransporterProgramConf{
			PipelineID:  pipeID,
			HookType:    hookType,
			ContainerID: containerID,
			ProgramType: progType,
		}, nil
	case Program_HTTPFilter:
		return &HTTPFilterProgramConf{
			PipelineID:  pipeID,
			HookType:    hookType,
			ContainerID: containerID,
			ProgramType: progType,
		}, nil
	case Program_DNSFilter:
		return &DNSFilterProgramConf{
			PipelineID:  pipeID,
			HookType:    hookType,
			ContainerID: containerID,
			ProgramType: progType,
		}, nil
	case Program_BPFSample:
		return &BPFSampleProgramConf{
			PipelineID:    pipeID,
			HookType:      hookType,
			ContainerID:   containerID,
			ProgramType:   progType,
			EPNetworkInfo: epNInfo,
		}, nil
	default:
		return nil, types.ErrUnknownProgramType

	}
}

// NewProgram is a util method used by pipeline package (among others).
// It returns an emtpty Program object of the desired type. The idea
// is to localize program specific code to this (programs) package.
func NewProgram(t string) (programapi.Program, error) {
	switch t {
	case Program_UnitTest:
		return &UnitTestProgram{}, nil
	case Program_Sample:
		return &SampleProgram{}, nil
	case Program_Transporter:
		return &TransporterProgram{}, nil
	case Program_HTTPFilter:
		return &HTTPFilterProgram{}, nil
	case Program_DNSFilter:
		return &DNSFilterProgram{}, nil
	case Program_BPFSample:
		return &BPFSampleProgram{}, nil
	default:
		return nil, types.ErrUnknownProgramType

	}
}
