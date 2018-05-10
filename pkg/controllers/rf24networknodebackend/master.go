package rf24networknodebackend

import (
	"sync"
	"time"

	"github.com/Arvinderpal/RF24Network"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/messagerouter"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

var rfhook *RF24NetworkHook

// Implements RF24NetworkNodeBackend interface
type RF24NetworkNodeMaster struct {
	mu            sync.RWMutex
	id            string
	address       uint16
	subscriptions []int32
	killChan      chan struct{}
	pollInterval  time.Duration

	controllerSndQ *message.Queue // messages received on rf24network are placed here for consumption by drivers/controllers on this node.
	controllerRcvQ *message.Queue // messages are read from this queue and sent out over the rf24network

	router       messagerouter.MessageRouter
	routerQueues *messagerouter.NodeQueues // This set of send and receive queues are used by the router.

	nodes map[string]*childNode // set of all children nodes

}

func NewRF24NetworkNodeMaster(id string, address uint16, subs []int32, n RF24Network.RF24Network, pollInterval int, sndQ, rcvQ *message.Queue, routerWorkers int) *RF24NetworkNodeMaster {

	master := &RF24NetworkNodeMaster{
		id:             id,
		address:        address,
		subscriptions:  subs,
		killChan:       make(chan struct{}),
		pollInterval:   time.Duration(pollInterval),
		controllerSndQ: sndQ,
		controllerRcvQ: rcvQ,
		nodes:          make(map[string]*childNode),
	}

	rfhook = NewRF24NetworkHook(n, master.killChan)

	// Create router on master:
	master.router = messagerouter.NewCentralizedMessageRouter()
	// Create queues used to route rf24network traffic on master
	master.routerQueues = master.router.AddNode(id, routerWorkers)

	return master
}

func (r *RF24NetworkNodeMaster) Stop() error {
	close(r.killChan)
	r.router.Stop()
	for _, n := range r.nodes {
		n.stop()
	}
	return nil
}

func (r *RF24NetworkNodeMaster) Run() error {

	for i := 0; i < 2; i++ {
		// processFrames takes whole frames of the wire and converts them to
		// messages; this func should run frequently so as to keep the router
		// funcs below properly supplied with new messages.
		go rfhook.processFrames()

		// Sender is func that demuxes messages to the queues of individual
		// nodes; how frequently this func runs is important for overall throughput.
		go r.sender()

		// messageHandler demuxes messages received over rf to the queues of
		// nodes that have subscribed; how frequently we run this func
		// impacts routing performance.
		go r.messageHandler()

		// removeStaleRoutes will remove stale nodes and their route entries in the router.
		go r.removeStaleRoutes()
	}

	tickChan := time.NewTicker(r.pollInterval * time.Millisecond).C
	// Main listener loop for RF24Network:
	for {
		select {
		case <-r.killChan:
			return nil
		case <-tickChan:
			rfhook.Receive()
		}
	}
	return nil
}

// TODO: What if the queue gets too long? We should probably drop stale messages that have been sitting too long.
func (r *RF24NetworkNodeMaster) sender() {
	for {
		select {
		case <-r.killChan:
			return
		default:
			iMsg, shutdown := r.controllerRcvQ.Get()
			if shutdown {
				logger.Debugf("stopping sender\n")
				return
			}
			r.routerQueues.RcvQ.Add(iMsg) // send to sole recieve queue on router
			r.controllerRcvQ.Done(iMsg)
		}
	}
}

// messageHandler: messages received from rf24network
func (r *RF24NetworkNodeMaster) messageHandler() {
	for {
		select {
		case <-r.killChan:
			return
		case iMsg := <-rfhook.messageChan:
			logger.Debugf("Got Message: %v\n", iMsg)
			if iMsg.ID.Type == seguepb.MessageType_RF24NetworkNodeHeartbeat {
				r.heartbeatHandler(iMsg)
			} else {
				r.controllerSndQ.Add(iMsg)
				r.routerQueues.RcvQ.Add(iMsg) // send to sole recieve queue on router which will route message to individual nodes.
			}
		}
	}
}

func (r *RF24NetworkNodeMaster) heartbeatHandler(msg message.Message) {

	r.mu.Lock()
	defer r.mu.Unlock()

	data := msg.Data.(*seguepb.RF24NetworkNodeHeartbeatData)
	logger.Debugf("Got hearbeat: %v\n", msg)

	node, present := r.nodes[data.Id]
	if !present {
		child := &childNode{
			id:                data.Id,
			address:           uint16(data.Address),
			routerQueues:      r.router.AddNode(data.Id, 0),
			lastTimestamp:     time.Now(),
			heartbeatInterval: time.Duration(data.Heartbeatinterval),
			paused:            false,
			condition:         sync.NewCond(&sync.Mutex{}),
		}
		r.nodes[data.Id] = child
		// TODO: if subscriptions have changed, we should remove old subs and add the new ones. This requires us to do a diff. Not sure we'll require this dynamic handling of subscriptions...
		for _, sub := range data.Subscriptions {
			entry := messagerouter.RouteEntry{
				MsgType: seguepb.MessageType(sub),
				NodeID:  data.Id,
			}
			r.router.AddRoute(entry)
		}

		logger.Debugf("New node discovered: %v\n", child)
		go child.worker()

	} else {
		node.lastTimestamp = time.Now()
		node.unpause()
		logger.Debugf("Node timestamp updated: %v\n", node.id, node.lastTimestamp)
	}
}

const staleRoutineRunInterval = 5 // run below routine every x seconds
const intervalsToWait = 3

// removeStaleRoutes removes nodes form whom we have not receive a heartbeat over several intervals.
func (r *RF24NetworkNodeMaster) removeStaleRoutes() {

	tickChan := time.NewTicker(staleRoutineRunInterval * time.Second).C
	for {
		select {
		case <-r.killChan:
			return
		case <-tickChan:
			cleanupNode := func(id string) {
				r.router.RemoveNode(id)
				node := r.nodes[id]
				node.stop()
				delete(r.nodes, id)
			}
			findStaleNodes := func() []string {
				var ids []string
				currentTime := time.Now()
				for id, node := range r.nodes {
					nodeTS := node.lastTimestamp.Add(intervalsToWait * time.Second * node.heartbeatInterval) // no hb in the last 3 intervals
					if nodeTS.Before(currentTime) {
						// expired node
						logger.Debugf("rf24NetworkRouter: removing node - no heartbeat recevied from node %v (%s > %s)", node.id, nodeTS, currentTime)
						ids = append(ids, id)
					}
				}
				return ids
			}

			r.mu.Lock()
			staleIDs := findStaleNodes()
			for _, sid := range staleIDs {
				cleanupNode(sid)
			}
			r.mu.Unlock()
		}
	}
}

type childNode struct {
	id                string
	address           uint16
	routerQueues      *messagerouter.NodeQueues // Send and Recieve queues to the message router for this node.
	lastTimestamp     time.Time
	heartbeatInterval time.Duration

	paused    bool       // If paused, we will not send on rf network.
	condition *sync.Cond // Used in conjunction with paused to pause/unpause worker routine.

}

// worker: this is a per node routine. it read messages from outboud queues of the node and sends it out over the rf24network
func (c *childNode) worker() {
	logger.Debugf("starting worker for node: %v\n", c)
	defer logger.Debugf("stopping worker\n")
	for {
		c.condition.L.Lock()
		for c.paused && !c.routerQueues.SndQ.IsShuttingDown() {
			// We wait if node is paused and we're not shuting down.
			c.condition.Wait()
		}
		c.condition.L.Unlock()

		// NOTE: We need this check in addition to the shutdown check
		// in the Get() method below. Get() will only show shutdown when
		// the queue is empty; however, in our case, we want to exist
		// immediately and not wait for queue to be drained.
		if c.routerQueues.SndQ.IsShuttingDown() {
			logger.Debugf("queue shutdown called on node: %s\n", c.id)
			return
		}

		// read from queue and attempt to send
		iMsg, shutdown := c.routerQueues.SndQ.Get()
		if shutdown {
			logger.Debugf("queue shutdown called on node: %s\n", c.id)
			return
		}

		logger.Debugf("sender (%s): sending message %v\n", c.id, iMsg)
		err := rfhook.rf24NetworkSend(iMsg, c.address)
		if err != nil {
			// TODO: we should filter for rf related errors and only pause if there is an rf error.
			logger.Errorf("error in rf24 send routine: %v", err)
			c.pause()
		}
		c.routerQueues.SndQ.Done(iMsg)
	}
}

func (c *childNode) pause() {
	c.condition.L.Lock()
	defer c.condition.L.Unlock()

	c.paused = true
}

func (c *childNode) unpause() {
	c.condition.L.Lock()
	defer c.condition.L.Unlock()

	c.paused = false
	c.condition.Broadcast()
}

func (c *childNode) stop() {

	// c.routerQueues.SndQ.ShutDown() // RemoveNode() should do the shutdown
	c.unpause() // if the worker is paused, we need to unpause it so that the routine can exit cleanly.
}
