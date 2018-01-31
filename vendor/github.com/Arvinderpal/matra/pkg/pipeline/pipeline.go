package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/programs"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("pipeline")
)

type Pipeline struct {
	ID          string                     `json:"id"`           // pipeline id
	Type        string                     `json:"type"`         // hook type for pipe
	ContainerID string                     `json:"container-id"` // Container ID.
	Programs    []*ProgramWrapper          `json:"programs"`     // all programs in the pipeline
	Started     bool                       `json:"started"`      // pipline has been started
	inChan      chan programapi.MatraEvent // input chan to the first Program
	outChan     chan programapi.MatraEvent // output chan from last Program
	killChan    chan struct{}              // used to terminate pipeline

	doneChan          chan programapi.MatraEvent // used to notify hook when pipe is done
	progDoneChan      chan programapi.MatraEvent // prog done chan which is initialized
	ptrToProgDoneChan chan programapi.MatraEvent // points to above done chan
}

func NewPipeline(pipelineID string, confs []programapi.ProgramConf) (*Pipeline, error) {
	pipe := &Pipeline{
		ID: pipelineID,
		// We take the hook and container id from the config of the 1st prog.
		Type:        confs[0].GetHookType(),
		ContainerID: confs[0].GetContainerID(),
	}
	for _, conf := range confs {
		prog, err := programapi.NewProgram(conf)
		if err != nil {
			return nil, err
		}
		pipe.Programs = append(pipe.Programs, &ProgramWrapper{prog})
	}
	return pipe, nil
}

func (pipe *Pipeline) Start() error {
	lastInPipeline := false
	pipe.killChan = make(chan struct{})
	pipe.doneChan = make(chan programapi.MatraEvent)
	pipe.progDoneChan = make(chan programapi.MatraEvent)
	// pipe.ptrToProgDoneChan = pipe.progDoneChan
	in := make(chan programapi.MatraEvent)
	out := make(chan programapi.MatraEvent)
	for i, prog := range pipe.Programs {
		if i == 0 {
			// first Program in pipeline
			pipe.inChan = in
		}
		if i == len(pipe.Programs)-1 {
			// last program
			lastInPipeline = true
			pipe.outChan = out
		}
		err := prog.Start(in, out, pipe.ptrToProgDoneChan, lastInPipeline)
		if err != nil {
			return err
		}
		if !lastInPipeline {
			in = out
			out = make(chan programapi.MatraEvent)
		}
	}

	pipe.Started = true
	pipe.progDoneReceiver()
	pipe.outChanHandler()
	return nil
}

// Copy will copy the pipeline's program state by calling program specific Copy funcs. It's not a clone and should not be used to update internal data structures. It should only be called while holding a RLock() on the hook to which the pipeline is attached.
func (pipe *Pipeline) Copy() Pipeline {
	cpy := Pipeline{
		ID:          pipe.ID,
		Type:        pipe.Type,
		ContainerID: pipe.ContainerID,
		Started:     pipe.Started,
	}
	// cpy.Programs = nil // = make([]*ProgramWrapper, len(pipe.Programs))
	for _, w := range pipe.Programs {
		pCpy := w.Program.Copy()
		cpy.Programs = append(cpy.Programs, &ProgramWrapper{pCpy})
	}
	return cpy
}

func (pipe *Pipeline) Stop() error {
	for _, prog := range pipe.Programs {
		// NOTE: prog.Stop() will call close(inChan). This should be enough
		// to allow force all packet processing go routines in the Programs
		// to exit since they read from their inChan. We can do this b/c Stop
		// pipeline is only called when the pipeline is inactive (i.e. there
		// are no packets in these chans).
		err := prog.Stop()
		if err != nil {
			return err
		}
	}
	pipe.Started = false
	close(pipe.killChan)
	close(pipe.doneChan)
	return nil
}

