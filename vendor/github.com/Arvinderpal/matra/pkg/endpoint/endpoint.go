package endpoint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Arvinderpal/matra/common"
	"github.com/Arvinderpal/matra/common/networkhookapi"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/common/types"
	"github.com/Arvinderpal/matra/pkg/networkhooks/bpf"
	"github.com/Arvinderpal/matra/pkg/networkhooks/netfilterqueue"
	"github.com/Arvinderpal/matra/pkg/networkhooks/packetmmap"
	"github.com/Arvinderpal/matra/pkg/option"

	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("matra-endpoint")
)

const (
	maxLogs = 16
)

// Endpoint contains all the details for a particular container and the host
// interface(s) to which it's connected.
type Endpoint struct {
	mutex       sync.RWMutex
	ContainerID string                          `json:"container-id"` // ContainerID ID.
	NetworkInfo *programapi.EndpointNetworkInfo `json:"network-info"` // Network info.
	Hooks       []*NetworkHookWrapper           `json:"hooks"`        // Network Stack Hooks
	Opts        *option.BoolOptions             `json:"options"`      // Endpoint options.
	Status      *EndpointStatus                 `json:"status,omitempty"`
}

type statusLog struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type EndpointStatus struct {
	Log     []*statusLog `json:"log,omitempty"`
	Index   int          `json:"index"`
	indexMU sync.RWMutex
}

func (e *EndpointStatus) lastIndex() int {
	lastIndex := e.Index - 1
	if lastIndex < 0 {
		return maxLogs - 1
	}
	return lastIndex
}

func (e *EndpointStatus) getAndIncIdx() int {
	idx := e.Index
	e.Index++
	if e.Index >= maxLogs {
		e.Index = 0
	}
	return idx
}

func (e *EndpointStatus) addStatusLog(s *statusLog) {
	idx := e.getAndIncIdx()
	if len(e.Log) < maxLogs {
		e.Log = append(e.Log, s)
	} else {
		e.Log[idx] = s
	}
}

func (e *EndpointStatus) String() string {
	e.indexMU.RLock()
	defer e.indexMU.RUnlock()
	if len(e.Log) > 0 {
		lastLog := e.Log[e.lastIndex()]
		if lastLog != nil {
			return fmt.Sprintf("%s", lastLog.Status.Code)
		}
	}
	return OK.String()
}

func (e *EndpointStatus) DumpLog() string {
	e.indexMU.RLock()
	defer e.indexMU.RUnlock()
	logs := []string{}
	for i := e.lastIndex(); ; i-- {
		if i < 0 {
			i = maxLogs - 1
		}
		if i < len(e.Log) && e.Log[i] != nil {
			logs = append(logs, fmt.Sprintf("%s - %s",
				e.Log[i].Timestamp.Format(time.RFC3339), e.Log[i].Status))
		}
		if i == e.Index {
			break
		}
	}
	if len(logs) == 0 {
		return OK.String()
	}
	return strings.Join(logs, "\n")
}

func (es *EndpointStatus) DeepCopy() *EndpointStatus {
	cpy := &EndpointStatus{}
	es.indexMU.RLock()
	defer es.indexMU.RUnlock()
	cpy.Index = es.Index
	cpy.Log = []*statusLog{}
	for _, v := range es.Log {
		cpy.Log = append(cpy.Log, v)
	}
	return cpy
}

