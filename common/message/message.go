package message

type MessageID struct {
	Type    string
	SubType string
	Version int // Version can be used serialize messages of the same type+subtype in a queue. If two messages have the same version, then any existing message in the queue will be overwritten.
}

// Message is the basic unit of data passed between modules.
// A message is uniquely identified by the tuple = {Type, SubType, Version}
type Message struct {
	ID   MessageID
	Data []byte
}
