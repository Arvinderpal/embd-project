package netfilterqueue

import (
	"fmt"
	"sync"

	"github.com/Arvinderpal/matra/common/networkhookapi"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/pipeline"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("netfilter-queue")
)

// NetfilterQueueConf implements NetworkHookConf interface
type NetfilterQueueConf struct {
	Type        string      `json:"type"`         // Type.
	ContainerID string      `json:"container-id"` // Container ID.
	NFQConf     NFQueueConf `json:"nfq-conf"`     // internal netfilter queue conf.

	// Populated from the endpoint object:
	EPNetworkInfo programapi.EndpointNetworkInfo // Network info from ep object.
}

func (c NetfilterQueueConf) NewNetworkHook() (networkhookapi.NetworkHook, error) {
	// Set up a kill channel that's shared by the whole pipeline,
	// and close that channel when this hook is detached, as a signal
	// for all the goroutines we started to exit.

	c.NFQConf.Iface = c.EPNetworkInfo.IfName // Interface comes from the endpoint.

	kC := make(chan struct{})
	h := NetfilterQueue{
		State: &NetfilterQueueInternalState{
			Conf:     c,
			killChan: kC,
		},
	}
	err := h.run()
	if err != nil {
		return nil, err
	}
	h.State.isAttached = true
	return &h, nil
}

func (c NetfilterQueueConf) GetType() string {
	return c.Type
}

func (c NetfilterQueueConf) GetContainerID() string {
	return c.ContainerID
}

func (c NetfilterQueueConf) ValidateConf() error {
	if c.ContainerID == "" {
		return fmt.Errorf("no container id specified in configuration")
	}
	if c.Type == "" {
		return fmt.Errorf("no hook type specified in configuration")
	}
	if err := c.NFQConf.ValidateConf(); err != nil {
		return err
	}
	return nil
}

// NetfilterQueue implements the NetworkHook interface
type NetfilterQueue struct {
	sync.RWMutex
	State *NetfilterQueueInternalState `json:"state"`
}

type NetfilterQueueInternalState struct {
	Conf       NetfilterQueueConf   `json:"conf"`
	Pipelines  []*pipeline.Pipeline `json:"pipelines"` // Pipelines on this hook
	nfqueue    *NFQueue
	killChan   chan struct{}
	isAttached bool
}

func (h *NetfilterQueue) String() string {
	h.RLock()
	defer h.RUnlock()
	return fmt.Sprintf("%#v", h)
}

func (h *NetfilterQueue) StartPipeline(confs []programapi.ProgramConf) (string, error) {
	h.Lock()
	defer h.Unlock()
	return h.startPipeline(confs)
}

func (h *NetfilterQueue) startPipeline(confs []programapi.ProgramConf) (string, error) {
	if !h.State.isAttached {
		return "", fmt.Errorf("Hook %s on container %s isn't attached, but start pipeline called", h.State.Conf.GetType(), h.State.Conf.GetContainerID())
	}
	if len(h.State.Pipelines) == 1 {
		return "", fmt.Errorf("Hook %s only supports a single pipeline. Consider removing the existing pipeline %s first.", h.State.Conf.GetType(), h.State.Pipelines[0].ID)
	}

	pipelineID := confs[0].GetPipelineID()
	pipe := h.lookupPipeline(pipelineID)
	if pipe != nil {
		return "", fmt.Errorf("Pipeline %s already exists", pipe.ID)
	}
	pipe, err := pipeline.NewPipeline(pipelineID, confs)
	if err != nil {
		return pipelineID, err
	}
	h.State.Pipelines = append(h.State.Pipelines, pipe)

	err = pipe.Start()
	if err != nil {
		// TODO: do we leave it up to the user to clean up here?
		return pipelineID, err
	}
	return pipelineID, nil
}

func (h *NetfilterQueue) StopPipeline(pipelineID string) error {
	h.Lock()
	defer h.Unlock()
	return h.stopPipeline(pipelineID)
}

func (h *NetfilterQueue) stopPipeline(pipelineID string) error {
	for i, pipe := range h.State.Pipelines {
		if pipe.ID == pipelineID {
			err := pipe.Stop()
			if err != nil {
				// TODO: do we leave it up to the user to clean up here?
				return err
			}
			h.State.Pipelines = append(h.State.Pipelines[:i], h.State.Pipelines[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Pipelines %s not found on hook %s in container %s", pipelineID, h.State.Conf.GetType(), h.State.Conf.ContainerID)
}

// LookupPipeline returns matching pipeline if any
func (h *NetfilterQueue) LookupPipeline(pipelineID string) *pipeline.Pipeline {
	h.RLock()
	defer h.RUnlock()
	return h.lookupPipeline(pipelineID)
}

// lookupPipeline returns matching hooktype
func (h *NetfilterQueue) lookupPipeline(pipelineID string) *pipeline.Pipeline {
	for _, p := range h.State.Pipelines {
		if p.ID == pipelineID {
			return p
		}
	}
	return nil
}

// Run is the main execution routine. It sets up NFQueue and
// sends recieved data to the pipelines for processing.
func (h *NetfilterQueue) run() error {
	var err error

	err = h.addIptablesRules()
	if err != nil {
		logger.Errorf("Failed to setup iptables rules: %s", err)
		return err
	}

	h.State.nfqueue, err = NewNFQueue(h.State.Conf.NFQConf)
	if err != nil {
		logger.Errorf("Failed to open the netfilter-queue: %s", err)
		return err
	}
	err = h.Listen()
	if err != nil {
		logger.Errorf("Listening stopped with an error: %s", err)
		return err
	}
	return nil
}

func (h *NetfilterQueue) DetachNetworkHook() error {
	h.Lock() // Acquiring this lock means, pipeline is inactive.
	defer h.Unlock()
	h.State.isAttached = false
	close(h.State.killChan)
	// h.State.nfqueue.Close() // this should be done in main Listen() go routine
	err := h.removeIptablesRules()
	if err != nil {
		logger.Errorf("Failed to setup iptables rules: %s", err)
		return err
	}

	// We need to stop all pipelines.
	for _, p := range h.State.Pipelines {
		err := h.stopPipeline(p.ID)
		if err != nil {
			logger.Errorf("Error while stopping pipeline %s on hook %s (continuing): %s", p.ID, h.State.Conf.GetType(), err)
			continue
		}
	}
	return nil
}

func (h *NetfilterQueue) GetConf() networkhookapi.NetworkHookConf {
	return h.State.Conf
}

func (h *NetfilterQueue) GetPipelines() []*pipeline.Pipeline {
	h.RLock()
	defer h.RUnlock()
	return h.State.Pipelines
}

func (h *NetfilterQueue) Copy() networkhookapi.NetworkHook {
	h.RLock()
	defer h.RUnlock()
	hCpy := &NetfilterQueue{
		State: &NetfilterQueueInternalState{
			Conf: h.State.Conf,
		},
	}
	for _, pipe := range h.State.Pipelines {
		pipeCpy := pipe.Copy()
		hCpy.State.Pipelines = append(hCpy.State.Pipelines, &pipeCpy)
	}
	return hCpy

}
