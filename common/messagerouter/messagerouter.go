package messagerouter

import (
	"fmt"
	"sync"

	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
	logging "github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("messagerouter")
)

type MessageRouter interface {
	AddRoute(entry RouteEntry) error
	AddNode(id string, workerCount int) *NodeQueues
	RemoveNode(id string) bool
	GetNode(id string) *NodeQueues
	Stop()
}

// RouteEntry defines a route entry in the router.
type RouteEntry struct {
	MsgType seguepb.MessageType
	NodeID  string
}

// We define a Node  object that consists of a receive and send queue; a
// router may have several Nodes and is responsible for routing messages to
// nodes that subscribe to the specific message types.
type NodeQueues struct {
	ID   string
	RcvQ *message.Queue
	SndQ *message.Queue
}

// CentralizedMessageRouter implements centralized routing of messages
// at master.
// Nodes/Children send hearbeats to Master; the heartbeats include message
// subscription information. Master keeps track of which node has subscribed
// to which message type and routes those messages accordingly.
type CentralizedMessageRouter struct {
	mu       sync.RWMutex
	routes   map[seguepb.MessageType]map[string]*NodeQueues // Maps message types to set of nodes that subscribe to this type.
	nodes    map[string]*NodeQueues                         // set of known nodes
	killChan chan struct{}
}

func NewCentralizedMessageRouter() *CentralizedMessageRouter {

	router := &CentralizedMessageRouter{
		routes:   make(map[seguepb.MessageType]map[string]*NodeQueues),
		nodes:    make(map[string]*NodeQueues),
		killChan: make(chan struct{}),
	}

	return router
}

// Stop: shutsdown all router routines
func (r *CentralizedMessageRouter) Stop() {
	for _, n := range r.nodes {
		r.RemoveNode(n.ID)
	}
}

// AddRoute: given a message type and id of the node interested in receiving
// messages of this type, it will add a route entry internally.
func (r *CentralizedMessageRouter) AddRoute(entry RouteEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, present := r.nodes[entry.NodeID]
	if !present {
		return fmt.Errorf("AddRoute: no entry for node %s found", entry.NodeID)
	}
	nodeSet := r.routes[entry.MsgType]
	if nodeSet == nil {
		nodeSet = make(map[string]*NodeQueues)
		nodeSet[entry.NodeID] = node
		r.routes[entry.MsgType] = nodeSet
	} else {
		nodeSet[entry.NodeID] = node
	}
	return nil
}

// AddNode: will create a node entry in the router. workerCount specifies
// the number of go routines to run for reading message off the node's receive
// queue. If no receive queue is defined, no workers are created.
// NOTE: for rf24network, we don't have a receive queues for each node, instead
// we have a single receive queue from which traffic for all nodes is received;
// in that particular case, it would  make sense to have a higher number of
// workers defined for the singel receive queue. For GRPC, where we have a
// per node/client recieve queue, we can use 1 worker per receive queue.
func (r *CentralizedMessageRouter) AddNode(id string, workerCount int) *NodeQueues {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, present := r.nodes[id]
	if present {
		return node
	}
	node = &NodeQueues{
		ID:   id,
		RcvQ: message.NewQueue(id),
		SndQ: message.NewQueue(id),
	}
	r.nodes[id] = node

	for i := 0; i < workerCount; i++ {
		go r.worker(node)
	}
	return node
}

// RemoveNode: will remove a node entry from the router if it exists
func (r *CentralizedMessageRouter) RemoveNode(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, present := r.nodes[id]
	if !present {
		return false
	}

	removeRoutes := func(id string) {
		// remove all routes to node
		for _, nodeSet := range r.routes {
			_, present := nodeSet[id]
			if present {
				delete(nodeSet, id)
			}
		}
	}

	removeRoutes(id)
	node.RcvQ.ShutDown()
	node.SndQ.ShutDown()

	delete(r.nodes, id)
	return true
}

func (r *CentralizedMessageRouter) GetNode(id string) *NodeQueues {
	node, present := r.nodes[id]
	if present {
		return node
	}
	return nil
}

// worker: will round-robin through all the node receive queues, checking if
// their is a message present, if so, routing the message to the appropriate
// send queues of the nodes.
func (r *CentralizedMessageRouter) worker(node *NodeQueues) {
	logger.Debugf("starting worker for node: %v\n", node.ID)
	defer logger.Debugf("stopping worker\n")
	for {
		// read from queue and attempt to send
		iMsg, shutdown := node.RcvQ.Get()
		if shutdown {
			logger.Debugf("queue shutdown called on node: %s\n", node.ID)
			return
		}
		// lookup all nodes interested in receiving this message and put it in theire respective queues
		r.mu.RLock()
		nodeSet, present := r.routes[iMsg.ID.Type]
		if present {
			for _, n := range nodeSet {
				n.SndQ.Add(iMsg)
			}
		}
		r.mu.RUnlock()
		node.RcvQ.Done(iMsg)
	}
}
