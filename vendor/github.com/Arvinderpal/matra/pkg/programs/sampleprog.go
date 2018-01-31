package programs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/Arvinderpal/matra/common"
	pb "github.com/Arvinderpal/matra/common/matrapb"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/gogo/protobuf/proto"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// SampleProgramConf implements programapi.ProgramConf interface
type SampleProgramConf struct {
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
	RunFor          time.Duration `json:"run-for"`            // Program will stop after running after this time expires.
	LogFilePathname string        `json:"log-file-path-name"` // program will write to this file
}

func (c SampleProgramConf) NewProgram() (programapi.Program, error) {
	prog := SampleProgram{
		State: &SampleProgramInternalState{
			Conf: c,
			Data: pb.SampleProgramData{
				PacketCount: 0,
				ByteCount:   0,
			},
		},
	}
	return &prog, nil
}

func (c SampleProgramConf) ValidateConf() error {
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
	if c.ProgramType != Program_Sample {
		return fmt.Errorf("Invalid program type specified. Expected %s, but got %s", Program_Sample, c.ProgramType)
	}
	return nil
}

func (c SampleProgramConf) GetContainerID() string {
	return c.ContainerID
}

func (c SampleProgramConf) GetProgramType() string {
	return c.ProgramType
}

func (c SampleProgramConf) GetHookType() string {
	return c.HookType
}

func (c SampleProgramConf) GetName() string {
	return c.Name
}

func (c SampleProgramConf) GetPipelineID() string {
	return c.PipelineID
}

// SampleProgram implements the Program interface
type SampleProgram struct {
	mu    sync.RWMutex
	State *SampleProgramInternalState
}

type SampleProgramInternalState struct {
	Conf SampleProgramConf    `json:"conf"`
	Data pb.SampleProgramData `json:"data"`

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

func (p *SampleProgram) Start(in, out chan programapi.MatraEvent, done chan programapi.MatraEvent, lastInPipeline bool) error {
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
					prob := rand.Float64()
					if prob < 0.2 {
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

func (p *SampleProgram) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.State.inChan)
	return nil
}

func (p *SampleProgram) cleanup() {
	if p.State.pcapFile != nil {
		p.State.pcapFile.Close()
	}
}

func (p *SampleProgram) signalDone(pkt programapi.MatraEvent) {
	if p.State.doneChan != nil {
		p.State.doneChan <- pkt
	} else {
		logger.Errorf("Done called on program {%s,%s,%s,%s,%s} but done chan is nil -- this indicates a possible error in the program or pipeline configuration", p.State.Conf.GetName(), p.State.Conf.GetProgramType(), p.State.Conf.GetPipelineID(), p.State.Conf.GetHookType(), p.State.Conf.GetContainerID())
	}
}

func (p *SampleProgram) processData(pkt *programapi.MatraEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(pkt.Msgs) == 0 {
		// the 1st program in the pipeline will count packets and bytes
		p.State.Data.PacketCount += 1
		p.State.Data.ByteCount += uint64(len(pkt.Data)) // probably wrong...
		mhData, err := proto.Marshal(&p.State.Data)
		if err != nil {
			logger.Errorf("Error marshaling data in program %q: %s", p.State.Conf.Name, err)
			return
		}
		msg := &pb.Message{
			Type:        pb.MessageType_MsgSample,
			ContainerID: p.State.Conf.GetContainerID(),
			Entries:     []*pb.Entry{{Type: pb.EntryType_EntryNormal, Data: mhData}},
			Context:     []byte(p.State.Conf.Name),
		}
		pkt.Msgs = append(pkt.Msgs, msg)
	} else {
		for _, m := range pkt.Msgs {
			if m.Type == pb.MessageType_MsgSample {
				// assuming the packets and bytes have already been counted by
				// the same program that ran earlier in the pipeline, we increment
				// the Count100Bytes field for every 100bytes received.
				if len(m.Entries) > 0 {
					inData := &pb.SampleProgramData{}
					err := proto.Unmarshal(m.Entries[0].Data, inData)
					if err != nil {
						logger.Errorf("Error Unmarshaling input data to program %q: %s", p.State.Conf.Name, err)
						return
					}
					// NOTE: this is not very accurate :(
					p.State.Data.Count100Bytes += ((inData.ByteCount - p.State.prevByteCount) / 100)
					p.State.prevByteCount = inData.ByteCount
				}
			}
		}
	}

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

	// Example of how to get pkt.Data out of specific layers
	//        for _, layerType := range decoded {
	//            switch layerType {
	//                case layers.LayerTypeIPv4:
	//                    log.Printf("src: %v, dst: %v, proto: %v", ip.State.SrcIP, ip.State.DstIP, ip.State.Protocol)
	//            }
	//        }

}

// GetInChan returns the inChan of the program. Really meant for unit testing.
func (p *SampleProgram) GetInChan() chan programapi.MatraEvent {
	return p.State.inChan
}

// GetOutChan returns the outChan of the program. Really meant for unit testing.
func (p *SampleProgram) GetOutChan() chan programapi.MatraEvent {
	return p.State.outChan
}

func (p *SampleProgram) GetConf() programapi.ProgramConf {
	return p.State.Conf
}

func (p *SampleProgram) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("%#v", p)
	// return fmt.Sprintf("Conf{%+v}, Data{%+v}", p.State.Conf, p.State.Data)
}

func (p *SampleProgram) Copy() programapi.Program {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cpy := &SampleProgram{
		State: &SampleProgramInternalState{
			Conf: p.State.Conf,
			Data: p.State.Data,
		},
	}
	return cpy
}

func (p *SampleProgram) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(p.State)
}

func (p *SampleProgram) UnmarshalJSON(data []byte) error {
	p.State = &SampleProgramInternalState{}
	err := json.Unmarshal(data, p.State)
	return err
}
