package adaptors

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
)

// UnitTestConf implements programapi.ProgramConf interface
type UnitTestConf struct {
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

func (c UnitTestConf) ValidateConf() error {
	if c.AdaptorType == "" {
		return fmt.Errorf("no adaptor type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.AdaptorType != Adaptor_UnitTest {
		return fmt.Errorf("Invalid adaptor type specified. Expected %s, but got %s", Adaptor_UnitTest, c.AdaptorType)
	}
	return nil
}

func (c UnitTestConf) GetType() string {
	return c.AdaptorType
}

func (c UnitTestConf) GetID() string {
	return c.ID
}

func (c UnitTestConf) NewAdaptor() (adaptorapi.Adaptor, error) {

	adaptor := UnitTest{
		State: &unittestInternal{
			Conf: c,
		},
	}
	return &adaptor, nil
}

// UnitTest implements the Adaptor interface
type UnitTest struct {
	mu    sync.RWMutex
	State *unittestInternal
}

type unittestInternal struct {
	Conf UnitTestConf `json:"conf"`
}

// Attach: attaches the adaptor.
func (d *UnitTest) Attach() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("Attach")
	return nil
}

// Detach: detaches the adaptor.
func (d *UnitTest) Detach() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("Detach")
	return nil
}

func (d *UnitTest) DigitalWrite(pin string, level byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("DigitalWrite on pin: %s", pin)
	return err
}

func (d *UnitTest) PwmWrite(pin string, level byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("PwmWrite on pin: %s", pin)
	return err
}

func (d *UnitTest) DigitalRead(pin string) (val int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("DigitalRead on pin: %s", pin)
	return 42, err
}

func (d *UnitTest) ServoWrite(pin string, angle byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("ServoWrite on pin: %s angle %d", pin, angle)
	return err
}

func (d *UnitTest) AnalogRead(pin string) (val int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	logger.Infof("AnalogRead on pin: %s", pin)
	return 42, err
}

func (d *UnitTest) GetConf() adaptorapi.AdaptorConf {
	return d.State.Conf
}

func (d *UnitTest) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *UnitTest) Copy() adaptorapi.Adaptor {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &UnitTest{
		State: &unittestInternal{
			Conf: d.State.Conf,
			// Data: p.State.Data,
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
