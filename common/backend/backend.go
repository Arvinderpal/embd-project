package backend

import (
	"github.com/Arvinderpal/embd-project/common/types"
	"github.com/Arvinderpal/embd-project/pkg/machine"
	"github.com/Arvinderpal/embd-project/pkg/option"
)

type machineBackend interface {
	MachineJoin(mh machine.Machine) error
	MachineLeave(machineID string) error
	MachineGet(machineID string) (*machine.Machine, error)
	MachineUpdate(machineID string, opts option.OptionMap) error
	MachinesGet() ([]machine.Machine, error)
}

type driversBackend interface {
	StartDrivers(conf []byte) error
	StopDriver(machineID, driverType, driverID string) error
}

type adaptorsBackend interface {
	AttachAdaptors(conf []byte) error
	DetachAdaptor(machineID, adaptorType, adaptorID string) error
}

type control interface {
	Ping() (*types.PingResponse, error)
	Update(opts option.OptionMap) error
	GlobalStatus() (string, error)
}

// interface for both client and daemon.
type SegueBackend interface {
	machineBackend
	driversBackend
	adaptorsBackend
	control
}

// SegueDaemonBackend is the interface for daemon only.
type SegueDaemonBackend interface {
	SegueBackend
}