func (pipe *Pipeline) ProcessData(data programapi.MatraEvent) {
	pipe.ptrToProgDoneChan = pipe.progDoneChan
	pipe.inChan <- data
}

func (pipe *Pipeline) progDoneReceiver() {
	// If we see a done from any of the programs running in the pipeline,
	// we will notifiy the main packet dispatcher (in the hook) that this
	// pipeline is done processing the packet
	go func() {
		for {
			select {
			case result := <-pipe.progDoneChan:
				pipe.ptrToProgDoneChan = nil
				pipe.doneChan <- result
			case <-pipe.killChan:
				return
			}
		}
	}()
}

// outChanHandler will read packets that have finished processing in the
// pipeline and are ready to leave. In most cases, we simply drop them here.
func (pipe *Pipeline) outChanHandler() {
	go func() {
		// Receive packets until outChan is closed and
		// the buffer queue of outChan is empty.
		for {
			select {
			case result := <-pipe.outChan:
				pipe.signalDone(result)
			case <-pipe.killChan:
				return
			}
		}
	}()
}

func (pipe *Pipeline) signalDone(result programapi.MatraEvent) {
	select {
	// if ptrToProgDoneChan is nil, the case statement is ignored.
	case pipe.ptrToProgDoneChan <- result:
	}
}

func (pipe *Pipeline) GetDoneChan() chan programapi.MatraEvent {
	return pipe.doneChan
}

// PipelineConfEnvelope is used primarly for easy marshalling/unmarshalling
// of various PipelineConf.
type PipelineConfEnvelope struct {
	ID        string
	Type      string
	Container string
	Confs     []programapi.ProgramConfEnvelope
}

// -----------------------------------------------------------------------
// NOTE: Originally we wrote a custom marshal func for Program type.
// Later this deemed unnecessary because the unmarshal func was modified
// to extract the "type" from the default marshal format.
// -----------------------------------------------------------------------
// MarshalJSON returns a json object with pipeline
// func (pipe *Pipeline) MarshalJSON() ([]byte, error) {
// 	for _, p := range pipe.Programs {
// 		switch prog := p.(type) {
// 		case *programs.SampleProgram:
// 			pd := programMarshal{
// 				Type:    programs.Program_Sample,
// 				Program: prog,
// 			}
// 			hd.Programs = append(hd.Programs, pd)
// 		default:
// 			return nil, ErrUnknownProgramType
// 		}
// 	}
// 	result, err := json.Marshal(hd)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }
//
// type programMarshal struct {
// 	Type    string      `json:"type,omitempty"`
// 	Program interface{} `json:"program,omitempty"`
// }

// ProgramWrapper is a helper type used handle the program type
// based serialization of its configuration.
type ProgramWrapper struct {
	programapi.Program
}

type programUnmarshal struct {
	// Type    string
	// Program json.RawMessage
	Program interface{}
}

// UnmarshalJSON returns a representation of the JSON object
// as a Program type.
func (wrap *ProgramWrapper) UnmarshalJSON(data []byte) error {

	progType, err := getProgramType(data)
	if err != nil {
		return err
	}

	var rawJson json.RawMessage
	env := programUnmarshal{
		Program: &rawJson,
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	// logger.Infof("rawJson: %+v\n", string(rawJson)) // REMOVE
	prog, err := programs.NewProgram(progType)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rawJson, &prog); err != nil {
		return err
	}
	wrap.Program = prog
	return nil
}

// getProgramType will extract the program type from []byte
func getProgramType(data []byte) (string, error) {
	tmp := programUnmarshal{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return "", err
	}
	hookMap := tmp.Program.(map[string]interface{})
	confMap := hookMap["conf"].(map[string]interface{})
	progType, ok := confMap["program-type"]
	if !ok {
		return "", fmt.Errorf("no program type found")
	}
	return progType.(string), nil
}

func (wrap *ProgramWrapper) String() string {
	return fmt.Sprintf("%#v", wrap.Program)
}
