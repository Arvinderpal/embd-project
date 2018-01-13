package machine

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

	"github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/pkg/adaptors"
	"github.com/Arvinderpal/embd-project/pkg/drivers"
	"github.com/Arvinderpal/embd-project/pkg/option"

	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("segue-machine")
)

const (
	maxLogs = 16
)

// Machine contains all the details for a particular machine.
type Machine struct {
	mutex     sync.RWMutex
	MachineID string                 `json:"machine-id"`    // Machine ID.
	Adaptors  []*AdaptorWrapper      `json:"adaptors`       // Machine adaptor for communicating with hardware.
	Drivers   []*DriverWrapper       `json:"drivers"`       // Drivers that part of this machine.
	MsgRouter *message.MessageRouter `json:"message-router` // Route messages between drivers/controllers.
	Opts      *option.BoolOptions    `json:"options"`       // Machine options.
	Status    *MachineStatus         `json:"status,omitempty"`
}

type statusLog struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type MachineStatus struct {
	Log     []*statusLog `json:"log,omitempty"`
	Index   int          `json:"index"`
	indexMU sync.RWMutex
}

func (e *MachineStatus) lastIndex() int {
	lastIndex := e.Index - 1
	if lastIndex < 0 {
		return maxLogs - 1
	}
	return lastIndex
}

func (e *MachineStatus) getAndIncIdx() int {
	idx := e.Index
	e.Index++
	if e.Index >= maxLogs {
		e.Index = 0
	}
	return idx
}

func (e *MachineStatus) addStatusLog(s *statusLog) {
	idx := e.getAndIncIdx()
	if len(e.Log) < maxLogs {
		e.Log = append(e.Log, s)
	} else {
		e.Log[idx] = s
	}
}

