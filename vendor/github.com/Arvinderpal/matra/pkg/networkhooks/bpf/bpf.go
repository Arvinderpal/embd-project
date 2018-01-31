package bpf

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/Arvinderpal/matra/common"
	"github.com/Arvinderpal/matra/common/networkhookapi"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/bpf"
	"github.com/Arvinderpal/matra/pkg/pipeline"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("bpf")
)

const eventChanSize = 16

// BPFConf implements NetworkHookConf interface
type BPFConf struct {
	Type        string `json:"type"`         // Type.
	ContainerID string `json:"container-id"` // Container ID.

	NumCpus  int `json:"num-cpus"`  // Number of CPUs (default is runtime.NumCPU()).
	NumPages int `json:"num-pages"` // Number of pages for the event ring buffer (default is 64).

	// Populated from the endpoint object:
	EPNetworkInfo programapi.EndpointNetworkInfo `json:"-"` // Network info from ep object.
}

func (c BPFConf) NewNetworkHook() (networkhookapi.NetworkHook, error) {
	// Set up a kill channel that's shared by the whole pipeline,
	// and close that channel when this hook is detached, as a signal
	// for all the goroutines we started to exit.
	if c.NumCpus == 0 {
		c.NumCpus = runtime.NumCPU()
	}
	if c.NumPages == 0 {
		c.NumPages = 64
	}
	h := &BPF{
		State: &BPFInternalState{
			Conf:            c,
			PipelineHandles: make(map[string]*PipelineHandle),
			killChan:        make(chan struct{}),
		},
	}
	err := h.run()
	if err != nil {
		return nil, err
	}
	h.State.isAttached = true
	return h, nil
}

func (c BPFConf) GetType() string {
	return c.Type
}

func (c BPFConf) GetContainerID() string {
	return c.ContainerID
}

func (c BPFConf) ValidateConf() error {
	if c.ContainerID == "" {
		return fmt.Errorf("no container id specified in configuration")
	}
	if c.Type == "" || c.Type != networkhookapi.NetworkHookType_BPF {
		return fmt.Errorf("incorrect hook type (%s) specified in conf", c.Type)
	}
	return nil
}

// BPF implements the NetworkHook interface
type BPF struct {
	mu    sync.RWMutex
	State *BPFInternalState `json:"state"`
}

type BPFInternalState struct {
	Conf            BPFConf                    `json:"conf"`
	PipelineHandles map[string]*PipelineHandle `json:"pipelines"`

	killChan   chan struct{} // Hook's kill chan
	isAttached bool
}

// PipelineHandle contains per pipeline event config and state.
type PipelineHandle struct {
	mu                sync.RWMutex
	Pipeline          *pipeline.Pipeline   `json:"pipeline"`
	pipeKillChan      chan struct{}        // Pipeline's kill chan.
	perfConfig        *bpf.PerfEventConfig //
	internalEventChan chan []byte          // Internal event channel. Events are written to this chan by the perf event listener and are read by the go routine and forward to pipelines.
}

func (h *BPF) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return fmt.Sprintf("%#v", h)
}

func (h *BPF) StartPipeline(confs []programapi.ProgramConf) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.startPipeline(confs)
}

func (h *BPF) startPipeline(confs []programapi.ProgramConf) (string, error) {
	if !h.State.isAttached {
		return "", fmt.Errorf("hook %s on container %s isn't attached, but start pipeline called", h.State.Conf.GetType(), h.State.Conf.GetContainerID())
	}

	pipelineID := confs[0].GetPipelineID()
	pipe := h.lookupPipeline(pipelineID)
	if pipe != nil {
		return "", fmt.Errorf("pipeline %s already exists", pipe.ID)
	}
	pipe, err := pipeline.NewPipeline(pipelineID, confs)
	if err != nil {
		return pipelineID, err
	}

	err = pipe.Start()
	if err != nil {
		// TODO: do we leave it up to the user to clean up here?
		return pipelineID, err
	}

	// TODO: we should add config parameter that allows us to disable event handling.
	err = h.startPipelineEventMonitor(pipe)
	if err != nil {
		logger.Warningf("unable to start event monitor on pipeline %s (continuing...)", pipelineID)
		// NOTE: This is a non-fatal error. We continue execution.
	}

	return pipelineID, nil
}

func (h *BPF) StopPipeline(pipelineID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stopPipeline(pipelineID)
}

func (h *BPF) stopPipeline(pipelineID string) error {

	pipeH, found := h.State.PipelineHandles[pipelineID]
	if !found {
		return fmt.Errorf("pipeline %s not found on hook %s in container %s", pipelineID, h.State.Conf.GetType(), h.State.Conf.ContainerID)
	}

	pipeH.mu.Lock()
	defer pipeH.mu.Unlock()

	err := h.stopPipelineEventMonitor(pipeH)
	if err != nil {
		return err
	}
	err = pipeH.Pipeline.Stop()
	if err != nil {
		// TODO: do we leave it up to the user to clean up here?
		return err
	}
	delete(h.State.PipelineHandles, pipelineID)
	return nil
}

