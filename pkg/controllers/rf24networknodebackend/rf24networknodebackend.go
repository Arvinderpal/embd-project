package rf24networknodebackend

import (
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("rf24networknode-backend")
)

const Master_Node_Address = 0x00
const rf24NetworkReadBufferrSize = 1024
const frameChanCapacity = 10

type RF24NetworkNodeBackend interface {
	Run() error
	Stop() error
}
