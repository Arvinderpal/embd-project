package messagerouter

import (
	"fmt"
	"testing"

	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

// TestCentralizedMessageRouter_MasterWith2Slaves: create a master, slave-1, and slave-2. slave-1 sends a message to both master and slave-2.
func TestCentralizedMessageRouter_MasterWith2Slaves(t *testing.T) {

	// Create router on master:
	router := NewCentralizedMessageRouter()
	// Create queues used to route rf24network traffic on master
	masterRouterQueues := router.AddNode("master", 1)
	slave1RouterQueues := router.AddNode("slave-1", 1)
	slave2RouterQueues := router.AddNode("slave-2", 1)

	// master subscribes to MessageType_UnitTest
	entry := RouteEntry{
		MsgType: seguepb.MessageType_UnitTest,
		NodeID:  "master",
	}
	router.AddRoute(entry)

	// slave-2 subscribes to MessageType_UnitTest
	entry = RouteEntry{
		MsgType: seguepb.MessageType_UnitTest,
		NodeID:  "slave-2",
	}
	router.AddRoute(entry)

	// slave-1 sends a MessageType_UnitTest
	msg := message.Message{
		ID: seguepb.Message_MessageID{
			Type:    seguepb.MessageType_UnitTest,
			SubType: "n/a",
			Version: uint64(1),
		},
		Data: &seguepb.UnitTestData{TestMessage: fmt.Sprintf("msg: 1")},
	}
	slave1RouterQueues.RcvQ.Add(msg)

	// see the message shows up at slave-2

	msgRvced, _ := slave2RouterQueues.SndQ.Get()
	if msgRvced.ID.Type != msg.ID.Type {
		t.Errorf("expected message %+q got %+q")
	}
	slave2RouterQueues.SndQ.Done(msgRvced)

	// see the message shows up at master
	msgRvced, _ = masterRouterQueues.SndQ.Get()
	if msgRvced.ID.Type != msg.ID.Type {
		t.Errorf("expected message %+q got %+q")
	}
	masterRouterQueues.SndQ.Done(msgRvced)

	router.Stop()
}
