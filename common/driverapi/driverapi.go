package driverapi

import (
	"github.com/Arvinderpal/embd-project/common/adaptorapi"
	"github.com/Arvinderpal/embd-project/common/message"
)

type DriverConf interface {
	NewDriver(adaptorapi.Adaptor, *message.Queue, *message.Queue) (Driver, error)
	ValidateConf() error
	GetType() string
	GetID() string
	GetAdaptorID() string
	GetSubscriptions() []string
}

func NewDriver(config DriverConf, apiAdpt adaptorapi.Adaptor, rcvQ *message.Queue, sndQ *message.Queue) (Driver, error) {
	return config.NewDriver(apiAdpt, rcvQ, sndQ)
}

type Driver interface {
	Start() error
	Stop() error
	GetConf() DriverConf
	Copy() Driver
	String() string
}

// DriverConfEnvelope is used primarly for easy marshalling/unmarshalling
// of various DriverConf.
type DriverConfEnvelope struct {
	Type          string      `json:"type"`
	ID            string      `json:"id"`
	AdaptorID     string      `json:"adaptor-id"`
	Qualifier     string      `json:"qualifier"`
	Subscriptions []string    `json:"subscriptions"`
	Conf          interface{} `json:"conf"`
}

// DriversConfEnvelope is used primarly for easy marshalling/unmarshalling
// of 1 or more dirvers.
type DriversConfEnvelope struct {
	MachineID string `json:"machine-id"`
	Confs     []DriverConfEnvelope
}
