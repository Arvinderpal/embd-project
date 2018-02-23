package adaptors

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"

	"gobot.io/x/gobot/platforms/raspi"
)

// RaspiConf implements programapi.ProgramConf interface
type RaspiConf struct {
	//////////////////////////////////////////////////////
	// All adaptor confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID   string `json:"machine-id"`
	AdaptorType string `json:"adaptor-type"`
	ID          string `json:"id"` // unique

	////////////////////////////////////////////
	// The fields below are adaptor specific. //
	////////////////////////////////////////////
}

func (c RaspiConf) ValidateConf() error {
	if c.AdaptorType == "" {
		return fmt.Errorf("no adaptor type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.AdaptorType != Adaptor_Raspi {
		return fmt.Errorf("Invalid adaptor type specified. Expected %s, but got %s", Adaptor_Raspi, c.AdaptorType)
	}
	return nil
}

func (c RaspiConf) GetType() string {
	return c.AdaptorType
}

func (c RaspiConf) GetID() string {
	return c.ID
}

func (c RaspiConf) NewAdaptor() (adaptorapi.Adaptor, error) {

	adaptor := Raspi{
		State: &raspiInternal{
			Conf:    c,
			adaptor: raspi.NewAdaptor(),
		},
	}
	return &adaptor, nil
}

// Raspi implements the Adaptor interface
type Raspi struct {
	mu    sync.RWMutex
	State *raspiInternal
}

type raspiInternal struct {
	Conf    RaspiConf `json:"conf"`
	adaptor *raspi.Adaptor
}

// Attach: attaches the adaptor.
func (d *Raspi) Attach() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.Connect()
}

// Detach: detaches the adaptor.
func (d *Raspi) Detach() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.Finalize()
}

func (d *Raspi) DigitalWrite(pin string, level byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// fmt.Printf("DigitalWrite on pin: %s", pin)
	return d.State.adaptor.DigitalWrite(pin, level)
}

func (d *Raspi) PwmWrite(pin string, level byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// fmt.Printf("PwmWrite on pin: %s", pin)
	err = d.State.adaptor.PwmWrite(pin, level)
	return err
}

func (d *Raspi) DigitalRead(pin string) (val int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.DigitalRead(pin)
}

func (d *Raspi) ServoWrite(pin string, angle byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.ServoWrite(pin, angle)
}

func (d *Raspi) AnalogRead(pin string) (val int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return 0, fmt.Errorf("AnalogRead not supported on adaptor %s", d.State.Conf.GetType())
}

func (d *Raspi) GetConf() adaptorapi.AdaptorConf {
	return d.State.Conf
}

func (d *Raspi) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *Raspi) Copy() adaptorapi.Adaptor {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &Raspi{
		State: &raspiInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *Raspi) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *Raspi) UnmarshalJSON(data []byte) error {
	d.State = &raspiInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
