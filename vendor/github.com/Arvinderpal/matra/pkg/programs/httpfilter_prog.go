package programs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Arvinderpal/matra/common"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

// How the program works:
// 1. The Assembler  will detect individual TCP streams and for each stream it will create an httpStream object and feed the payload into it. The httpStream object can read the HTTP content from (http.ReadRequest)
// 2. We create a program that:
// 	a. Uses the Assembler to detect individual streams and create httpStreams.
// 	b. When an httpStream is create it starts  streamRunner() in a separate go routine -
// 		i. streamRunner() is actually a func in HTTPFilterProgram.
// 		ii. Streamrunner() will filter based on operation. For example, the program config contains list of ops allowed (e.g. GET/PUT/POST/DELETE)
// 		iii. It will also read the remining body of any request. For example, a request may be split across many packets: after reading the 1st packet with the http header, the runner still needs to read the rest of the body (though the body is of little use to us).
// 3. The PRIMARY problem that comes up is that we have to map streams to individual packets b/c in the end we are dropping packets. One approach to this decision process is as follows:
// 	a. Do everything serially where we wait for the bytes to show up in the appropriate stream after a packet arrives.
// 		i. How do know which stream, the packet ended up on?
// 			1) The "internal" assembler uses a hash of network and transport flow as keys to id the stream. We also maintain a map using these keys. Each stream can be one of the following states:
// 				a) VERDICT_UNKOWN: when the operation in the stream is yet to be determined, the stream is in UNKNOWN state.
// 				b) NF_ACCEPT: once we see that the stream contains a valid operation, we mark it has NF_ACCEPT. The packet containing the http header and subsequent packets of the same request are all allowed through. When all the body bytes are read, we mark state back to UNKNOWN.
// 				c) NF_DROP: same logic as ACCEPT. The only difference is that once, we start dropping, there is no way for the tcp stream to recover.
// 			2) NOTE: some packets like tcp RST, FIN, SYN will not generate any stream data and are ignored by assembler. We can always check the len of the payload to see if Read is appropriate. See top of AssembleWithTimestamp()

// TODO: Do more complete testing. HTTP fitler is very much a rough 1st attempt at L7 filtering.
// TODO: We need to test this program when mulltiple requests are in the same stream; that is GET followed by POST. GET will be allowed but POST req packets must be dropped.

// HTTPFilterProgramConf implements programapi.ProgramConf interface
type HTTPFilterProgramConf struct {
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
	AllowedOperations string `json:"allowed-ops"`        // GET/PUT/POST/DELETE
	LogFilePathname   string `json:"log-file-path-name"` // program will write to this file
}

func (c HTTPFilterProgramConf) NewProgram() (programapi.Program, error) {
	allowedOps := strings.Split(c.AllowedOperations, ",")
	prog := HTTPFilterProgram{
		State: &HTTPFilterProgramInternalState{
			Conf:              c,
			streams:           make(map[key]*httpStream, 5),
			allowedOps:        allowedOps,
			localDoneChan:     make(chan struct{}),
			streamVerdictChan: make(chan uint, 2), // TODO: program blocks with a chan of 1 -- why?
			streamState:       make(map[key]uint, 5),
		},
	}
	return &prog, nil
}

func (c HTTPFilterProgramConf) ValidateConf() error {
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
	if c.ProgramType != Program_HTTPFilter {
		return fmt.Errorf("Invalid program type specified. Expected %s, but got %s", Program_Sample, c.ProgramType)
	}
	return nil
}

func (c HTTPFilterProgramConf) GetContainerID() string {
	return c.ContainerID
}

func (c HTTPFilterProgramConf) GetProgramType() string {
	return c.ProgramType
}

func (c HTTPFilterProgramConf) GetHookType() string {
	return c.HookType
}

func (c HTTPFilterProgramConf) GetName() string {
	return c.Name
}

func (c HTTPFilterProgramConf) GetPipelineID() string {
	return c.PipelineID
}

// key used to identify Streams (same as that used in assembly.go)
type key [2]gopacket.Flow

func (k *key) String() string {
	return fmt.Sprintf("%s:%s", k[0], k[1])
}

// HTTPFilterProgram implements the Program interface
type HTTPFilterProgram struct {
	mu    sync.RWMutex
	State *HTTPFilterProgramInternalState
}

type HTTPFilterProgramInternalState struct {
	Conf HTTPFilterProgramConf `json:"conf"`
	// Data pb.HTTPFilterProgramData `json:"data"`

	inChan   chan programapi.MatraEvent
	outChan  chan programapi.MatraEvent
	doneChan chan programapi.MatraEvent

	allowedOps        []string            // http operations allowed
	streams           map[key]*httpStream // map of streams currently active
	localDoneChan     chan struct{}       // used internaly by this program
	assembler         *tcpassembly.Assembler
	streamVerdictChan chan uint // local chan used by streams to signal to the program the verdict on the received packet
	streamState       map[key]uint
}