func (e *MachineStatus) String() string {
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

func (e *MachineStatus) DumpLog() string {
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

func (es *MachineStatus) DeepCopy() *MachineStatus {
	cpy := &MachineStatus{}
	es.indexMU.RLock()
	defer es.indexMU.RUnlock()
	cpy.Index = es.Index
	cpy.Log = []*statusLog{}
	for _, v := range es.Log {
		cpy.Log = append(cpy.Log, v)
	}
	return cpy
}

func (e Machine) Validate() error {
	return nil
}

func (mh *Machine) DeepCopy() *Machine {

	mh.mutex.RLock()
	defer mh.mutex.RUnlock()

	cpy := &Machine{
		MachineID: mh.MachineID,
	}
	if mh.Opts != nil {
		cpy.Opts = mh.Opts.DeepCopy()
	}
	if mh.Status != nil {
		cpy.Status = mh.Status.DeepCopy()
	}
	for _, d := range mh.Drivers {
		dCpy := d.Copy()
		cpy.Drivers = append(cpy.Drivers, &DriverWrapper{dCpy})
	}
	for _, a := range mh.Adaptors {
		aCpy := a.Copy()
		cpy.Adaptors = append(cpy.Adaptors, &AdaptorWrapper{aCpy})
	}
	cpy.MsgRouter = message.NewMessageRouter()
	for msgTyp, set := range mh.MsgRouter.RouteMap {
		var cpySet []*message.Queue
		for _, q := range set {
			cpySet = append(cpySet, &message.Queue{QId: q.ID()})
		}
		cpy.MsgRouter.RouteMap[msgTyp] = cpySet
	}
	return cpy
}

// Snapshot will write the snapshot json in the mh's directory
func (mh *Machine) Snapshot() error {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()
	return mh.snapshot()
}

// snapshot will write the snapshot json in the mh's directory
func (mh *Machine) snapshot() error {

	bytes, err := json.MarshalIndent(mh, "", " ")
	if err != nil {
		return err
	}
	tmpSnap := filepath.Join(".", mh.MachineID, common.SnapshotFileName+".tmp")
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
	oldSnap := path.Join(".", mh.MachineID, common.SnapshotFileName)
	err = os.Rename(tmpSnap, oldSnap)
	return err
}

func (mh *Machine) PrettyPrint() string {
	b, err := json.MarshalIndent(mh, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return fmt.Sprintf(string(b))
}

func OptionChanged(key string, value bool, data interface{}) {
}

func (mh *Machine) ApplyOpts(opts option.OptionMap) bool {
	return mh.Opts.Apply(opts, OptionChanged, mh) > 0
}

func (mh *Machine) SetDefaultOpts(opts *option.BoolOptions) {
}

func (mh *Machine) LogStatus(code StatusCode, msg string) {
	mh.Status.indexMU.Lock()
	defer mh.Status.indexMU.Unlock()
	sts := &statusLog{
		Status: Status{
			Code: code,
			Msg:  msg,
		},
		Timestamp: time.Now(),
	}
	mh.Status.addStatusLog(sts)
}

func (mh *Machine) LogStatusOK(msg string) {
	mh.Status.indexMU.Lock()
	defer mh.Status.indexMU.Unlock()
	sts := &statusLog{
		Status:    NewStatusOK(msg),
		Timestamp: time.Now(),
	}
	mh.Status.addStatusLog(sts)
}

// LookupDriver returns matching driverType
// IMPORTANT: if mutex lock is already held by calling func, use lookupDriver
// instead of this LookupHook
func (mh *Machine) LookupDriver(driverType string) driverapi.Driver {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()
	return mh.lookupDriver(driverType)
}

// lookupDriver returns matching driverType
// IMPORTANT: aquire a mutex lock before calling this func
func (mh *Machine) lookupDriver(driverType string) driverapi.Driver {
	for _, h := range mh.Drivers {
		if h.GetConf().GetType() == driverType {
			return h
		}
	}
	return nil
}

func (mh *Machine) StartDrivers(confs []driverapi.DriverConf) error {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()
	err := mh.startDrivers(confs)
	err2 := mh.snapshot() // NOTE: we snapshot even if error occured during start
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (mh *Machine) startDrivers(confs []driverapi.DriverConf) error {
	if len(confs) <= 0 {
		return fmt.Errorf("No driver(s) specified")
	}

	for _, conf := range confs {
		driverType := conf.GetType()
		drv := mh.lookupDriver(driverType)
		if drv != nil {
			return fmt.Errorf("driver %s already running on machine %s", driverType, mh.MachineID)
		}
		adptID := conf.GetAdaptorID()
		adpt := mh.lookupAdaptor(adptID)
		if adpt == nil {
			return fmt.Errorf("no adapter with id %s found on machine %s", adptID, mh.MachineID)
		}
		// Note: queue id = driver id. we use it to remove entries from the route map when a driver is stopped.
		drvRcvQ := message.NewQueue(conf.GetID())
		drvSndQ := message.NewQueue(conf.GetID())
		drv, err := driverapi.NewDriver(conf, adpt, drvRcvQ, drvSndQ)
		if err != nil {
			return err
		}
		mh.Drivers = append(mh.Drivers, &DriverWrapper{drv})
		// Add subscriptions for desired message types
		subsMsgTypes := conf.GetSubscriptions()
		logger.Infof("Adding subscription: %s", subsMsgTypes)
		for _, sMsgT := range subsMsgTypes {
			err := mh.MsgRouter.AddSubscriber(sMsgT, drvRcvQ)
			if err != nil {
				return err
			}
		}

		err = drv.Start()
		if err != nil {
			mh.LogStatus(Failure, fmt.Sprintf("Could not start driver: %s", err))
			return err
		}
	}

	mh.LogStatusOK(fmt.Sprintf("Sucessfully started drivers on machine %s", mh.MachineID))
	return nil
}

func (mh *Machine) StopDriver(driverType, driverID string) error {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()
	err := mh.stopDriver(driverType, driverID)
	err2 := mh.snapshot()
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (mh *Machine) stopDriver(driverType, driverID string) error {

	for i, drv := range mh.Drivers {
		if drv.GetConf().GetType() == driverType && drv.GetConf().GetID() == driverID {
			err := drv.Stop()
			if err != nil {
				mh.LogStatus(Failure, fmt.Sprintf("Could not stop driver: %s", err))
				return err
			}

			subsMsgTypes := drv.GetConf().GetSubscriptions()
			for _, sMsgT := range subsMsgTypes {
				err := mh.MsgRouter.RemoveSubscriber(sMsgT, drv.GetConf().GetID())
				if err != nil {
					mh.LogStatus(Failure, fmt.Sprintf("could not remove queue subscription on message type %s: %s", sMsgT, err))
					return err
				}
			}

			mh.Drivers = append(mh.Drivers[:i], mh.Drivers[i+1:]...)
			mh.LogStatusOK(fmt.Sprintf("Stopped driver %s (id:%s) on machine %s ", driverType, driverID, mh.MachineID))
			return nil
		}
	}
	return fmt.Errorf("driver %s (id:%s) not found on machine %s", driverType, driverID, mh.MachineID)
}

// LookupAdaptor returns matching adaptorID
// IMPORTANT: if mutex lock is already held by calling func, use lookupAdaptor
// instead of this LookupHook
func (mh *Machine) LookupAdaptor(adaptorID string) adaptorapi.Adaptor {
	mh.mutex.RLock()
	defer mh.mutex.RUnlock()
	return mh.lookupAdaptor(adaptorID)
}

// lookupAdaptor returns matching adaptorID
// IMPORTANT: aquire a mutex lock before calling this func
func (mh *Machine) lookupAdaptor(adaptorID string) adaptorapi.Adaptor {
	for _, h := range mh.Adaptors {
		if h.GetConf().GetID() == adaptorID {
			return h
		}
	}
	return nil
}

func (mh *Machine) AttachAdaptors(confs []adaptorapi.AdaptorConf) error {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()
	err := mh.attachAdaptors(confs)
	err2 := mh.snapshot() // NOTE: we snapshot even if error occured during attach
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (mh *Machine) attachAdaptors(confs []adaptorapi.AdaptorConf) error {
	if len(confs) <= 0 {
		return fmt.Errorf("No adaptor(s) specified")
	}
	for _, conf := range confs {
		adpt, err := adaptorapi.NewAdaptor(conf)
		if err != nil {
			return err
		}
		mh.Adaptors = append(mh.Adaptors, &AdaptorWrapper{adpt})
		err = adpt.Attach()
		if err != nil {
			mh.LogStatus(Failure, fmt.Sprintf("Could not attach adaptor: %s", err))
			return err
		}
	}

	mh.LogStatusOK(fmt.Sprintf("Sucessfully attached adaptor(s) on machine %s", mh.MachineID))
	return nil
}

func (mh *Machine) DetachAdaptor(adaptorType, adaptorID string) error {
	mh.mutex.Lock()
	defer mh.mutex.Unlock()
	err := mh.detachAdaptor(adaptorType, adaptorID)
	err2 := mh.snapshot()
	if err != nil || err2 != nil {
		return fmt.Errorf("Error(s): %s, %s", err, err2)
	}
	return nil
}

func (mh *Machine) detachAdaptor(adaptorType, adaptorID string) error {

	for i, drv := range mh.Adaptors {
		if drv.GetConf().GetType() == adaptorType && drv.GetConf().GetID() == adaptorID {
			err := drv.Detach()
			if err != nil {
				mh.LogStatus(Failure, fmt.Sprintf("Could not detach adaptor: %s", err))
				return err
			}
			mh.Adaptors = append(mh.Adaptors[:i], mh.Adaptors[i+1:]...)
			mh.LogStatusOK(fmt.Sprintf("Detached adaptor %s (id:%s) on machine %s ", adaptorType, adaptorID, mh.MachineID))
			return nil
		}
	}
	return fmt.Errorf("adaptor %s (id:%s) not found on machine %s", adaptorType, adaptorID, mh.MachineID)
}

// DriverWrapper is a helper type used handle the driver type
// based serialization of its configuration.
type DriverWrapper struct {
	driverapi.Driver
}

type driverUnmarshal struct {
	Driver interface{}
}

// UnmarshalJSON creates the driver type specified in the JSON.
func (wrap *DriverWrapper) UnmarshalJSON(data []byte) error {

	driverType, err := getDriverType(data)
	if err != nil {
		return err
	}

	var rawJson json.RawMessage
	env := driverUnmarshal{
		Driver: &rawJson,
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	// logger.Infof("rawJson: %+v\n", string(rawJson)) // REMOVE
	drv, err := drivers.NewDriver(driverType)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rawJson, &drv); err != nil {
		return err
	}
	wrap.Driver = drv
	return nil
}

// getDriverType will extract the driver type from []byte
func getDriverType(data []byte) (string, error) {
	tmp := driverUnmarshal{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return "", err
	}
	driverMap := tmp.Driver.(map[string]interface{})
	confMap := driverMap["conf"].(map[string]interface{})
	driverType, ok := confMap["driver-type"]
	if !ok {
		return "", fmt.Errorf("no driver type found")
	}
	return driverType.(string), nil
}

func (wrap *DriverWrapper) String() string {
	return fmt.Sprintf("%#v", wrap.Driver)
}

// AdaptorWrapper is a helper type used handle the adaptor type
// based serialization of its configuration.
type AdaptorWrapper struct {
	adaptorapi.Adaptor
}

type adaptorUnmarshal struct {
	Adaptor interface{}
}

// UnmarshalJSON creates the adaptor type specified in the JSON.
func (wrap *AdaptorWrapper) UnmarshalJSON(data []byte) error {

	adaptorType, err := getAdaptorType(data)
	if err != nil {
		return err
	}

	var rawJson json.RawMessage
	env := adaptorUnmarshal{
		Adaptor: &rawJson,
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	adpt, err := adaptors.NewAdaptor(adaptorType)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rawJson, &adpt); err != nil {
		return err
	}
	wrap.Adaptor = adpt
	return nil
}

// getAdaptorType will extract the adaptor type from []byte
func getAdaptorType(data []byte) (string, error) {
	tmp := adaptorUnmarshal{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return "", err
	}
	adaptorMap := tmp.Adaptor.(map[string]interface{})
	confMap := adaptorMap["conf"].(map[string]interface{})
	adaptorType, ok := confMap["adaptor-type"]
	if !ok {
		return "", fmt.Errorf("no adaptor type found")
	}
	return adaptorType.(string), nil
}

func (wrap *AdaptorWrapper) String() string {
	return fmt.Sprintf("%#v", wrap.Adaptor)
}
