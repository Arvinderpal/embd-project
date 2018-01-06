package daemon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/driverapi"
	"github.com/Arvinderpal/embd-project/pkg/machine"
)

// removeSnapshotFile will remove the json state file from mh's directory.
// Typically called when an mh is remove from segue.
func (d *Daemon) removeSnapshotFile(mh *machine.Machine) error {
	snapFile := path.Join(".", mh.MachineID, common.SnapshotFileName)
	err := os.Remove(snapFile)
	return err
}

// restore will go through all existing mh directories and restore mhs that have a json file present.
func (d *Daemon) restore() error {

	restored := 0
	errored := 0

	// Grab a list of all the existing mh directories
	files, err := ioutil.ReadDir(".")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Loop over each of the mh directories and see if there is an mh json
	// to restore there
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		stateFile := filepath.Join(".", file.Name(), common.SnapshotFileName)
		if _, err := os.Stat(stateFile); os.IsNotExist(err) {
			continue
		}
		logger.Infof("Restoring machine: %s", file.Name())
		mh, err := d.restoreMachine(stateFile)
		if err != nil {
			logger.Errorf("Failed to restore mh %s: %v", file.Name(), err)
			if mh != nil {
				mh.LogStatus(machine.Failure, fmt.Sprintf("Restore failed: %s", err))
			}
			errored++
			continue
		}
		restored++
	}
	logger.Infof("Restored %d machine and %d errored.", restored, errored)
	return nil
}

// restoreMachine will create an machine object from the state file and start it..
func (d *Daemon) restoreMachine(stateFile string) (*machine.Machine, error) {

	// open the state file
	f, err := os.Open(stateFile)
	if err != nil {
		logger.Warningf("Error opening state file %v", err)
		return nil, err
	}
	defer f.Close()

	// attempt to unmarshal the mh state
	var tMh *machine.Machine
	err = json.NewDecoder(f).Decode(&tMh)
	if err != nil {
		logger.Warningf("Error parsing state file: %v", err)
		return nil, err
	}

	// The mh object created from the JSON file provides us with the list of drivers that were running when the snapshot was taken. We will create a new mh object to which we will attach and start those drivers.

	mh := tMh.DeepCopy()
	mh.Drivers = nil // These will be populated as we attach drivers below
	d.InsertMachine(mh)

	mh.LogStatus(machine.Info, "Restoring machine...")

	// attach adaptors
	eAdpts := adaptorapi.AdaptorsConfEnvelope{
		MachineID: tMh.MachineID,
	}
	for _, adpt := range tMh.Adaptors {
		eAdpt := adaptorapi.AdaptorConfEnvelope{
			Type: adpt.GetConf().GetType(),
			Conf: adpt.GetConf(),
		}
		eAdpts.Confs = append(eAdpts.Confs, eAdpt)
	}
	confb, err := json.Marshal(eAdpts)
	if err != nil {
		return mh, err
	}
	err = d.AttachAdaptors(confb)
	if err != nil {
		return mh, err
	}

	// start drivers
	eDrvs := driverapi.DriversConfEnvelope{
		MachineID: tMh.MachineID,
	}
	for _, drv := range tMh.Drivers {
		eDrv := driverapi.DriverConfEnvelope{
			Type: drv.GetConf().GetType(),
			Conf: drv.GetConf(),
		}
		eDrvs.Confs = append(eDrvs.Confs, eDrv)
	}
	confb, err = json.Marshal(eDrvs)
	if err != nil {
		return mh, err
	}
	err = d.StartDrivers(confb)
	if err != nil {
		return mh, err
	}

	err = mh.Snapshot() // Snapshot the new mh object.
	return mh, err

}