func (p *HTTPFilterProgram) Start(in, out chan programapi.MatraEvent, done chan programapi.MatraEvent, lastInPipeline bool) error {

	// Set up assembly
	streamFactory := &httpStreamFactory{p}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	p.State.assembler = tcpassembly.NewAssembler(streamPool)

	// Every minute, flush connections that haven't seen activity in the past 2 minutes.
	ticker := time.Tick(time.Minute)
	go func() {
		for {
			select {
			case <-ticker:
				logger.Infof("Flushing old http streams...")
				p.State.assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
			case <-p.State.localDoneChan:
				return
			}
		}
	}()

	p.State.inChan = in
	p.State.outChan = out
	p.State.doneChan = done
	// Main goroutine that reads pkts from the input chan.
	go func() {
		// Receive packets until inChan is closed and
		// the buffer queue of inChan is empty.
		for pktData := range p.State.inChan {
			p.processData(&pktData)
			// if lastInPipeline {
			// 	if pktData.VerdictChannel != nil {
			// 		pktData.VerdictChannel <- uint(common.NF_ACCEPT)
			// 	}
			// }
			p.State.outChan <- pktData
		}
		p.cleanup()
	}()
	return nil
}

func (p *HTTPFilterProgram) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.State.inChan)
	close(p.State.localDoneChan)
	return nil
}

func (p *HTTPFilterProgram) cleanup() {
}

func (p *HTTPFilterProgram) signalDone(pkt programapi.MatraEvent) {
	if p.State.doneChan != nil {
		p.State.doneChan <- pkt
	} else {
		logger.Errorf("Done called on program {%s,%s,%s,%s,%s} but done chan is nil -- this indicates a possible error in the program or pipeline configuration", p.State.Conf.GetName(), p.State.Conf.GetProgramType(), p.State.Conf.GetPipelineID(), p.State.Conf.GetHookType(), p.State.Conf.GetContainerID())
	}
}

// FIXME: if an error occurs during decode, we will not issue a verdict. We should write on the verdict chan using defer.
func (p *HTTPFilterProgram) processData(pkt *programapi.MatraEvent) {

	// Layers that we care about decoding
	var (
		ipLayer  layers.IPv4
		tcpLayer layers.TCP
		payload  gopacket.Payload
	)

	// Array to store which layers were decoded
	decoded := []gopacket.LayerType{}
	// Faster, predefined layer parser that doesn't make copies of the layer slices
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeIPv4,
		&ipLayer,
		&tcpLayer,
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
	k := key{ipLayer.NetworkFlow(), tcpLayer.TransportFlow()}
	logger.Infof("^^^^^^^^^^^^^^(%s)^^^^^^^^^^^^^^", k)

	p.State.assembler.AssembleWithTimestamp(ipLayer.NetworkFlow(), &tcpLayer, time.Now())

	logger.Infof(" - AssembleWithTimestamp complete...")

	var v uint
	v = common.VERDICT_UNKOWN
	if tcpLayer.SYN || tcpLayer.FIN || tcpLayer.RST || len(tcpLayer.LayerPayload()) == 0 {
		// these packets will not have any associated payload so tcpreader will not get any data below.
		logger.Infof(" - Packet has no payload")
		v = common.NF_ACCEPT
	} else {
		p.mu.RLock()
		v = p.State.streamState[key{ipLayer.NetworkFlow(), tcpLayer.TransportFlow()}]
		p.mu.RUnlock()
		if v == common.VERDICT_UNKOWN {
			logger.Infof(" - Waiting on verdict...")
			v = <-p.State.streamVerdictChan
			logger.Infof(" - Verdict reached: %d", v)
		} else if v == common.NF_DROP {
			// If a drop verdict was written to chan, then we need to read it so it does not block. When a drop verdict is issued on a stream, all subsequent packets are dropped as well.
			var ok bool
			select {
			case v, ok = <-p.State.streamVerdictChan:
				if ok {
					logger.Infof("verdict %d read of chan\n", v)
				}
			default:
				// nop
			}
		}
	}
	if pkt.VerdictChannel != nil {
		logger.Infof(" - Sending verdict %d", v)
		pkt.VerdictChannel <- uint(v)
	}
}

// GetInChan returns the inChan of the program. Really meant for unit testing.
func (p *HTTPFilterProgram) GetInChan() chan programapi.MatraEvent {
	return p.State.inChan
}

