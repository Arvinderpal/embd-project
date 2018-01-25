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
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/drivers/gpio"
)

// UltraSonicConf implements programapi.ProgramConf interface
type UltraSonicConf struct {
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
	TrigPin string `json:"trig-pin"`
	EchoPin string `json:"echo-pin"`
}

func (c UltraSonicConf) ValidateConf() error {
	if c.DriverType == "" {
		return fmt.Errorf("no driver type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no name specified")
	}
	if c.DriverType != Driver_UltraSonic {
		return fmt.Errorf("Invalid driver type specified. Expected %s, but got %s", Driver_UltraSonic, c.DriverType)
	}
	return nil
}

func (c UltraSonicConf) GetType() string {
	return c.DriverType
}

func (c UltraSonicConf) GetID() string {
	return c.ID
}

func (c UltraSonicConf) GetAdaptorID() string {
	return c.AdaptorID
}

func (c UltraSonicConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c UltraSonicConf) NewDriver(apiAdpt adaptorapi.Adaptor, rcvQ *message.Queue, sndQ *message.Queue) (driverapi.Driver, error) {

	// trig := gpio.NewDirectPinDriver(apiAdpt, c.TrigPin)
	trig := gpio.NewLedDriver(apiAdpt, c.TrigPin)
	echo := aio.NewAnalogSensorDriver(apiAdpt, c.EchoPin)

	drv := UltraSonic{
		State: &ultraSonicInternal{
			Conf:     c,
			trig:     trig,
			echo:     echo,
			Running:  false,
			killChan: make(chan struct{}),
			rcvQ:     rcvQ,
			sndQ:     sndQ,
		},
	}

	drv.State.robot = driverapi.NewRobot(c.ID,
		[]gobot.Device{trig, echo},
		drv.work,
	)

	return &drv, nil
}

// UltraSonic implements the Driver interface
type UltraSonic struct {
	mu    sync.RWMutex
	State *ultraSonicInternal
}

type ultraSonicInternal struct {
	Conf UltraSonicConf `json:"conf"`
	// trig     *gpio.DirectPinDriver
	trig     *gpio.LedDriver
	echo     *aio.AnalogSensorDriver
	robot    *driverapi.Robot
	Running  bool `json:"running"`
	killChan chan struct{}
	rcvQ     *message.Queue
	sndQ     *message.Queue
}

// Start: starts the driver logic.
func (d *UltraSonic) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.State.robot.Start()
	d.State.Running = true
	return nil
}

// Stop: stops the driver logic.
func (d *UltraSonic) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := d.State.robot.Stop()
	if err != nil {
		return err
	}
	d.State.Running = false
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan) // broadcast
	return nil
}

// work: Runs periodically and generates messages/events.
func (d *UltraSonic) work() {
	var version uint64
	d.State.echo.On(aio.Data, func(data interface{}) {
		// logger.Infof("ultra-sonic reading:", data)
		msg := message.Message{
			ID: seguepb.Message_MessageID{
				Type:    seguepb.MessageType_SensorUltraSonic,
				SubType: "raw",
				Version: version,
			},
			Data: seguepb.SensorUltraSonicData{EchoSample: int64(data.(int))},
		}
		d.State.sndQ.Add(msg)
		version += 1
	})

	logger.Debugf("Starting Triag...")
	// Assert Trig Pin
	for {
		select {
		case <-d.State.killChan:
			logger.Debugf("stopping worker on driver %s", d.State.Conf.GetID())
			return
		default:
			// fmt.Printf("OFF-")
			if err := d.State.trig.Off(); err != nil {
				logger.Errorf("ultrasonic worker: error: %s", err)
			}
			time.Sleep(time.Duration(2) * time.Microsecond)
			// fmt.Printf("ON-")
			if err := d.State.trig.On(); err != nil {
				logger.Errorf("ultrasonic worker: error: %s", err)
			}
			time.Sleep(time.Duration(20) * time.Microsecond)
		}
	}

}

func (d *UltraSonic) GetConf() driverapi.DriverConf {
	return d.State.Conf
}

func (d *UltraSonic) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *UltraSonic) Copy() driverapi.Driver {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &UltraSonic{
		State: &ultraSonicInternal{
			Conf:    d.State.Conf,
			trig:    d.State.trig,
			echo:    d.State.echo,
			robot:   d.State.robot,
			Running: d.State.Running,
		},
	}
	return cpy
}

func (d *UltraSonic) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *UltraSonic) UnmarshalJSON(data []byte) error {
	d.State = &ultraSonicInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