func (e Endpoint) Validate() error {
	if e.NetworkInfo != nil {
		err := e.NetworkInfo.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Endpoint) DeepCopy() *Endpoint {

	cpy := &Endpoint{
		ContainerID: e.ContainerID,
	}
	if e.NetworkInfo != nil {
		cpy.NetworkInfo = e.NetworkInfo.DeepCopy()
	}
	if e.Opts != nil {
		cpy.Opts = e.Opts.DeepCopy()
	}
	if e.Status != nil {
		cpy.Status = e.Status.DeepCopy()
	}
	for _, h := range e.Hooks {
		hCpy := h.NetworkHook.Copy()
		cpy.Hooks = append(cpy.Hooks, &NetworkHookWrapper{hCpy})
	}

	return cpy
}

// GetNetworkInfo will return a copy of NetworkInfo
func (e *Endpoint) GetNetworkInfo() *programapi.EndpointNetworkInfo {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.NetworkInfo.DeepCopy()
}

// Snapshot will write the snapshot json in the ep's directory
func (e *Endpoint) Snapshot() error {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.snapshot()
}

// snapshot will write the snapshot json in the ep's directory
func (e *Endpoint) snapshot() error {

	bytes, err := json.MarshalIndent(e, "", " ")
	if err != nil {
		return err
	}
	tmpSnap := filepath.Join(".", e.ContainerID, common.SnapshotFileName+".tmp")
	err = ioutil.WriteFile(tmpSnap, bytes, os.FileMode(0644))
	// Defer a cleanup routine if there is an error somepoint later.
	defer func() {
		if err != nil {
			newErr := os.Remove(tmpSnap)
			if newErr != nil {
				logger.Warningf("Warning: error while removing %s: %s", common.SnapshotFileName+".tmp", newErr)
			}
		}
	}()
	if err != nil {
		return err
	}
	// Atomically swap temporary file out with the newly written file.
	oldSnap := path.Join(".", e.ContainerID, common.SnapshotFileName)
	err = os.Rename(tmpSnap, oldSnap)
	return err
}

func (e *Endpoint) PrettyPrint() string {
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return fmt.Sprintf(string(b))
}

func OptionChanged(key string, value bool, data interface{}) {
}

func (ep *Endpoint) ApplyOpts(opts option.OptionMap) bool {
	return ep.Opts.Apply(opts, OptionChanged, ep) > 0
}

func (ep *Endpoint) SetDefaultOpts(opts *option.BoolOptions) {
}

func (e *Endpoint) LogStatus(code StatusCode, msg string) {
	e.Status.indexMU.Lock()
	defer e.Status.indexMU.Unlock()
	sts := &statusLog{
		Status: Status{
			Code: code,
			Msg:  msg,
		},
		Timestamp: time.Now(),
	}
	e.Status.addStatusLog(sts)
}

func (e *Endpoint) LogStatusOK(msg string) {
	e.Status.indexMU.Lock()
	defer e.Status.indexMU.Unlock()
	sts := &statusLog{
		Status:    NewStatusOK(msg),
		Timestamp: time.Now(),
	}
	e.Status.addStatusLog(sts)
}

// LookupHook returns matching hooktype
// IMPORTANT: if mutex lock is already held by calling func, use lookupHook
// instead of this LookupHook
func (e *Endpoint) LookupHook(hookType string) networkhookapi.NetworkHook {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.lookupHook(hookType)
}

// lookupHook returns matching hooktype
// IMPORTANT: aquire a mutex lock before calling this func
func (e *Endpoint) lookupHook(hookType string) networkhookapi.NetworkHook {
	for _, h := range e.Hooks {
		if h.GetConf().GetType() == hookType {
			return h
		}
	}
	return nil
}

func (e *Endpoint) StartPipeline(confs []programapi.ProgramConf) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	err := e.startPipeline(confs)
	err2 := e.snapshot() // NOTE: we snapshot even if error occured during start
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (e *Endpoint) startPipeline(confs []programapi.ProgramConf) error {
	if len(confs) <= 0 {
		return fmt.Errorf("No programs specified in pipeline")
	}
	hookType := confs[0].GetHookType()
	hook := e.lookupHook(hookType)
	if hook == nil {
		return fmt.Errorf("hook %s not found on ep %s", hookType, e.ContainerID)
	}
	id, err := hook.StartPipeline(confs)
	if err != nil {
		e.LogStatus(Failure, fmt.Sprintf("Could not start pipeline: %s", err))
		return err
	}
	e.LogStatusOK(fmt.Sprintf("Started pipeline %s on Hook %s for Container %s", id, hook.GetConf().GetType(), e.ContainerID))
	return nil
}

func (e *Endpoint) StopPipeline(hookType, pipelineID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	err := e.stopPipeline(hookType, pipelineID)
	err2 := e.snapshot()
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (e *Endpoint) stopPipeline(hookType, pipelineID string) error {
	hook := e.lookupHook(hookType)
	if hook == nil {
		return fmt.Errorf("hook %s not found on ep %s", hookType, e.ContainerID)
	}
	err := hook.StopPipeline(pipelineID)
	if err != nil {
		e.LogStatus(Failure, fmt.Sprintf("Could not stop pipeline: %s", err))
		return err
	}
	e.LogStatusOK(fmt.Sprintf("Stopped pipeline %s on Hook %s for Container %s", pipelineID, hookType, e.ContainerID))
	return nil
}

func (e *Endpoint) AttachNetworkHook(hook networkhookapi.NetworkHook) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.attachNetworkHook(hook)
	err := e.snapshot()
	return err
}

func (e *Endpoint) attachNetworkHook(hook networkhookapi.NetworkHook) error {
	wrap := &NetworkHookWrapper{hook}
	e.Hooks = append(e.Hooks, wrap)
	e.LogStatusOK(fmt.Sprintf("Attached Network Hook %s", hook.GetConf().GetType()))
	return nil
}

func (e *Endpoint) DetachNetworkHook(hookType string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	err := e.detachNetworkHook(hookType)
	err2 := e.snapshot()
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (e *Endpoint) detachNetworkHook(hookType string) error {
	for i, h := range e.Hooks {
		if h.GetConf().GetType() == hookType {
			err := h.DetachNetworkHook()
			if err != nil {
				e.LogStatus(Failure, fmt.Sprintf("Error while detaching hook %s: %s", hookType, err))
				// TODO: need a better solution here to cleanup after an error occurs.
			}
			e.Hooks = append(e.Hooks[:i], e.Hooks[i+1:]...)
			e.LogStatusOK(fmt.Sprintf("Detached hook %s and stopped all associated programs", hookType))
			return nil
		}
	}
	return fmt.Errorf("Hook %s not found in container %s", hookType, e.ContainerID)
}

// NetworkHookWrapper is a helper type used handle the hook type
// based serialization of its configuration.
type NetworkHookWrapper struct {
	networkhookapi.NetworkHook
}

type networkHookUnmarshal struct {
	NetworkHook interface{}
}

// -----------------------------------------------------------------------
// NOTE: Originally we wrote a custom marshal func for NetworkHook.
// Later this deemed unnecessary because the unmarshal func was modified
// to extract the "type" from the default marshal format.
// -----------------------------------------------------------------------
// type networkHookMarshal struct {
// 	Type     string           `json:"type,omitempty"`
// 	Config   interface{}      `json:"config,omitempty"`
// 	Programs []pipeline.PipelineMarshal `json:"programs,omitempty"`
// }
// // MarshalJSON returns a json object with the network hook.
// func (wrap *NetworkHookWrapper) MarshalJSON() ([]byte, error) {
// 	switch hook := wrap.Hook.(type) {
// 	case *packetmmap.PacketMMAP:
// 		hd := networkHookMarshal{
// 			Type:   networkhookapi.NetworkHookType_PacketMMAP,
// 			Config: hook.Conf,
// 		}
// 		for _, p := range hook.Pipelines {
// 			switch prog := p.(type) {
// 			case *programs.SampleProgram:
// 				pd := programMarshal{
// 					Type:    programs.Program_Sample,
// 					Program: prog,
// 				}
// 				hd.Programs = append(hd.Programs, pd)
// 			default:
// 				return nil, ErrUnknownProgramType
// 			}
// 		}
// 		result, err := json.Marshal(hd)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return result, nil
// 	}
// 	return nil, ErrUnknownHookType
// }

// UnmarshalJSON returns a representation of the JSON object
// as a network hook type.
// The Marshalled form of NetworkHookWrapper is as so:
// hook: map[
// 			conf:map[afpacket-conf:map[num-blocks:128 block-timeout:10
// 					poll-timeout:0 iface:eth0 frame-size:4096
// 					block-size:524288] type:packet_mmap container-id:ep1]
// 			pipelines:<nil>
// 	]
// Use: logger.Infof("hook: %+v\n", hookMap)
//
//  We obtain the hooktype from the conf map. We obtain Pipeline info
// from the second map.
func (wrap *NetworkHookWrapper) UnmarshalJSON(data []byte) error {

	hookType, err := getNetworkHookType(data)
	if err != nil {
		return err
	}

	var rawJson json.RawMessage
	env := networkHookUnmarshal{
		NetworkHook: &rawJson,
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	// logger.Infof("rawJson: %+v\n", string(rawJson)) // REMOVE

	switch hookType {
	case networkhookapi.NetworkHookType_PacketMMAP:
		hook := &packetmmap.PacketMMAP{}
		if err := json.Unmarshal(rawJson, &hook); err != nil {
			return err
		}
		wrap.NetworkHook = hook
		return nil
	case networkhookapi.NetworkHookType_NetfilterQueue:
		hook := &netfilterqueue.NetfilterQueue{}
		if err := json.Unmarshal(rawJson, &hook); err != nil {
			return err
		}
		wrap.NetworkHook = hook
		return nil
	case networkhookapi.NetworkHookType_BPF:
		hook := &bpf.BPF{}
		if err := json.Unmarshal(rawJson, &hook); err != nil {
			return err
		}
		wrap.NetworkHook = hook
		return nil
	}
	return types.ErrUnknownHookType
}

// getNetworkHookType will extract the hook type from the []byte
func getNetworkHookType(data []byte) (string, error) {
	tmp := networkHookUnmarshal{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return "", err
	}
	hookMap := tmp.NetworkHook.(map[string]interface{})
	stateMap := hookMap["state"].(map[string]interface{})
	confMap := stateMap["conf"].(map[string]interface{})
	hookType, ok := confMap["type"]
	if !ok {
		return "", fmt.Errorf("no hook type found")
	}
	return hookType.(string), nil
}

// DeepCopy will go through each hook type and copy the desired fields
// IMPORTANT: the returned object is only for informational (e.g. GET)
// operations and should NOT be used to update damon state!
// func (wrap *NetworkHookWrapper) DeepCopy() (*NetworkHookWrapper, error) {
// 	switch hook := wrap.NetworkHook.(type) {
// 	case *packetmmap.PacketMMAP:
// 		// hook.RLock()
// 		// defer hook.RUnlock()
// 		// cpy := &packetmmap.PacketMMAP{
// 		// 	Conf: hook.Conf,
// 		// }
// 		// for _, p := range hook.Pipelines {
// 		// 	cpy.Pipelines = append(cpy.Pipelines, p)
// 		// }
// 		cpy := hook.Copy()
// 		return &NetworkHookWrapper{cpy}, nil
// 	case *netfilterqueue.NetfilterQueue:
// 		hook.RLock()
// 		defer hook.RUnlock()
// 		cpy := &netfilterqueue.NetfilterQueue{
// 			Conf: hook.Conf,
// 		}
// 		for _, p := range hook.Pipelines {
// 			cpy.Pipelines = append(cpy.Pipelines, p)
// 		}
// 		return &NetworkHookWrapper{cpy}, nil

// 	}
// 	return nil, types.ErrUnknownHookType
// }

func (wrap *NetworkHookWrapper) String() string {
	return wrap.NetworkHook.String()
}
