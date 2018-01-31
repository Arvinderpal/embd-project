package programs

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Arvinderpal/matra/common"
	pb "github.com/Arvinderpal/matra/common/matrapb"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// DNSFilterProgramConf implements programapi.ProgramConf interface
type DNSFilterProgramConf struct {
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
	AllowedDomains string `json:"allowed-domains"` // comma seperated list of allowed domain names
}

func (c DNSFilterProgramConf) NewProgram() (programapi.Program, error) {
	p := DNSFilterProgram{
		State: &DNSFilterProgramInternalState{
			Conf:           c,
			allowedDomains: make(map[string]int),
			Data: pb.DNSFilterProgramData{
				BlockedRequestsCount: make(map[string]int64, 10),
			},
		},
	}
	allowedDomains := strings.Split(c.AllowedDomains, ",")
	for _, d := range allowedDomains {
		p.State.allowedDomains[d] = 0
	}
	return &p, nil
}

func (c DNSFilterProgramConf) ValidateConf() error {
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
	if c.ProgramType != Program_DNSFilter {
		return fmt.Errorf("Invalid program type specified. Expected %s, but got %s", Program_DNSFilter, c.ProgramType)
	}
	return nil
}

func (c DNSFilterProgramConf) GetContainerID() string {
	return c.ContainerID
}

func (c DNSFilterProgramConf) GetProgramType() string {
	return c.ProgramType
}

func (c DNSFilterProgramConf) GetHookType() string {
	return c.HookType
}

func (c DNSFilterProgramConf) GetName() string {
	return c.Name
}

func (c DNSFilterProgramConf) GetPipelineID() string {
	return c.PipelineID
}

// DNSFilterProgram implements the Program interface
type DNSFilterProgram struct {
	mu    sync.RWMutex
	State *DNSFilterProgramInternalState
}

type DNSFilterProgramInternalState struct {
	Conf DNSFilterProgramConf    `json:"conf"`
	Data pb.DNSFilterProgramData `json:"data"`

	inChan   chan programapi.MatraEvent
	outChan  chan programapi.MatraEvent
	doneChan chan programapi.MatraEvent

	allowedDomains map[string]int // domains allowed
}

func (p *DNSFilterProgram) Start(in, out chan programapi.MatraEvent, done chan programapi.MatraEvent, lastInPipeline bool) error {

	p.State.inChan = in
	p.State.outChan = out
	p.State.doneChan = done
	// Main goroutine that reads pkts from the input chan.
	go func() {
		// Receive packets until inChan is closed and
		// the buffer queue of inChan is empty.
		for pktData := range p.State.inChan {
			v := p.processData(&pktData)
			// if lastInPipeline {
			if pktData.VerdictChannel != nil {
				pktData.VerdictChannel <- v
			}
			// }
			p.State.outChan <- pktData
		}
		p.cleanup()
	}()
	return nil
}

func (p *DNSFilterProgram) cleanup() {
}

func (p *DNSFilterProgram) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.State.inChan)
	return nil
}

func (p *DNSFilterProgram) signalDone(pkt programapi.MatraEvent) {
	if p.State.doneChan != nil {
		p.State.doneChan <- pkt
	} else {
		logger.Errorf("Done called on program {%s,%s,%s,%s,%s} but done chan is nil -- this indicates a possible error in the program or pipeline configuration", p.State.Conf.GetName(), p.State.Conf.GetProgramType(), p.State.Conf.GetPipelineID(), p.State.Conf.GetHookType(), p.State.Conf.GetContainerID())
	}
}

func (p *DNSFilterProgram) processData(pkt *programapi.MatraEvent) (v uint) {

	v = common.NF_DROP
	// Layers that we care about decoding
	var (
		ipLayer  layers.IPv4
		udpLayer layers.UDP
		dns      layers.DNS
		payload  gopacket.Payload
	)

	// Array to store which layers were decoded
	decoded := []gopacket.LayerType{}
	// Faster, predefined layer parser that doesn't make copies of the layer slices
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeIPv4,
		&ipLayer,
		&udpLayer,
		&dns,
		&payload)

	err := parser.DecodeLayers(pkt.Data, &decoded)
	if err != nil {
		logger.Errorf("Error decoding packet: %v", err)
		return
	}
	if len(decoded) == 0 {
		logger.Warningf("Packet contained no valid layers")
		return
	}

	// If we have more than 1 query in the request, then we check that all of them are valid, otherwise we drop the entire packet
	for _, q := range dns.Questions {
		domain := string(q.Name)
		logger.Debugf("Question name: %s", domain)
		if _, ok := p.State.allowedDomains[string(domain)]; !ok {
			logger.Debugf("Request %s not allowed, dropping.", domain)
			v = common.NF_DROP
			// Stats gather: we keep track of blocked domains
			p.mu.Lock()
			if _, ok := p.State.Data.BlockedRequestsCount[domain]; ok {
				p.State.Data.BlockedRequestsCount[domain] = p.State.Data.BlockedRequestsCount[domain] + 1
			} else {
				// insert blocked request into map
				p.State.Data.BlockedRequestsCount[domain] = 1
			}
			p.mu.Unlock()
			return
		}
	}
	v = common.NF_ACCEPT
	return
}

// GetInChan returns the inChan of the program. Really meant for unit testing.
func (p *DNSFilterProgram) GetInChan() chan programapi.MatraEvent {
	return p.State.inChan
}

// GetOutChan returns the outChan of the program. Really meant for unit testing.
func (p *DNSFilterProgram) GetOutChan() chan programapi.MatraEvent {
	return p.State.outChan
}

func (p *DNSFilterProgram) GetConf() programapi.ProgramConf {
	return p.State.Conf
}

func (p *DNSFilterProgram) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("%#v", p)
}

func (p *DNSFilterProgram) Copy() programapi.Program {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cpy := &DNSFilterProgram{
		State: &DNSFilterProgramInternalState{
			Conf: p.State.Conf,
			Data: p.State.Data,
		},
	}
	return cpy
}

func (p *DNSFilterProgram) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(p.State)
}

func (p *DNSFilterProgram) UnmarshalJSON(data []byte) error {
	p.State = &DNSFilterProgramInternalState{}
	err := json.Unmarshal(data, p.State)
	return err
}
