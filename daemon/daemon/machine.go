package daemon

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/pkg/machine"
	"github.com/Arvinderpal/embd-project/pkg/option"
)

func (d *Daemon) lookupMachine(machineID string) *machine.Machine {
	if mh, ok := d.machines[machineID]; ok {
		return mh
	} else {
		return nil
	}
}

// Returns a pointer of a copy machine if the machine was found, nil
// otherwise. It also updates the daemon map with IDs by which the machine
// can be retreived.
func (d *Daemon) getMachineAndUpdateIDs(machineID string) *machine.Machine {
	var (
		mh *machine.Machine
		ok bool
	)

	setIfNotEmpty := func(receiver *string, provider string) {
		if receiver != nil && *receiver == "" && provider != "" {
			*receiver = provider
		}
	}

	d.machinesMU.Lock()
	defer d.machinesMU.Unlock()

	if machineID != "" {
		mh, ok = d.machines[machineID]
	} else {
		return nil
	}

	if ok {
		setIfNotEmpty(&mh.MachineID, machineID)

		// Update all IDs in respective MAPs
		d.insertMachine(mh)
		return mh.DeepCopy()
	}

	return nil
}

// Public API to insert an machine
func (d *Daemon) InsertMachine(mh *machine.Machine) {
	d.machinesMU.Lock()
	d.insertMachine(mh)
	d.machinesMU.Unlock()
}

// insertMachine inserts the mh in the machines map. To be used with machinesMU locked.
func (d *Daemon) insertMachine(mh *machine.Machine) {
	if mh.Status == nil {
		mh.Status = &machine.MachineStatus{}
	}

	mh.MsgRouter = message.NewMessageRouter()

	if mh.MachineID != "" {
		d.machines[mh.MachineID] = mh
	}
}

// MachineJoin sets up the machine working directory.
func (d *Daemon) MachineJoin(mh machine.Machine) error {

	if e := d.lookupMachine(mh.MachineID); e != nil {
		return fmt.Errorf("machine %s already exists", mh.MachineID)
	}

	mhDir := filepath.Join(".", mh.MachineID)
	if err := os.MkdirAll(mhDir, 0777); err != nil {
		logger.Warningf("Failed to create machine temporary directory: %s", err)
		return fmt.Errorf("failed to create temporary directory: %s", err)
	}

	d.conf.OptsMU.RLock()
	mh.SetDefaultOpts(d.conf.Opts)
	d.conf.OptsMU.RUnlock()

	d.InsertMachine(&mh)
	mh.Snapshot()

	logger.Infof("Machine (%s) Join Successful", mh.MachineID)
	return nil
}

// MachineLeave cleans the directory used by the machine and all
// relevant details.
func (d *Daemon) MachineLeave(machineID string) error {
	d.machinesMU.Lock()
	defer d.machinesMU.Unlock()

	mh := d.lookupMachine(machineID)
	if mh == nil {
		return fmt.Errorf("machine %s not found", machineID)
	}

	for _, drv := range mh.Drivers {
		err := mh.StopDriver(drv.GetConf().GetType(), drv.GetConf().GetID())
		if err != nil {
			return err
		}
		logger.Infof("Successfully removed driver %s on machine %s", drv.GetConf().GetType(), machineID)
	}

	for _, adpt := range mh.Adaptors {
		err := mh.DetachAdaptor(adpt.GetConf().GetType(), adpt.GetConf().GetID())
		if err != nil {
			return err
		}
		logger.Infof("Successfully removed adaptor %s on machine %s", adpt.GetConf().GetType(), machineID)
	}

	delete(d.machines, machineID)

	// TODO: remove the snapshot json file (or remove the directory all together?)
	d.removeSnapshotFile(mh)

	logger.Infof("Machine Leave successful on container: %s", machineID)
	return nil
}

// MachineGet returns a copy of the machine for the given machineID, or nil if the
// machine was not found.
func (d *Daemon) MachineGet(machineID string) (*machine.Machine, error) {
	d.machinesMU.RLock()
	defer d.machinesMU.RUnlock()

	if mh := d.lookupMachine(machineID); mh != nil {
		cpy := mh.DeepCopy()
		// logger.Infof("Returning mh {%+v}", cpy)
		return cpy, nil
	}
	return nil, fmt.Errorf("machine %s not found", machineID)
}

// MachineUpdate updates the given machine.
func (d *Daemon) MachineUpdate(machineID string, opts option.OptionMap) error {
	d.machinesMU.Lock()
	defer d.machinesMU.Unlock()

	mh := d.lookupMachine(machineID)
	if mh != nil {
		if err := mh.Opts.Validate(opts); err != nil {
			return err
		}

		if opts != nil && !mh.ApplyOpts(opts) {
			// No changes have been applied, skip update
			return nil
		}
	} else {
		return fmt.Errorf("machine %s not found", machineID)
	}

	mh.Snapshot()
	return nil
}

// MachinesGet returns a copy of all the machines or nil if there are no machines.
func (d *Daemon) MachinesGet() ([]machine.Machine, error) {
	d.machinesMU.RLock()
	defer d.machinesMU.RUnlock()

	mhs := []machine.Machine{}
	mhsSet := map[*machine.Machine]bool{}
	for _, v := range d.machines {
		mhsSet[v] = true
	}
	if len(mhsSet) == 0 {
		return nil, nil
	}
	for k := range mhsSet {
		mhCopy := k.DeepCopy()
		mhs = append(mhs, *mhCopy)
	}
	return mhs, nil
}
