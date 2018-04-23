package rf24networknodebackend

import "github.com/Arvinderpal/embd-project/common/message"

const Master_Node_Address = 0x00

type RF24NetworkNodeBackend interface {
	Run() error
	Send(message.Message) error
	Stop() error
}
