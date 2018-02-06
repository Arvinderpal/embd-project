package adaptors

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/types"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("segue-adaptors")
)

// List of all adaptors
const (
	Adaptor_UnitTest       = "adaptor_unittest"
	Adaptor_Firmata_Serial = "adaptor_firmata_serial"
	// Adaptor_Firmata_TCP    = "adaptor_firmata_tcp"
	// Adaptor_Firmata_BLE    = "adaptor_firmata_ble"
	Adaptor_Raspi = "adaptor_raspi"
)

// NewConf is a util method used to get adaptor conf of a particular type.
// The idea is to localize new adaptor creation code to this package, so that
// each time a new adaptor is added, the below code can be updated.
func NewAdaptorConf(adaptorType, adaptorID, machineID string) (adaptorapi.AdaptorConf, error) {
	switch adaptorType {
	case Adaptor_UnitTest:
		return &UnitTestConf{
			MachineID:   machineID,
			AdaptorType: adaptorType,
			ID:          adaptorID,
		}, nil
	case Adaptor_Firmata_Serial:
		return &FirmataSerialConf{
			MachineID:   machineID,
			AdaptorType: adaptorType,
			ID:          adaptorID,
		}, nil
	case Adaptor_Raspi:
		return &RaspiConf{
			MachineID:   machineID,
			AdaptorType: adaptorType,
			ID:          adaptorID,
		}, nil
	default:
		return nil, types.ErrUnknownAdaptorType

	}
}

// NewAdaptor returns an emtpty Adaptor object of the desired type.
func NewAdaptor(t string) (adaptorapi.Adaptor, error) {
	switch t {
	case Adaptor_UnitTest:
		return &UnitTest{}, nil
	case Adaptor_Firmata_Serial:
		return &FirmataSerial{}, nil
	case Adaptor_Raspi:
		return &Raspi{}, nil
	default:
		return nil, types.ErrUnknownAdaptorType
	}
}

func NewAdaptorConfs(env adaptorapi.AdaptorsConfEnvelope) ([]adaptorapi.AdaptorConf, error) {
	var returnAConfs []adaptorapi.AdaptorConf

	// The adaptor confs contains 1 or more adaptor configurations.
	for i, aEnv := range env.Confs {
		// This is a little trick to make unmarshalling of the AdaptorConf a
		// little easier. There may be an alternative (easier) way...
		var jc json.RawMessage
		rawAEnv := adaptorapi.AdaptorConfEnvelope{
			Conf: &jc,
		}
		// AdaptorConfEnvelope (aEnv) is type Conf:map[string]interface{}.
		// In order to Unmarshall, we have to put back into a json []byte.
		bytes, err := json.Marshal(aEnv)
		if err != nil {
			return nil, fmt.Errorf("Marshal of adaptor conf %d (into []byte) failed: %s\n", i, err)
		}
		if err := json.Unmarshal(bytes, &rawAEnv); err != nil {
			return nil, fmt.Errorf("Unmarshal of adaptor conf %d (raw) failed: %s", i, err)
		}

		aConf, err := NewAdaptorConf(aEnv.Type, aEnv.ID, env.MachineID)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jc, &aConf); err != nil {
			return nil, fmt.Errorf("Unmarshal of %s failed: %s", aEnv.Type, err)
		}
		err = aConf.ValidateConf()
		if err != nil {
			return nil, err
		}
		returnAConfs = append(returnAConfs, aConf)
	}
	return returnAConfs, nil
}
