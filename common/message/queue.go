// Queueprovides a simple queue that supports the following
// features:
//  * Fair: items processed in the order in which they are added.
//  * Stingy: a single item will not be processed multiple times concurrently,
//      and if an item is added multiple times before it can be processed, it
//      will only be processed once.
//  * Multiple consumers and producers. In particular, it is allowed for an
//      item to be reenqueued while it is being processed.
//  * Shutdown notifications.
package message

import "sync"

type Interface interface {
	Add(item Message)
	Len() int
	Get() (item Message, shutdown bool)
	Done(item Message)
	ShutDown()
	IsShuttingDown() bool
}

type set map[MessageID]struct{}

type Queue struct {
	queue        []Message
	dirty        set
	processing   set
	condition    *sync.Cond
	shuttingdown bool
}

func New() *Queue {
	return &Queue{
		dirty:      set{},
		processing: set{},
		condition:  sync.NewCond(&sync.Mutex{}),
	}
}

func (q *Queue) Add(msg Message) {

	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	if q.shuttingdown {
		return
	}
	if _, exists := q.dirty[msg.ID]; exists {
		return
	}
	q.dirty[msg.ID] = struct{}{}
	if _, exists := q.processing[msg.ID]; exists {
		// we'll add it from dirty when Done() is called
		return
	}
	q.queue = append(q.queue, msg)
	q.condition.Signal()
}

func (q *Queue) Get() (Message, bool) {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	for len(q.queue) == 0 && !q.shuttingdown {
		// we wait if queue is empty
		// if we're shuttingdown, then we stop waiting and let the queue
		// be drained (we don't allow new items to be added)
		q.condition.Wait()
	}
	if len(q.queue) == 0 {
		// we must be shuttingdown, if queue is empty, notify workers
		return Message{}, true
	}
	msg := q.queue[0]
	q.queue = q.queue[1:]
	q.processing[msg.ID] = struct{}{} // add to processing
	delete(q.dirty, msg.ID)           // remove from dirty
	return msg, false
}

func (q *Queue) Done(msg Message) {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	delete(q.processing, msg.ID)
	if _, exists := q.dirty[msg.ID]; !exists {
		return
	}
	delete(q.dirty, msg.ID)
	q.queue = append(q.queue, msg)
	q.condition.Signal()
}

func (q *Queue) ShutDown() {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()

	q.shuttingdown = true
	q.condition.Broadcast()
}

func (q *Queue) IsShuttingDown() bool {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	return q.shuttingdown
}

func (q *Queue) Len() int {
	q.condition.L.Lock()
	defer q.condition.L.Unlock()
	return len(q.queue)
}
