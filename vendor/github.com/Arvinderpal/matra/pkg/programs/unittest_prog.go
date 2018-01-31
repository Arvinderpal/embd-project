package programs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"

	"github.com/Arvinderpal/matra/common"
	pb "github.com/Arvinderpal/matra/common/matrapb"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// Unit test only
const DummyUnitTestField_Value = 99

// UnitTestProgramConf implements programapi.ProgramConf interface
type UnitTestProgramConf struct {
	//////////////////////////////////////////////////////
	// All programs should define the following fields. //
	//////////////////////////////////////////////////////
	PipelineID  string `json:"pipeline-id"`
	HookType    string `json:"hook-type"`
	ProgramType string `json:"program-type"`
	ContainerID string `json:"container-id"`
	Name        string `json:"name"` // unique per pipeline

	////////////////////////////////////////////
	// The fields below are program specifc. //
	////////////////////////////////////////////
	LogFilePathname string  `json:"log-file-path-name"` // program will write to this file
	DropProbability float64 `json:"drop-probability"`   // applicable to NFQ verdicts
}

func (c UnitTestProgramConf) NewProgram() (programapi.Program, error) {
	prog := UnitTestProgram{
		State: &UnitTestProgramInternalState{
			Conf: c,
			Data: pb.UnitTestProgramData{
				DummyUnitTestField: DummyUnitTestField_Value,
				PacketCount:        0,
				ByteCount:          0,
			},
		},
	}
	return &prog, nil
}

func (c UnitTestProgramConf) ValidateConf() error {
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
	if c.ProgramType != Program_UnitTest {
		return fmt.Errorf("Invalid program type specified. Expected %s, but got %s", Program_UnitTest, c.ProgramType)
	}
	return nil
}

func (c UnitTestProgramConf) GetContainerID() string {
	return c.ContainerID
}

func (c UnitTestProgramConf) GetProgramType() string {
	return c.ProgramType
}

func (c UnitTestProgramConf) GetHookType() string {
	return c.HookType
}

func (c UnitTestProgramConf) GetName() string {
	return c.Name
}

func (c UnitTestProgramConf) GetPipelineID() string {
	return c.PipelineID
}

// UnitTestProgram implements the Program interface
type UnitTestProgram struct {
	mu    sync.RWMutex
	State *UnitTestProgramInternalState
}

// UnitTestProgramInternalState is the internal state of the program.
// Only this struct is snapshoted to disk. See the Marshal and UnMarshal funcs below.
type UnitTestProgramInternalState struct {
	Conf UnitTestProgramConf    `json:"conf"`
	Data pb.UnitTestProgramData `json:"data"`

	inChan  chan programapi.MatraEvent
	outChan chan programapi.MatraEvent

	doneChan chan programapi.MatraEvent

	pcapFile      *os.File
	pcapWriter    *pcapgo.Writer
	prevByteCount uint64

	// Layers that we care about decoding
	eth     layers.Ethernet
	arp     layers.ARP
	ip      layers.IPv4
	ipv6    layers.IPv6
	tcp     layers.TCP
	udp     layers.UDP
	icmp    layers.ICMPv4
	icmpv6  layers.ICMPv6
	dns     layers.DNS
	payload gopacket.Payload
}

func (p *UnitTestProgram) Start(in, out chan programapi.MatraEvent, done chan programapi.MatraEvent, lastInPipeline bool) error {
	// Open pcap log output
	var err error
	p.State.pcapFile, p.State.pcapWriter, err = openPcap(p.State.Conf.LogFilePathname)
	if err != nil {
		return fmt.Errorf("Error opening file: %s", err)
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
					// let's drop some packets for fun
					drop := rand.Float64()
					if drop < p.State.Conf.DropProbability {
						pktData.VerdictChannel <- uint(common.NF_DROP)
					} else {
						pktData.VerdictChannel <- uint(common.NF_ACCEPT)
					}
				}
			}
			p.State.outChan <- pktData
		}
		p.cleanup()
	}()
	return nil
}

func (p *UnitTestProgram) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.State.inChan)
	return nil
}

func (p *UnitTestProgram) cleanup() {
	if p.State.pcapFile != nil {
		p.State.pcapFile.Close()
	}
}

func (p *UnitTestProgram) signalDone(pkt programapi.MatraEvent) {
	if p.State.doneChan != nil {
		p.State.doneChan <- pkt
	} else {
		logger.Errorf("Done called on program {%s,%s,%s,%s,%s} but done chan is nil -- this indicates a possible error in the program or pipeline configuration", p.State.Conf.GetName(), p.State.Conf.GetProgramType(), p.State.Conf.GetPipelineID(), p.State.Conf.GetHookType(), p.State.Conf.GetContainerID())
	}
}

func (p *UnitTestProgram) processData(pkt *programapi.MatraEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.State.Data.PacketCount += 1
	p.State.Data.ByteCount += uint64(len(pkt.Data)) // probably wrong...

	if p.State.pcapWriter != nil {
		if pkt.CI.Length != 0 {
			// TODO: NFQ does not populate the CI for the packet. We need to update go_callback() in netfilter.go to populate this info.
			err := p.State.pcapWriter.WritePacket(pkt.CI, pkt.Data)
			if err != nil {
				logger.Errorf("Error while writing pcap file: %s", err)
			}
		} else {
			logger.Warningf("Unable to write pcap - capture info not present")
		}
	}

	// Array to store which layers were decoded
	decoded := []gopacket.LayerType{}
	// Faster, predefined layer parser that doesn't make copies of the layer slices
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet, // TODO: for nfq, the first layer is LayerTypeIPv4. This will throw an error as is.
		&p.State.eth,
		&p.State.arp,
		&p.State.ip,
		&p.State.ipv6,
		&p.State.tcp,
		&p.State.udp,
		&p.State.icmp,
		&p.State.icmpv6,
		&p.State.dns,
		&p.State.payload)

	err := parser.DecodeLayers(pkt.Data, &decoded)
	if err != nil {
		logger.Errorf("Error decoding packet: %v", err)
		return
	}
	if len(decoded) == 0 {
		logger.Warningf("Packet contained no valid layers")
		return
	}
}

// GetInChan returns the inChan of the program. Really meant for unit testing.
func (p *UnitTestProgram) GetInChan() chan programapi.MatraEvent {
	return p.State.inChan
}

// GetOutChan returns the outChan of the program. Really meant for unit testing.
func (p *UnitTestProgram) GetOutChan() chan programapi.MatraEvent {
	return p.State.outChan
}

func (p *UnitTestProgram) GetConf() programapi.ProgramConf {
	return p.State.Conf
}

func (p *UnitTestProgram) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("%#v", p)
	// return fmt.Sprintf("Conf{%+v}, Data{%+v}", p.State.Conf, p.State.Data)
}

func (p *UnitTestProgram) GetData() pb.UnitTestProgramData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.State.Data
}

func (p *UnitTestProgram) Copy() programapi.Program {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cpy := &UnitTestProgram{
		State: &UnitTestProgramInternalState{
			Conf: p.State.Conf,
			Data: p.State.Data,
		},
	}
	return cpy
}

func (p *UnitTestProgram) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(p.State)
}

func (p *UnitTestProgram) UnmarshalJSON(data []byte) error {
	p.State = &UnitTestProgramInternalState{}
	err := json.Unmarshal(data, p.State)
	return err
}
