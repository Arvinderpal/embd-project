package programs

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/Arvinderpal/matra/common"
	pb "github.com/Arvinderpal/matra/common/matrapb"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/gogo/protobuf/proto"
)

// The program will transport data via UDP/TCP) to a remote server.

// TransporterProgramConf implements programapi.ProgramConf interface
type TransporterProgramConf struct {
	//////////////////////////////////////////////////////
	// All programs should define the following fields. //
	//////////////////////////////////////////////////////
	PipelineID  string `json:"pipeline-id"`
	HookType    string `json:"hook-type"`
	ProgramType string `json:"program-type"`
	ContainerID string `json:"container-id"`
	Name        string `json:"name"` // unique per pipeline

	////////////////////////////////////////////
	// The fields below are programs specifc. //
	////////////////////////////////////////////
	Protocol string `json:"protocol"` // either UDP or TCP
	Host     string `json:"host"`     // Host IP
	Port     int    `json:"port"`     // port number
}

func (c TransporterProgramConf) NewProgram() (programapi.Program, error) {
	prog := TransporterProgram{
		State: &TransporterProgramInternalState{
			Conf: c,
		},
	}
	return &prog, nil
}

func (c TransporterProgramConf) ValidateConf() error {
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
	if c.ProgramType != Program_Transporter {
		return fmt.Errorf("Invalid program type specified. Expected %s, but got %s", Program_Transporter, c.ProgramType)
	}
	if strings.ToLower(c.Protocol) != "udp" && strings.ToLower(c.Protocol) == "tcp" {
		return fmt.Errorf("Unknown protocol type specified: %s", c.Protocol)
	}
	if c.Host == "" {
		return fmt.Errorf("no host specified")
	}
	if c.Port < 0 || c.Port >= (2<<16) {
		return fmt.Errorf("port out of range: %q", c.Port)
	}
	return nil
}

func (c TransporterProgramConf) GetContainerID() string {
	return c.ContainerID
}

func (c TransporterProgramConf) GetProgramType() string {
	return c.ProgramType
}

func (c TransporterProgramConf) GetHookType() string {
	return c.HookType
}

func (c TransporterProgramConf) GetName() string {
	return c.Name
}

func (c TransporterProgramConf) GetPipelineID() string {
	return c.PipelineID
}

// TransporterProgram implements the Program interface
type TransporterProgram struct {
	mu    sync.RWMutex
	State *TransporterProgramInternalState
}

type TransporterProgramInternalState struct {
	Conf TransporterProgramConf `json:"conf"`

	inChan   chan programapi.MatraEvent
	outChan  chan programapi.MatraEvent
	doneChan chan programapi.MatraEvent

	conn net.Conn
}

func (p *TransporterProgram) Start(in, out chan programapi.MatraEvent, done chan programapi.MatraEvent, lastInPipeline bool) error {
	var err error

	// connect with server
	p.State.conn, err = net.Dial(strings.ToLower(p.State.Conf.Protocol), fmt.Sprintf("%s:%d", p.State.Conf.Host, p.State.Conf.Port))
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
		for pktData := range p.State.inChan {
			p.processData(&pktData)
			if lastInPipeline {
				// verdict is applicable only for NFQ hook
				// pktData.Verdict = uint(NF_ACCEPT)
				if pktData.VerdictChannel != nil {
					pktData.VerdictChannel <- uint(common.NF_ACCEPT)
				}
			}
			p.State.outChan <- pktData
		}
		p.cleanup()
	}()
	return nil
}

func (p *TransporterProgram) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.State.inChan)
	p.State.conn.Close()
	return nil
}

func (p *TransporterProgram) cleanup() {
}

func (p *TransporterProgram) signalDone(pkt programapi.MatraEvent) {
	if p.State.doneChan != nil {
		p.State.doneChan <- pkt
	} else {
		logger.Errorf("Done called on program {%s,%s,%s,%s,%s} but done chan is nil -- this indicates a possible error in the program or pipeline configuration", p.State.Conf.GetName(), p.State.Conf.GetProgramType(), p.State.Conf.GetPipelineID(), p.State.Conf.GetHookType(), p.State.Conf.GetContainerID())
	}
}

func (p *TransporterProgram) processData(pkt *programapi.MatraEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, m := range pkt.Msgs {
		if isUDP(p.State.Conf.Protocol) {
			p.transmitUDP(m)
		} else {
			// TODO: add TCP logic
		}
	}
}

func (p *TransporterProgram) transmitUDP(msg *pb.Message) {
	// TODO: we should add some buffering; that is pack data into a
	// 1500 byte packet before transmitting.
	out, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalln("Failed to encode message:", err)
	}
	p.State.conn.Write(out)
	return
}

// GetInChan returns the inChan of the program. Really meant for unit testing.
func (p *TransporterProgram) GetInChan() chan programapi.MatraEvent {
	return p.State.inChan
}

// GetOutChan returns the outChan of the program. Really meant for unit testing.
func (p *TransporterProgram) GetOutChan() chan programapi.MatraEvent {
	return p.State.outChan
}

func (p *TransporterProgram) GetConf() programapi.ProgramConf {
	return p.State.Conf
}

func (p *TransporterProgram) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("%#v", p)
}

func isUDP(proto string) bool {
	if strings.ToLower(proto) == "udp" {
		return true
	}
	return false
}

func (p *TransporterProgram) Copy() programapi.Program {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cpy := &TransporterProgram{
		State: &TransporterProgramInternalState{
			Conf: p.State.Conf,
		},
	}
	return cpy
}

func (p *TransporterProgram) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(p.State)
}

func (p *TransporterProgram) UnmarshalJSON(data []byte) error {
	p.State = &TransporterProgramInternalState{}
	err := json.Unmarshal(data, p.State)
	return err
}
