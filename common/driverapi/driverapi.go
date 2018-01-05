package driverapi

type DriverConf interface {
	NewDriver() (Driver, error)
	ValidateConf() error
	GetType() string
	GetID() string
}

func NewDriver(config DriverConf) (Driver, error) {
	return config.NewDriver()
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
	Type string      `json:"type"`
	ID   string      `json:"id"`
	Conf interface{} `json:"conf"`
}

// DriversConfEnvelope is used primarly for easy marshalling/unmarshalling
// of 1 or more dirvers.
type DriversConfEnvelope struct {
	MachineID string `json:"machine-id"`
	Confs     []DriverConfEnvelope
}
