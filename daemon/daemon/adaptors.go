package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/pkg/adaptors"
)

func (d *Daemon) AttachAdaptors(confB []byte) error {

	// We first unmarshall into a AdaptorsConfEnvelope to get  machine id,
	// and also to determine the exact number of adaptors specified.
	env := adaptorapi.AdaptorsConfEnvelope{}
	if err := json.Unmarshal(confB, &env); err != nil {
		return err
	}
	mh := d.lookupMachine(env.MachineID)
	if mh == nil {
		return fmt.Errorf("Could not find machine associated with %s", env.MachineID)
	}
	confs, err := adaptors.NewAdaptorConfs(env)
	if err != nil {
		return err
	}
	if len(confs) <= 0 {
		return fmt.Errorf("No adaptor configurations specified")
	}

	err = mh.AttachAdaptors(confs)
	if err != nil {
		return err
	}

	logger.Infof("Attached adaptor(s) successful on machine: %s", env.MachineID)
	return nil
}

func (d *Daemon) DetachAdaptor(machineID, adaptorType, adaptorID string) error {
	// lookup mh associated with id
	mh := d.lookupMachine(machineID)
	if mh == nil {
		return fmt.Errorf("Could not find machine associated with %s", machineID)
	}
	err := mh.DetachAdaptor(adaptorType, adaptorID)
	if err != nil {
		return err
	}

	logger.Infof("Detach adaptor successful on machine: %s", machineID)
	return nil
}
