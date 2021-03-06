package drivers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"

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
	MachineID     string   `json:"machine-id"`
	ID            string   `json:"id"`
	DriverType    string   `json:"driver-type"`
	AdaptorID     string   `json:"adaptor-id"`
	Qualifier     string   `json:"qualifier"`
	Subscriptions []string `json:"subscriptions"` // Message Type Subscriptions.
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

func (c DualMotorsConf) GetQualifier() string {
	return c.Qualifier
}

func (c DualMotorsConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c DualMotorsConf) NewDriver(apiAdpt adaptorapi.Adaptor, rcvQ *message.Queue, sndQ *message.Queue) (driverapi.Driver, error) {

	// writer, ok := apiAdpt.(gpio.DigitalWriter)
	// if !ok {
	// 	return nil, fmt.Errorf("adaptor does not implement DigitalWriter interface required by driver")
	// }
	rightmotor := gpio.NewMotorDriver(apiAdpt, c.RightMotor.SpeedPin)
	rightmotor.ForwardPin = c.RightMotor.ForwardPin
	rightmotor.BackwardPin = c.RightMotor.BackwardPin

	leftmotor := gpio.NewMotorDriver(apiAdpt, c.LeftMotor.SpeedPin)
	leftmotor.ForwardPin = c.LeftMotor.ForwardPin
	leftmotor.BackwardPin = c.LeftMotor.BackwardPin

	drv := DualMotors{
		State: &dualMotorsInternal{
			Conf:       c,
			RightMotor: rightmotor,
			LeftMotor:  leftmotor,
			killChan:   make(chan struct{}),
			rcvQ:       rcvQ,
			sndQ:       sndQ,
		},
	}

	drv.State.robot = driverapi.NewRobot(c.ID,
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
	RightMotor *gpio.MotorDriver `json:"right-motor"`
	LeftMotor  *gpio.MotorDriver `json:"left-motor"`
	robot      *driverapi.Robot
	Running    bool `json:"running"`
	killChan   chan struct{}
	rcvQ       *message.Queue
	sndQ       *message.Queue
}

// Start: starts the driver logic.
func (d *DualMotors) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.State.robot.Start()
	d.State.Running = true
	return nil
}

// Stop: stops the driver logic.
func (d *DualMotors) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := d.State.robot.Stop()
	if err != nil {
		return err
	}
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	d.State.Running = false
	close(d.State.killChan) // broadcast
	return nil
}

// work: processes motor commands from controllers.
func (d *DualMotors) work() {
	for {
		select {
		case <-d.State.killChan:
			return
		default:
			// NOTE: Get will block this routine until either the controller is stopped or a message arrives.
			msg, shutdown := d.State.rcvQ.Get()
			if shutdown {
				logger.Debugf("stopping worker on driver %s", d.State.Conf.GetID())
				return
			}
			if msg.ID.Qualifier == d.State.Conf.Qualifier {
				logger.Debugf("dualmotors: received msg: %q", msg)
				switch msg.ID.Type {
				case seguepb.MessageType_CmdDrive:
					d.ProcessDriveCmd(msg)
				default:
					logger.Errorf("dualmotors: unknown message type %s", msg.ID.Type)
				}
			}
			d.State.rcvQ.Done(msg)
		}
	}
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
			Conf:       d.State.Conf,
			RightMotor: d.State.RightMotor,
			LeftMotor:  d.State.LeftMotor,
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

// ProcessDriveCmd: activates the motor according to message.
func (d *DualMotors) ProcessDriveCmd(msg message.Message) {
	d.mu.Lock()
	defer d.mu.Unlock()
	cmd := msg.ID.SubType
	switch cmd {
	case "forward":
		speed := msg.Data.(*seguepb.CmdDriveData).Speed
		d.State.RightMotor.Forward(byte(speed))
		d.State.LeftMotor.Forward(byte(speed))
	case "backward":
		speed := msg.Data.(*seguepb.CmdDriveData).Speed
		d.State.RightMotor.Backward(byte(speed))
		d.State.LeftMotor.Backward(byte(speed))
	case "stop":
		d.State.RightMotor.Forward(byte(0))
		d.State.LeftMotor.Forward(byte(0))
	case "left":
		speed := msg.Data.(*seguepb.CmdDriveData).Speed
		d.State.RightMotor.Forward(byte(speed))
		d.State.LeftMotor.Backward(byte(speed))
	case "right":
		speed := msg.Data.(*seguepb.CmdDriveData).Speed
		d.State.RightMotor.Backward(byte(speed))
		d.State.LeftMotor.Forward(byte(speed))
	case "forward-right":
		speed := msg.Data.(*seguepb.CmdDriveData).Speed
		d.State.RightMotor.Forward(byte(speed - 50))
		d.State.LeftMotor.Forward(byte(speed + 50))
	case "forward-left":
		speed := msg.Data.(*seguepb.CmdDriveData).Speed
		d.State.RightMotor.Forward(byte(speed + 50))
		d.State.LeftMotor.Forward(byte(speed - 50))
	default:
		logger.Errorf("dual-motors: unknown message sub-type: %s", cmd)
	}
	return
}
