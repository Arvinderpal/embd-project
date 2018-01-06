package drivers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/driverapi"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

type MotorConf struct {
	ForwardPin  string `json:"forward-pin"`
	BackwardPin string `json:"backward-pin"`
	SpeedPin    string `json:"speed-pin"`
}

// DualMotorsConf implements programapi.ProgramConf interface
type DualMotorsConf struct {
	//////////////////////////////////////////////////////
	// All driver confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID  string `json:"machine-id"`
	ID         string `json:"id"`
	DriverType string `json:"driver-type"`
	AdaptorID  string `json:"adaptor-id"`

	////////////////////////////////////////////
	// The fields below are driver specific. //
	////////////////////////////////////////////

	RightMotor MotorConf `json:"right-motor"`
	LeftMotor  MotorConf `json:"left-motor"`

	LogFilePathname string `json:"log-file-path-name"` // logs will be wirten to this file.
}

func (c DualMotorsConf) ValidateConf() error {
	if c.DriverType == "" {
		return fmt.Errorf("no driver type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no name specified")
	}
	if c.DriverType != Driver_DualMotors {
		return fmt.Errorf("Invalid driver type specified. Expected %s, but got %s", Driver_DualMotors, c.DriverType)
	}
	return nil
}

func (c DualMotorsConf) GetType() string {
	return c.DriverType
}

func (c DualMotorsConf) GetID() string {
	return c.ID
}

func (c DualMotorsConf) GetAdaptorID() string {
	return c.AdaptorID
}

func (c DualMotorsConf) NewDriver(apiAdpt adaptorapi.Adaptor) (driverapi.Driver, error) {

	adptDWr, err := apiAdpt.GetDigitalWriter()
	if err != nil {
		return nil, err
	}
	rightmotor := gpio.NewMotorDriver(adptDWr, c.RightMotor.SpeedPin)
	rightmotor.ForwardPin = c.RightMotor.ForwardPin
	rightmotor.BackwardPin = c.RightMotor.BackwardPin

	leftmotor := gpio.NewMotorDriver(adptDWr, c.LeftMotor.SpeedPin)
	leftmotor.ForwardPin = c.LeftMotor.ForwardPin
	leftmotor.BackwardPin = c.LeftMotor.BackwardPin

	drv := DualMotors{
		State: &dualMotorsInternal{
			Conf:       c,
			rightMotor: rightmotor,
			leftMotor:  leftmotor,
		},
	}

	drv.State.robot = gobot.NewRobot(c.ID,
		[]gobot.Connection{apiAdpt.GetGobotAdaptor()},
		[]gobot.Device{rightmotor, leftmotor},
		drv.work,
	)

	return &drv, nil
}

// DualMotors implements the Driver interface
type DualMotors struct {
	mu    sync.RWMutex
	State *dualMotorsInternal
}

type dualMotorsInternal struct {
	Conf       DualMotorsConf    `json:"conf"`
	rightMotor *gpio.MotorDriver `json:"right-motor"`
	leftMotor  *gpio.MotorDriver `json:"left-motor"`
	robot      *gobot.Robot      `json:"robot"`
}

// Start: starts the driver logic.
func (d *DualMotors) Start() error {
	return d.State.robot.Start()
}

// Stop: stops the driver logic.
func (d *DualMotors) Stop() error {
	return d.State.robot.Stop()
}

// ProcessMessage: processes messages (i.e. commands such as move forward, backwards...)
func (d *DualMotors) ProcessMessage() {
	d.mu.Lock()
	defer d.mu.Unlock()
}

// work: Runs periodically and generates messages/events.
func (d *DualMotors) work() {
	speed := byte(0)
	fadeAmount := byte(15)

	d.mu.Lock()
	d.State.rightMotor.Forward(speed)
	d.mu.Unlock()
	gobot.Every(500*time.Millisecond, func() {
		d.mu.Lock()
		d.State.rightMotor.Speed(speed)
		speed = speed + fadeAmount
		if speed == 0 || speed == 255 {
			fadeAmount = -fadeAmount
		}
		d.mu.Unlock()
		fmt.Printf("%d, ", speed) // TODO: move logs to the logfile.
	})
}

func (d *DualMotors) GetConf() driverapi.DriverConf {
	return d.State.Conf
}

func (d *DualMotors) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *DualMotors) Copy() driverapi.Driver {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &DualMotors{
		State: &dualMotorsInternal{
			Conf: d.State.Conf,
			// Data: p.State.Data,
		},
	}
	return cpy
}

func (d *DualMotors) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *DualMotors) UnmarshalJSON(data []byte) error {
	d.State = &dualMotorsInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
