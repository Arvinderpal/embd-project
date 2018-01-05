package drivers

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/common/types"
)

// List of all drivers
const (
	Driver_UnitTest   = "driver_unittest"
	Driver_DualMotors = "driver_dualmotors"
)

// NewConf is a util method used to get driver conf of a particular type.
// The idea is to localize driver creation code to this package, so that each
// time a new driver is added, the below code can be updated.
func NewDriverConf(machineID, driverType, driverID string) (driverapi.DriverConf, error) {
	switch driverType {
	// case Driver_UnitTest:
	// 	return &UnitTestDriverConf{
	// 		PipelineID:  pipeID,
	// 		HookType:    hookType,
	// 		ContainerID: machineID,
	// 		DriverType: driverType,
	// 	}, nil
	case Driver_DualMotors:
		return &DualMotorsConf{
			MachineID:  machineID,
			DriverType: driverType,
			ID:         driverID,
		}, nil
	default:
		return nil, types.ErrUnknownDriverType

	}
}

// NewDriver is a util method used by pipeline package (among others).
// It returns an emtpty Driver object of the desired type.
func NewDriver(t string) (driverapi.Driver, error) {
	switch t {
	// case Driver_UnitTest:
	// 	return &UnitTestDriver{}, nil
	case Driver_DualMotors:
		return &DualMotors{}, nil
	default:
		return nil, types.ErrUnknownDriverType

	}
}

func NewDriverConfs(env driverapi.DriversConfEnvelope) ([]driverapi.DriverConf, error) {
	var returnDConfs []driverapi.DriverConf

	// The driver confs contains 1 or more driver configurations.
	for i, dEnv := range env.Confs {
		// This is a little trick to make unmarshalling of the DriverConf a
		// little easier. There may be an alternative (easier) way...
		var jc json.RawMessage
		rawDEnv := driverapi.DriverConfEnvelope{
			Conf: &jc,
		}
		// DriverConfEnvelope (dEnv) is type Conf:map[string]interface{}.
		// In order to Unmarshall, we have to put back into a json []byte.
		bytes, err := json.Marshal(dEnv)
		if err != nil {
			return nil, fmt.Errorf("Marshal of driver conf %d (into []byte) failed: %s\n", i, err)
		}
		if err := json.Unmarshal(bytes, &rawDEnv); err != nil {
			return nil, fmt.Errorf("Unmarshal of driver conf %d (raw) failed: %s", i, err)
		}

		dConf, err := NewDriverConf(dEnv.Type, dEnv.ID, env.MachineID)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jc, &dConf); err != nil {
			return nil, fmt.Errorf("Unmarshal of %s failed: %s", dEnv.Type, err)
		}
		err = dConf.ValidateConf()
		if err != nil {
			return nil, err
		}
		returnDConfs = append(returnDConfs, dConf)
	}
	return returnDConfs, nil
}
