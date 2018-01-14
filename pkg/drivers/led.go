package drivers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/common/message"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

// LEDConf implements programapi.ProgramConf interface
type LEDConf struct {
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
	LEDPin string `json:"led-pin"`
}

func (c LEDConf) ValidateConf() error {
	if c.DriverType == "" {
		return fmt.Errorf("no driver type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no name specified")
	}
	if c.DriverType != Driver_LED {
		return fmt.Errorf("Invalid driver type specified. Expected %s, but got %s", Driver_LED, c.DriverType)
	}
	return nil
}

func (c LEDConf) GetType() string {
	return c.DriverType
}

func (c LEDConf) GetID() string {
	return c.ID
}

func (c LEDConf) GetAdaptorID() string {
	return c.AdaptorID
}

func (c LEDConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c LEDConf) NewDriver(apiAdpt adaptorapi.Adaptor, rcvQ *message.Queue, sndQ *message.Queue) (driverapi.Driver, error) {

	led := gpio.NewLedDriver(apiAdpt, c.LEDPin)

	drv := LED{
		State: &ledInternal{
			Conf:     c,
			led:      led,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}

	drv.State.robot = driverapi.NewRobot(c.ID,
		[]gobot.Device{led},
		drv.work,
	)

	return &drv, nil
}

// LED implements the Driver interface
type LED struct {
	mu    sync.RWMutex
	State *ledInternal
}

type ledInternal struct {
	Conf     LEDConf         `json:"conf"`
	led      *gpio.LedDriver `json:"led"`
	robot    *driverapi.Robot
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}
}

// Start: starts the driver logic.
func (d *LED) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.State.robot.Start()
	return nil
}

// Stop: stops the driver logic.
func (d *LED) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := d.State.robot.Stop()
	if err != nil {
		return err
	}

	d.State.killChan <- struct{}{}
	return nil
}

// work: Runs periodically and generates messages/events.
func (d *LED) work() {

	for {
		select {
		case <-d.State.killChan:
			return
		default:
			// NOTE: Get will block this routine until either the driver is stopeed or a message arrives.
			msg, shutdown := d.State.rcvQ.Get()
			if shutdown {
				return
			}
			fmt.Printf("led-driver received msg: %q", msg)
			d.mu.Lock()
			d.State.led.Toggle()
			d.mu.Unlock()
		}
	}

	// gobot.Every(2*time.Second, func() {
	// 	d.mu.Lock()
	// 	if !d.State.Running {
	// 		d.mu.Unlock()
	// 		// TODO: we sould really use a killchan!
	// 		return
	// 	}
	// 	d.State.led.Toggle()
	// 	d.mu.Unlock()
	// })
}

func (d *LED) GetConf() driverapi.DriverConf {
	return d.State.Conf
}

func (d *LED) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *LED) Copy() driverapi.Driver {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &LED{
		State: &ledInternal{
			Conf: d.State.Conf,
			led:  d.State.led,
			// Data: p.State.Data,
		},
	}
	return cpy
}

func (d *LED) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *LED) UnmarshalJSON(data []byte) error {
	d.State = &ledInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
