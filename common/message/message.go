package message

import (
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/gogo/protobuf/proto"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("segue-message")
)

// Important: this func must be updated manually evertime we introduce a new message type.
// The converter should take in an Message in pb format (Data is []byte)
// and unmarshal it into the corresponding internal data format.
func ConvertToInternalFormat(m *seguepb.Message) (Message, error) {

	switch m.GetID().GetType() {
	case seguepb.MessageType_SensorUltraSonic:
		data := &seguepb.SensorUltraSonicData{}
		err := proto.Unmarshal(m.GetData(), data)
		if err != nil {
			return Message{}, err
		}
		return Message{
			ID:   *m.ID,
			Data: *data,
		}, nil
	case seguepb.MessageType_CmdDrive:
		data := &seguepb.CmdDriveData{}
		err := proto.Unmarshal(m.GetData(), data)
		if err != nil {
			return Message{}, err
		}
		return Message{
			ID:   *m.ID,
			Data: *data,
		}, nil
	default:
		return Message{}, fmt.Errorf("converter error: unknown message type %s", m.GetID().GetType())
	}

}

// Message is the basic unit of data passed between modules.
// A message is uniquely identified by the tuple = {Type, SubType, Version}
// NOTE: External messages (from sources outside of segue instance) must use
// protobuf Message definition.
type Message struct {
	ID   seguepb.Message_MessageID
	Data interface{} // []byte
}

type MessageRouter struct {
	mu       sync.RWMutex
	RouteMap map[seguepb.MessageType][]*Queue
}

func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		RouteMap: make(map[seguepb.MessageType][]*Queue),
	}
}

// AddListener: adds a listner on queue that will be used by the drivers/controllers to send messages (i.e. SndQ)
func (m *MessageRouter) AddListener(queue *Queue) {

	go func() {
		for {
			msg, shutdown := queue.Get()
			if shutdown {
				logger.Debugf("stopping listener on queue %s", queue.ID())
				return
			}
			// Forward message to all subscriber queues
			m.mu.RLock()
			set := m.RouteMap[msg.ID.Type]
			for _, q := range set {
				q.Add(msg)
			}
			queue.Done(msg)
			m.mu.RUnlock()
		}
	}()
	return
}

// AddSubscriberQueue: adds the subscriber queue for the message type specified
// This is typically the RcvQ for drivers/controllers
func (m *MessageRouter) AddSubscriberQueue(msgType string, queue *Queue) error {
	// We have to lookup the corresponding enum value based on the name
	enumVal, ok := seguepb.MessageType_value[msgType]
	if !ok {
		return fmt.Errorf("message type %s not found in internal enum map", msgType)
	}

	set := m.RouteMap[seguepb.MessageType(enumVal)]
	if queue.IsShuttingDown() {
		return fmt.Errorf("cannot add subscriber (%s) on message type %s, queue must not be shutting down", queue.ID(), msgType)
	}
	if queue.Len() != 0 {
		return fmt.Errorf("cannot add subscriber (%s) on message type %s, queue must be empty: current length: %d", queue.ID(), msgType, queue.Len())
	}
	for _, q := range set {
		if queue.ID() == q.ID() {
			return nil
		}
	}
	set = append(set, queue)
	m.RouteMap[seguepb.MessageType(enumVal)] = set

	return nil
}

// RemoveSubscriberQueue: removes the subscriber queue for message type specified
// IMPORTANT: user must ensure that the queue has been shutdown and all the messages drained before calling this func.
func (m *MessageRouter) RemoveSubscriberQueue(msgType, subID string) error {
	enumVal, ok := seguepb.MessageType_value[msgType]
	if !ok {
		return fmt.Errorf("message type %s not found in internal enum map", msgType)
	}
	// enumVal = seguepb.MessageType(enumVal)
	set := m.RouteMap[seguepb.MessageType(enumVal)]
	for i, q := range set {
		if subID == q.ID() {
			if !q.IsShuttingDown() {
				return fmt.Errorf("cannot remove subscriber (%s) on message type %s: queue must first be shutdown", subID, msgType)
			}
			set = append(set[:i], set[i+1:]...)
			m.RouteMap[seguepb.MessageType(enumVal)] = set
			// TODO: if map entry is empty for a particular message type, then we can delete that entry entirely.
			return nil
		}
	}
	return nil
}

// GetSubscribers: returns the set of subscriber queues for message type specified.
func (m *MessageRouter) GetSubscribers(msgType string) []*Queue {
	enumVal, ok := seguepb.MessageType_value[msgType]
	if !ok {
		return nil
	}
	if queues, found := m.RouteMap[seguepb.MessageType(enumVal)]; found {
		return queues
	}
	return nil
}
