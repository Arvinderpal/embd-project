package adaptors

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
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
	return nil
}

// Detach: detaches the adaptor.
func (d *FirmataSerial) Detach() error {
	return d.State.adaptor.Finalize()
}

func (d *FirmataSerial) GetGobotAdaptor() gobot.Adaptor {
	return d.State.adaptor
}

func (d *FirmataSerial) GetDigitalWriter() (gpio.DigitalWriter, error) {
	return d.State.adaptor, nil
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
