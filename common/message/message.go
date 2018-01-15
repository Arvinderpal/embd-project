package message

import (
	"fmt"
	"sync"

	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("segue-message")
)

type MessageID struct {
	Type    string
	SubType string
	Version int // Version can be used serialize messages of the same type+subtype in a queue. If two messages have the same version, then any existing message in the queue will be overwritten.
}

// Message is the basic unit of data passed between modules.
// A message is uniquely identified by the tuple = {Type, SubType, Version}
type Message struct {
	ID   MessageID
	Data interface{} // []byte
}

type MessageRouter struct {
	mu       sync.RWMutex
	RouteMap map[string][]*Queue
}

func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		RouteMap: make(map[string][]*Queue),
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
			m.mu.RUnlock()
		}
	}()
	return
}

// AddSubscriberQueue: adds the subscriber queue for the message type specified
// This is typically the RcvQ for drivers/controllers
func (m *MessageRouter) AddSubscriberQueue(msgTye string, queue *Queue) error {
	set := m.RouteMap[msgTye]
	if queue.IsShuttingDown() {
		return fmt.Errorf("cannot add subscriber (%s) on message type %s, queue must not be shutting down", queue.ID(), msgTye)
	}
	if queue.Len() != 0 {
		return fmt.Errorf("cannot add subscriber (%s) on message type %s, queue must be empty: current length: %d", queue.ID(), msgTye, queue.Len())
	}
	for _, q := range set {
		if queue.ID() == q.ID() {

			return nil
		}
	}
	set = append(set, queue)
	m.RouteMap[msgTye] = set

	return nil
}

// RemoveSubscriberQueue: removes the subscriber queue for message type specified
// IMPORTANT: user must ensure that the queue has been shutdown and all the messages drained before calling this func.
func (m *MessageRouter) RemoveSubscriberQueue(msgTye, subID string) error {
	set := m.RouteMap[msgTye]
	for i, q := range set {
		if subID == q.ID() {
			if !q.IsShuttingDown() {
				return fmt.Errorf("cannot remove subscriber (%s) on message type %s: queue must first be shutdown", subID, msgTye)
			}
			set = append(set[:i], set[i+1:]...)
			m.RouteMap[msgTye] = set
			// TODO: if map entry is empty for a particular message type, then we can delete that entry entirely.
			return nil
		}
	}
	return nil
}

// GetSubscribers: returns the set of subscriber queues for message type specified.
func (m *MessageRouter) GetSubscribers(msgType string) []*Queue {
	if queues, found := m.RouteMap[msgType]; found {
		return queues
	}
	return nil
}
