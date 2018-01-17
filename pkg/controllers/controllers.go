package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/types"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("segue-controllers")
)

// List of all controllers
const (
	Controller_UnitTest        = "unittest"
	Controller_AutonomousDrive = "autonomous-drive"
)

// NewConf is a util method used to get controller conf of a particular type.
// The idea is to localize controller creation code to this package, so that each
// time a new controller is added, the below code can be updated.
func NewControllerConf(controllerType, controllerID, machineID string, subs []string) (controllerapi.ControllerConf, error) {
	switch controllerType {
	// case Controller_UnitTest:
	// 	return &UnitTestControllerConf{
	// 		PipelineID:  pipeID,
	// 		HookType:    hookType,
	// 		ContainerID: machineID,
	// 		ControllerType: controllerType,
	// 	}, nil
	case Controller_AutonomousDrive:
		return &AutonomousDriveConf{
			MachineID:      machineID,
			ControllerType: controllerType,
			ID:             controllerID,
			Subscriptions:  subs,
		}, nil
	default:
		return nil, types.ErrUnknownControllerType

	}
}

// NewController returns an emtpty Controller object of the desired type.
func NewController(t string) (controllerapi.Controller, error) {
	switch t {
	// case Controller_UnitTest:
	// 	return &UnitTestController{}, nil
	case Controller_AutonomousDrive:
		return &AutonomousDrive{}, nil
	default:
		return nil, types.ErrUnknownControllerType
	}
}

func NewControllerConfs(env controllerapi.ControllersConfEnvelope) ([]controllerapi.ControllerConf, error) {
	var returnDConfs []controllerapi.ControllerConf

	// The controller confs contains 1 or more controller configurations.
	for i, dEnv := range env.Confs {
		// This is a little trick to make unmarshalling of the ControllerConf a
		// little easier. There may be an alternative (easier) way...
		var jc json.RawMessage
		rawDEnv := controllerapi.ControllerConfEnvelope{
			Conf: &jc,
		}
		// ControllerConfEnvelope (dEnv) is type Conf:map[string]interface{}.
		// In order to Unmarshall, we have to put back into a json []byte.
		bytes, err := json.Marshal(dEnv)
		if err != nil {
			return nil, fmt.Errorf("Marshal of controller conf %d (into []byte) failed: %s\n", i, err)
		}
		if err := json.Unmarshal(bytes, &rawDEnv); err != nil {
			return nil, fmt.Errorf("Unmarshal of controller conf %d (raw) failed: %s", i, err)
		}

		dConf, err := NewControllerConf(dEnv.Type, dEnv.ID, env.MachineID, dEnv.Subscriptions)
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
