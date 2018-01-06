package adaptorapi

import (
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

type AdaptorConf interface {
	NewAdaptor() (Adaptor, error)
	ValidateConf() error
	GetType() string
	GetID() string
}

func NewAdaptor(config AdaptorConf) (Adaptor, error) {
	return config.NewAdaptor()
}

type Adaptor interface {
	Attach() error
	Detach() error
	GetGobotAdaptor() gobot.Adaptor
	GetDigitalWriter() (gpio.DigitalWriter, error)
	GetConf() AdaptorConf
	Copy() Adaptor
	String() string
}

// AdaptorConfEnvelope is used primarly for easy marshalling/unmarshalling
// of various AdaptorConf.
type AdaptorConfEnvelope struct {
	Type string      `json:"type"`
	ID   string      `json:"id"`
	Conf interface{} `json:"conf"`
}

// AdaptorsConfEnvelope is used primarly for easy marshalling/unmarshalling
// of 1 or more dirvers.
type AdaptorsConfEnvelope struct {
	MachineID string `json:"machine-id"`
	Confs     []AdaptorConfEnvelope
}
