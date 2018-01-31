package programs

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	pb "github.com/Arvinderpal/matra/common/matrapb"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/programs/bpf/networking/l1"
)

// BPFSampleProgramConf implements programapi.ProgramConf interface
type BPFSampleProgramConf struct {
	//////////////////////////////////////////////////////
	// All programs should define the following fields. //
	//////////////////////////////////////////////////////
	PipelineID  string `json:"pipeline-id"`
	HookType    string `json:"hook-type"`
	ProgramType string `json:"program-type"`
	ContainerID string `json:"container-id"`
	Name        string `json:"name"` // unique per pipeline

	// Populated from the endpoint object:
	EPNetworkInfo *programapi.EndpointNetworkInfo `json:"-"` // Network info from ep.

	////////////////////////////////////////////
	// The fields below are programs specifc. //
	////////////////////////////////////////////
	LogFilePathname string `json:"log-file-path-name"` // program will write to this file.
	DissectEvent    bool   `json:"dissect-event"`      // dissect event data.

	L1BPFConf l1.L1BPFConf `json:"l1-bpf-conf"` // bpf-L1 program specific fields
}

func (c BPFSampleProgramConf) NewProgram() (programapi.Program, error) {

	c.L1BPFConf.EPNetworkInfo = c.EPNetworkInfo
	logFile, err := os.Create(c.LogFilePathname)
	if err != nil {
		return nil, fmt.Errorf("Error creating log file: %s", err)
	}
	prog := &BPFSampleProgram{
		State: &BPFSampleProgramInternalState{
			Conf: c,
			Data: pb.BPFSampleProgramData{
				PacketCount: 0,
				ByteCount:   0,
			},
			logFile: logFile,
			l1BPF:   l1.NewL1BPF(c.ContainerID, c.PipelineID, c.L1BPFConf),
		},
	}

	return prog, nil
}

func (c BPFSampleProgramConf) ValidateConf() error {
	if c.PipelineID == "" {
		return fmt.Errorf("no pipeline id specified in configuration")
	}
	if c.ContainerID == "" {
		return fmt.Errorf("no container id specified in configuration")
	}
	if c.HookType == "" {
		return fmt.Errorf("no hook type specified")
	}
	if c.Name == "" {
		return fmt.Errorf("no name specified in configuration")
	}
	if c.ProgramType != Program_BPFSample {
		return fmt.Errorf("invalid program type specified. Expected %s, but got %s", Program_BPFSample, c.ProgramType)
	}
	if c.EPNetworkInfo == nil {
		return fmt.Errorf("network info not found in program conf")
	}
	if err := c.L1BPFConf.ValidateConf(); err != nil {
		return err
	}
	return nil
}

func (c BPFSampleProgramConf) GetContainerID() string {
	return c.ContainerID
}

func (c BPFSampleProgramConf) GetProgramType() string {
	return c.ProgramType
}

func (c BPFSampleProgramConf) GetHookType() string {
	return c.HookType
}

func (c BPFSampleProgramConf) GetName() string {
	return c.Name
}

func (c BPFSampleProgramConf) GetPipelineID() string {
	return c.PipelineID
}

// BPFSampleProgram implements the Program interface
type BPFSampleProgram struct {
	mu    sync.RWMutex
	State *BPFSampleProgramInternalState
}

type BPFSampleProgramInternalState struct {
	Conf BPFSampleProgramConf    `json:"conf"`
	Data pb.BPFSampleProgramData `json:"data"`

	inChan   chan programapi.MatraEvent
	outChan  chan programapi.MatraEvent
	doneChan chan programapi.MatraEvent

	logFile *os.File
	l1BPF   *l1.L1BPF
}

func (p *BPFSampleProgram) Start(in, out chan programapi.MatraEvent, done chan programapi.MatraEvent, lastInPipeline bool) error {

	err := p.State.l1BPF.Start()
	if err != nil {
		return err
	}

	p.State.inChan = in
	p.State.outChan = out
	p.State.doneChan = done
	// Main goroutine that reads pkts from the input chan.
	go func() {
		// Receive packets until inChan is closed and
		// the buffer queue of inChan is empty.
		for evt := range p.State.inChan {
			p.processData(&evt)
			// if lastInPipeline {
			// }
			p.State.outChan <- evt
		}
		p.cleanup()
	}()
	return nil
}

func (p *BPFSampleProgram) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.State.inChan)

	err := p.State.l1BPF.Stop()
	if err != nil {
		return err
	}
	return nil
}

func (p *BPFSampleProgram) cleanup() {
	if p.State.logFile != nil {
		p.State.logFile.Close()
	}
}

func (p *BPFSampleProgram) signalDone(pkt programapi.MatraEvent) {
	if p.State.doneChan != nil {
		p.State.doneChan <- pkt
	} else {
		logger.Errorf("Done called on program {%s,%s,%s,%s,%s} but done chan is nil -- this indicates a possible error in the program or pipeline configuration", p.State.Conf.GetName(), p.State.Conf.GetProgramType(), p.State.Conf.GetPipelineID(), p.State.Conf.GetHookType(), p.State.Conf.GetContainerID())
	}
}

func (p *BPFSampleProgram) processData(pkt *programapi.MatraEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.State.logFile != nil {
		_, err := p.State.logFile.Write(pkt.Data)
		if err != nil {
			logger.Errorf("Error writing to log file: %s", err)
		}
	}

}

// GetInChan returns the inChan of the program. Really meant for unit testing.
func (p *BPFSampleProgram) GetInChan() chan programapi.MatraEvent {
	return p.State.inChan
}

// GetOutChan returns the outChan of the program. Really meant for unit testing.
func (p *BPFSampleProgram) GetOutChan() chan programapi.MatraEvent {
	return p.State.outChan
}

func (p *BPFSampleProgram) GetConf() programapi.ProgramConf {
	return p.State.Conf
}

func (p *BPFSampleProgram) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("%#v", p)
	// return fmt.Sprintf("Conf{%+v}, Data{%+v}", p.State.Conf, p.State.Data)
}

func (p *BPFSampleProgram) Copy() programapi.Program {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// FIXME: these are shallow copies. Use json marshal/unmarshal or allocate separate memory.

	cpy := &BPFSampleProgram{
		State: &BPFSampleProgramInternalState{
			Conf: p.State.Conf,
			Data: p.State.Data,
		},
	}
	return cpy
}

func (p *BPFSampleProgram) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(p.State)
}

func (p *BPFSampleProgram) UnmarshalJSON(data []byte) error {
	p.State = &BPFSampleProgramInternalState{}
	err := json.Unmarshal(data, p.State)
	return err
}
