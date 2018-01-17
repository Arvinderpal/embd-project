package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/pkg/controllers"
)

func (d *Daemon) StartControllers(confB []byte) error {

	// We first unmarshall into a ControllersConfEnvelope to get controller fields
	// such as id and container, and also to determine the exact
	// number of controllers specified.
	env := controllerapi.ControllersConfEnvelope{}
	if err := json.Unmarshal(confB, &env); err != nil {
		return err
	}
	mh := d.lookupMachine(env.MachineID)
	if mh == nil {
		return fmt.Errorf("Could not find machine %s", env.MachineID)
	}
	confs, err := controllers.NewControllerConfs(env)
	if err != nil {
		return err
	}
	if len(confs) <= 0 {
		return fmt.Errorf("No controller configurations specified")
	}

	err = mh.StartControllers(confs)
	if err != nil {
		return err
	}

	logger.Infof("Start controllers successful on machine: %s", env.MachineID)
	return nil
}

func (d *Daemon) StopController(machineID, controllerID string) error {
	// lookup mh associated with id
	mh := d.lookupMachine(machineID)
	if mh == nil {
		return fmt.Errorf("Could not find machine associated with %s", machineID)
	}
	err := mh.StopController(controllerID)
	if err != nil {
		return err
	}

	logger.Infof("Stop controller successful on machine: %s", machineID)
	return nil
}
