package adaptors

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"

	"gobot.io/x/gobot/platforms/firmata"
)

// FirmataSerialConf implements programapi.ProgramConf interface
type FirmataSerialConf struct {
	//////////////////////////////////////////////////////
	// All adaptor confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID   string `json:"machine-id"`
	AdaptorType string `json:"adaptor-type"`
	ID          string `json:"id"` // unique

	////////////////////////////////////////////
	// The fields below are adaptor specific. //
	////////////////////////////////////////////

	Address string `json:"address"` // address of serial port (e.g. "/dev/ttyACM0")
}

func (c FirmataSerialConf) ValidateConf() error {
	if c.AdaptorType == "" {
		return fmt.Errorf("no adaptor type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.AdaptorType != Adaptor_Firmata_Serial {
		return fmt.Errorf("Invalid adaptor type specified. Expected %s, but got %s", Adaptor_Firmata_Serial, c.AdaptorType)
	}
	return nil
}

func (c FirmataSerialConf) GetType() string {
	return c.AdaptorType
}

func (c FirmataSerialConf) GetID() string {
	return c.ID
}

func (c FirmataSerialConf) NewAdaptor() (adaptorapi.Adaptor, error) {

	adaptor := FirmataSerial{
		State: &firmataInternal{
			Conf:    c,
			adaptor: firmata.NewAdaptor(c.Address),
		},
	}
	return &adaptor, nil
}

// FirmataSerial implements the Adaptor interface
type FirmataSerial struct {
	mu    sync.RWMutex
	State *firmataInternal
}

type firmataInternal struct {
	Conf    FirmataSerialConf `json:"conf"`
	adaptor *firmata.Adaptor
}

// Attach: attaches the adaptor.
func (d *FirmataSerial) Attach() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.Connect()
}

// Detach: detaches the adaptor.
func (d *FirmataSerial) Detach() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.Finalize()
}

func (d *FirmataSerial) DigitalWrite(pin string, level byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// fmt.Printf("DigitalWrite on pin: %s", pin)
	return d.State.adaptor.DigitalWrite(pin, level)
}

func (d *FirmataSerial) PwmWrite(pin string, level byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// fmt.Printf("PwmWrite on pin: %s", pin)
	err = d.State.adaptor.PwmWrite(pin, level)
	return err
}

func (d *FirmataSerial) DigitalRead(pin string) (val int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.DigitalRead(pin)
}

func (d *FirmataSerial) ServoWrite(pin string, angle byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.ServoWrite(pin, angle)
}

func (d *FirmataSerial) AnalogRead(pin string) (val int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.State.adaptor.AnalogRead(pin)
}

func (d *FirmataSerial) GetConf() adaptorapi.AdaptorConf {
	return d.State.Conf
}

func (d *FirmataSerial) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *FirmataSerial) Copy() adaptorapi.Adaptor {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &FirmataSerial{
		State: &firmataInternal{
			Conf: d.State.Conf,
			// Data: p.State.Data,
		},
	}
	return cpy
}

func (d *FirmataSerial) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *FirmataSerial) UnmarshalJSON(data []byte) error {
	d.State = &firmataInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}
