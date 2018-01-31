package packetmmap

import (
	"fmt"
	"sync"

	"github.com/Arvinderpal/matra/common/networkhookapi"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/pipeline"
	"github.com/google/gopacket/afpacket"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("packet-mmap")
)

// PacketMMAPConf implements NetworkHookConf interface
type PacketMMAPConf struct {
	Type         string              `json:"type"`          // Type.
	ContainerID  string              `json:"container-id"`  // Container ID.
	AFpacketConf AfpacketSnifferConf `json:"afpacket-conf"` // af_packet conf.

	// Populated from the endpoint object:
	EPNetworkInfo programapi.EndpointNetworkInfo // Network info from ep object.
}

func (c PacketMMAPConf) NewNetworkHook() (networkhookapi.NetworkHook, error) {
	// Set up a kill channel that's shared by the whole pipeline,
	// and close that channel when this hook is detached, as a signal
	// for all the goroutines we started to exit.

	c.AFpacketConf.Iface = afpacket.OptInterface(c.EPNetworkInfo.IfName) // Interface comes from the endpoint.

	kC := make(chan struct{})
	h := PacketMMAP{
		State: &PacketMMAPInternalState{
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

func (c PacketMMAPConf) GetType() string {
	return c.Type
}

func (c PacketMMAPConf) GetContainerID() string {
	return c.ContainerID
}

func (c PacketMMAPConf) ValidateConf() error {
	if c.ContainerID == "" {
		return fmt.Errorf("no container id specified in configuration")
	}
	if c.Type == "" {
		return fmt.Errorf("no hook type specified in configuration")
	}
	if err := c.AFpacketConf.ValidateConf(); err != nil {
		return err
	}
	return nil
}

// PacketMMAP implements the NetworkHook interface
type PacketMMAP struct {
	mu    sync.RWMutex
	State *PacketMMAPInternalState `json:"state"`
}

type PacketMMAPInternalState struct {
	Conf       PacketMMAPConf       `json:"conf"`
	Pipelines  []*pipeline.Pipeline `json:"pipelines"` // Pipelines on this hook
	sniffer    *AfpacketSniffer
	killChan   chan struct{}
	isAttached bool
}

func (h *PacketMMAP) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return fmt.Sprintf("%#v", h)
}

func (h *PacketMMAP) StartPipeline(confs []programapi.ProgramConf) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.startPipeline(confs)
}

func (h *PacketMMAP) startPipeline(confs []programapi.ProgramConf) (string, error) {
	if !h.State.isAttached {
		return "", fmt.Errorf("Hook %s on container %s isn't attached, but start pipeline called", h.State.Conf.GetType(), h.State.Conf.GetContainerID())
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

func (h *PacketMMAP) StopPipeline(pipelineID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stopPipeline(pipelineID)
}

func (h *PacketMMAP) stopPipeline(pipelineID string) error {
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
func (h *PacketMMAP) LookupPipeline(pipelineID string) *pipeline.Pipeline {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lookupPipeline(pipelineID)
}

// lookupPipeline returns matching hooktype
func (h *PacketMMAP) lookupPipeline(pipelineID string) *pipeline.Pipeline {
	for _, p := range h.State.Pipelines {
		if p.ID == pipelineID {
			return p
		}
	}
	return nil
}

// Run is the main execution routine. It sets up AF_PACKET and
// send recieved data to the pipelines for processing.
func (h *PacketMMAP) run() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	var err error
	h.State.sniffer, err = NewAfpacketSniffer(h.State.Conf.AFpacketConf)
	if err != nil {
		logger.Errorf("Failed to open the sniffer: %s", err)
		return err
	}
	err = h.Listen()
	if err != nil {
		logger.Errorf("Listening stopped with an error: %s", err)
		return err
	}
	return nil
}

func (h *PacketMMAP) DetachNetworkHook() error {
	h.mu.Lock() // Acquiring this lock means, pipeline is inactive.
	defer h.mu.Unlock()
	h.State.isAttached = false
	close(h.State.killChan)
	// h.State.sniffer.Close() // this should be done in main Listen() go routine

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

func (h *PacketMMAP) GetConf() networkhookapi.NetworkHookConf {
	return h.State.Conf
}

func (h *PacketMMAP) GetPipelines() []*pipeline.Pipeline {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.State.Pipelines
}

// Copy will copy certain internal state.
// IMPORTANT: the returned object is only for informational (e.g. GET)
// operations and should NOT be used to update daemon state!
func (h *PacketMMAP) Copy() networkhookapi.NetworkHook {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// FIXME: these are shallow copies. Use json marshal/unmarshal or allocate separate memory.

	hCpy := &PacketMMAP{
		State: &PacketMMAPInternalState{
			Conf: h.State.Conf,
		},
	}
	for _, pipe := range h.State.Pipelines {
		pipeCpy := pipe.Copy()
		hCpy.State.Pipelines = append(hCpy.State.Pipelines, &pipeCpy)
	}
	return hCpy
}
