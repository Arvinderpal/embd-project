package drivers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"

	"gobot.io/x/gobot"
)

// UnitTestConf implements programapi.ProgramConf interface
type UnitTestConf struct {
	//////////////////////////////////////////////////////
	// All driver confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID     string   `json:"machine-id"`
	ID            string   `json:"id"`
	DriverType    string   `json:"driver-type"`
	AdaptorID     string   `json:"adaptor-id"`
	Subscriptions []string `json:"subscriptions"` // Message Type Subscriptions.

	////////////////////////////////////////////
	// The fields below are driver specific. //
	////////////////////////////////////////////
	MessageSendInterval time.Duration `json:"message-send-interval"` // rate at which messages are sent (units are Milliseconds)
}

func (c UnitTestConf) ValidateConf() error {
	if c.DriverType == "" {
		return fmt.Errorf("no driver type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no name specified")
	}
	if c.DriverType != Driver_UnitTest {
		return fmt.Errorf("Invalid driver type specified. Expected %s, but got %s", Driver_UnitTest, c.DriverType)
	}
	return nil
}

func (c UnitTestConf) GetType() string {
	return c.DriverType
}

func (c UnitTestConf) GetID() string {
	return c.ID
}

func (c UnitTestConf) GetAdaptorID() string {
	return c.AdaptorID
}

func (c UnitTestConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c UnitTestConf) NewDriver(apiAdpt adaptorapi.Adaptor, rcvQ *message.Queue, sndQ *message.Queue) (driverapi.Driver, error) {

	if c.MessageSendInterval == time.Duration(0) {
		c.MessageSendInterval = 100
	}

	drv := UnitTest{
		State: &unittestInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}

	drv.State.robot = driverapi.NewRobot(c.ID,
		[]gobot.Device{},
		drv.work,
	)

	return &drv, nil
}

// UnitTest implements the Driver interface
type UnitTest struct {
	mu    sync.RWMutex
	State *unittestInternal
}

type unittestInternal struct {
	Conf     UnitTestConf `json:"conf"`
	robot    *driverapi.Robot
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}
}

// Start: starts the driver logic.
func (d *UnitTest) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.State.robot.Start()
	return nil
}

// Stop: stops the driver logic.
func (d *UnitTest) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := d.State.robot.Stop()
	if err != nil {
		return err
	}
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)
	return nil
}

// work: Runs periodically and generates messages/events.
func (d *UnitTest) work() {

	version := 0
	tickChan := time.NewTicker(d.State.Conf.MessageSendInterval * time.Millisecond).C
	for {
		select {
		case <-d.State.killChan:
			return
		case <-tickChan:
			msg := message.Message{
				ID: seguepb.Message_MessageID{
					Type:    seguepb.MessageType_UnitTest,
					SubType: d.State.Conf.ID,
					Version: uint64(version),
				},
				Data: &seguepb.UnitTestData{TestMessage: fmt.Sprintf("msg: %d", version)},
			}
			d.State.sndQ.Add(msg)
			version += 1
			logger.Debugf("unittest-driver sent msg: %q", msg)
		default:
			// NOTE: Get will block this routine until either the driver is stopeed or a message arrives.
			msg, shutdown := d.State.rcvQ.Get()
			if shutdown {
				logger.Debugf("stopping worker on driver %s", d.State.Conf.GetID())
				return
			}
			d.State.rcvQ.Done(msg)
			logger.Debugf("unittest-driver received msg: %q", msg)
		}
	}
}

func (d *UnitTest) GetConf() driverapi.DriverConf {
	return d.State.Conf
}

func (d *UnitTest) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *UnitTest) Copy() driverapi.Driver {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &UnitTest{
		State: &unittestInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *UnitTest) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *UnitTest) UnmarshalJSON(data []byte) error {
	d.State = &unittestInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
