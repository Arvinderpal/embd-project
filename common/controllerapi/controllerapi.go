package controllerapi

import (
	"github.com/Arvinderpal/embd-project/common/message"
)

type ControllerConf interface {
	NewController(*message.Queue, *message.Queue) (Controller, error)
	ValidateConf() error
	GetType() string
	GetID() string
	GetSubscriptions() []string
}

func NewController(config ControllerConf, rcvQ *message.Queue, sndQ *message.Queue) (Controller, error) {
	return config.NewController(rcvQ, sndQ)
}

type Controller interface {
	Start() error
	Stop() error
	GetConf() ControllerConf
	Copy() Controller
	String() string
}

// ControllerConfEnvelope is used primarly for easy marshalling/unmarshalling
// of various ControllerConf.
type ControllerConfEnvelope struct {
	Type          string      `json:"type"`
	ID            string      `json:"id"`
	Subscriptions []string    `json:"subscriptions"`
	Conf          interface{} `json:"conf"`
}

// ControllersConfEnvelope is used primarly for easy marshalling/unmarshalling
// of 1 or more dirvers.
type ControllersConfEnvelope struct {
	MachineID string `json:"machine-id"`
	Confs     []ControllerConfEnvelope
}

type GRPCConf struct {
	HostAddress string `json:"host-address"` // IP address or just localhost
	Port        int    `json:"port"`         // Port to listen on
	TLSEnabled  bool   `json:"tls-enabled"`  // Will enable TLS on server
	CertFile    string `json:"cert-file"`    // TLS certfile (optional)
	KeyFile     string `json:"key-file"`     // TLS Key (optional)
}

// MPIMetaConf is passed to mpi-plugins
type MPIMetaConf struct {
	MachineID     string   `json:"machine-id"`
	ID            string   `json:"id"`            // IMPORTANT: ID should be unique for across all nodes in the RF24Network
	Subscriptions []string `json:"subscriptions"` // Message Type Subscriptions.
}