// LookupPipeline returns matching pipeline if any
func (h *BPF) LookupPipeline(pipelineID string) *pipeline.Pipeline {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lookupPipeline(pipelineID)
}

// lookupPipeline returns matching hooktype
func (h *BPF) lookupPipeline(pipelineID string) *pipeline.Pipeline {
	for _, pHdl := range h.State.PipelineHandles {
		if pHdl.Pipeline.ID == pipelineID {
			return pHdl.Pipeline
		}
	}
	return nil
}

// Run is the main execution routine. It builds and starts the BPF program.
func (h *BPF) run() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.bpfSetup()
	if err != nil {
		return err
	}

	return nil
}

func (h *BPF) bpfSetup() error {
	// Mount BPF Map directory if not already done
	args := []string{"-q", common.BPFMapRoot}
	_, err := exec.Command("mountpoint", args...).CombinedOutput()
	if err != nil {
		args = []string{"bpffs", common.BPFMapRoot, "-t", "bpf"}
		out, err := exec.Command("mount", args...).CombinedOutput()
		if err != nil {
			logger.Errorf("command execution failed: %s\n%s", err, out)
			return err
		}
	}
	return nil
}

// startPipelineEventMonitor starts the event monitor for the pipeline.
func (h *BPF) startPipelineEventMonitor(pipe *pipeline.Pipeline) error {

	pHdl := &PipelineHandle{
		Pipeline:          pipe,
		pipeKillChan:      make(chan struct{}),
		perfConfig:        NewPerfConfig(filepath.Join(common.BPFMatraGlobalMaps, common.ConstructPerfEventMapName(pipe.ContainerID, pipe.ID)), h.State.Conf.NumCpus, h.State.Conf.NumPages),
		internalEventChan: make(chan []byte),
	}
	err := pHdl.eventListener(h.State.killChan)
	if err != nil {
		logger.Errorf("failed to open the perf event listener: %s. config: %+q", err, pHdl.perfConfig)
		return err
	}
	h.State.PipelineHandles[pipe.ID] = pHdl

	err = pHdl.Broadcaster(h.State.killChan)
	if err != nil {
		logger.Errorf("pipeline bpf event broadcaster failed with error: %s", err)
		return err
	}
	return nil
}

// stopPipelineEventMonitor stops the event monitor for the pipeline.
func (h *BPF) stopPipelineEventMonitor(pipeH *PipelineHandle) error {

	close(pipeH.pipeKillChan)
	return nil
}

func (h *BPF) DetachNetworkHook() error {
	h.mu.Lock() // Acquiring this lock means, pipeline is inactive.
	defer h.mu.Unlock()
	h.State.isAttached = false
	close(h.State.killChan)

	// We need to stop all pipelines.
	for _, pHdl := range h.State.PipelineHandles {
		err := h.stopPipeline(pHdl.Pipeline.ID)
		if err != nil {
			logger.Errorf("error while stopping pipeline %s on hook %s (continuing): %s", pHdl.Pipeline.ID, h.State.Conf.GetType(), err)
			continue
		}
	}
	return nil
}

func (h *BPF) GetConf() networkhookapi.NetworkHookConf {
	return h.State.Conf
}

func (h *BPF) GetPipelines() []*pipeline.Pipeline {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.getPipelines()
}

func (h *BPF) getPipelines() []*pipeline.Pipeline {
	var retPipes []*pipeline.Pipeline
	for _, pHdl := range h.State.PipelineHandles {
		cpy := pHdl.Pipeline.Copy()
		retPipes = append(retPipes, &cpy)
	}
	return retPipes
}

// Copy will copy certain internal state.
// IMPORTANT: the returned object is only for informational (e.g. GET)
// operations and should NOT be used to update daemon state!
func (h *BPF) Copy() networkhookapi.NetworkHook {
	h.mu.RLock()
	defer h.mu.RUnlock()
	hCpy := &BPF{
		State: &BPFInternalState{
			Conf:            h.State.Conf,
			PipelineHandles: make(map[string]*PipelineHandle),
		},
	}
	for _, pHdl := range h.State.PipelineHandles {
		pipeCpy := pHdl.Pipeline.Copy()
		perfConfigCpy := *pHdl.perfConfig
		pHdl := &PipelineHandle{
			Pipeline:   &pipeCpy,
			perfConfig: &perfConfigCpy,
		}
		hCpy.State.PipelineHandles[pipeCpy.ID] = pHdl
	}
	return hCpy
}
