package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/pkg/drivers"
)

func (d *Daemon) StartDrivers(confB []byte) error {

	// We first unmarshall into a DriversConfEnvelope to get driver fields
	// such as id, driver type and container, and also to determine the exact
	// number of drivers specified.
	env := driverapi.DriversConfEnvelope{}
	if err := json.Unmarshal(confB, &env); err != nil {
		return err
	}
	mh := d.lookupMachine(env.MachineID)
	if mh == nil {
		return fmt.Errorf("Could not find machine associated with %s", env.MachineID)
	}
	confs, err := drivers.NewDriverConfs(env)
	if err != nil {
		return err
	}
	if len(confs) <= 0 {
		return fmt.Errorf("No driver configurations specified")
	}

	err = mh.StartDrivers(confs)
	if err != nil {
		return err
	}

	logger.Infof("Start drivers successful on machine: %s", env.MachineID)
	return nil
}

func (d *Daemon) StopDriver(machineID, driverType, driverID string) error {
	// lookup mh associated with id
	mh := d.lookupMachine(machineID)
	if mh == nil {
		return fmt.Errorf("Could not find machine associated with %s", machineID)
	}
	err := mh.StopDriver(driverType, driverID)
	if err != nil {
		return err
	}

	logger.Infof("Stop driver successful on machine: %s", machineID)
	return nil
}