// GetOutChan returns the outChan of the program. Really meant for unit testing.
func (p *HTTPFilterProgram) GetOutChan() chan programapi.MatraEvent {
	return p.State.outChan
}

func (p *HTTPFilterProgram) GetConf() programapi.ProgramConf {
	return p.State.Conf
}

func (p *HTTPFilterProgram) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("%#v", p)
}

// AddStream adds a newly created httpStream to the program map
func (p *HTTPFilterProgram) AddStream(s *httpStream) {
	p.mu.Lock()
	defer p.mu.Unlock()
	k := key{s.Net, s.Transport}
	p.State.streams[k] = s
	p.State.streamState[k] = common.VERDICT_UNKOWN
	logger.Infof(" # added stream %s %s", s.Net, s.Transport)
}

// Build a simple HTTP request parser using tcpassembly.StreamFactory and tcpassembly.Stream interfaces

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct {
	prog *HTTPFilterProgram
}

// httpStream will handle the actual decoding of http requests.
type httpStream struct {
	Net, Transport gopacket.Flow
	R              tcpreader.ReaderStream
	Buf            *bufio.Reader
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		Net:       net,
		Transport: transport,
		R:         tcpreader.NewReaderStream(),
	}

	logger.Infof(" # httpStream created...")
	// hstream.Buf = bufio.NewReader(&hstream.R)
	h.prog.AddStream(hstream)

	go h.prog.streamRunner(hstream)
	// go hstream.run() // Important... we must guarantee that data from the reader stream is read.

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.R
}

// TODO: will streamRunner go routine eventually exit if the program is stopeed? If not, then we need to add a check for the done chan in the loop.
func (p *HTTPFilterProgram) streamRunner(s *httpStream) {
	logger.Infof(" # streamRunner...")
	buf := bufio.NewReader(&s.R)
	for {
		req, err := http.ReadRequest(buf)
		if err == io.EOF {
			// p.State.streamVerdictChan <- common.NF_ACCEPT // FIXME: what to do here?
			// We must read until we see an EOF... very important!
			logger.Infof(" # streamRunner: err == EOF hit")
			return
		} else if err != nil {
			// p.State.streamVerdictChan <- common.NF_ACCEPT // FIXME: what to do here?
			logger.Errorf("Error reading http stream %q %q: %s\n", s.Net, s.Transport, err)
		} else {
			logger.Infof(" # streamRunner got a req: %q", req)

			// Now that we have the request, let's see if the operation is allowed
			if checkOperationAllowed(p.State.allowedOps, req.Method) {
				p.State.streamVerdictChan <- common.NF_ACCEPT
				p.mu.Lock()
				p.State.streamState[key{s.Net, s.Transport}] = common.NF_ACCEPT
				p.mu.Unlock()
				logger.Infof(" # streamRunner sent verdict NF_ACCEPT ")
			} else {
				p.State.streamVerdictChan <- common.NF_DROP
				p.mu.Lock()
				p.State.streamState[key{s.Net, s.Transport}] = common.NF_DROP
				p.mu.Unlock()
				logger.Infof(" # streamRunner sent verdict NF_DROP ")
			}

			// DiscardBytesToEOF will block until EOF. We need to let future packets go pass through until the entire request body is read.
			// When the next request comes in, we need again filter.
			bodyBytes := tcpreader.DiscardBytesToEOF(req.Body)
			req.Body.Close()
			logger.Infof(" ****** Received request from stream %q %q: %q with %d bytes in request body\n", s.Net, s.Transport, req, bodyBytes)

			// when the entire body has been read, we can mark the stream back to unknown state. A stream in DROP state never goes back to unknown; essentially that stream will have all subsequent packets dropped.
			if p.State.streamState[key{s.Net, s.Transport}] == common.NF_ACCEPT {
				p.mu.Lock()
				p.State.streamState[key{s.Net, s.Transport}] = common.VERDICT_UNKOWN
				p.mu.Unlock()
			}
		}
	}
}

func checkOperationAllowed(allowedOps []string, op string) bool {
	for _, o := range allowedOps {
		if o == op {
			return true
		}
	}
	return false
}

func (p *HTTPFilterProgram) Copy() programapi.Program {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cpy := &HTTPFilterProgram{
		State: &HTTPFilterProgramInternalState{
			Conf: p.State.Conf,
			// Data: p.State.Data,
		},
	}
	return cpy
}

func (p *HTTPFilterProgram) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(p.State)
}

func (p *HTTPFilterProgram) UnmarshalJSON(data []byte) error {
	p.State = &HTTPFilterProgramInternalState{}
	err := json.Unmarshal(data, p.State)
	return err
}
