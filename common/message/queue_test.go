package message

import (
	"sync"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {

	q := New()
	// create 50 producers and 10 consumers
	var wgProd sync.WaitGroup
	producers := 50
	wgProd.Add(producers)
	for i := 0; i < producers; i++ {
		go func(i int) {
			defer wgProd.Done()
			t.Logf("Producer %v: adding: %v\n", i, i)
			m := Message{
				ID: MessageID{
					Type:    "test",
					SubType: "na",
					Version: i,
				},
			}
			q.Add(m)
		}(i)
	}

	var wgCon sync.WaitGroup
	consumers := 10
	wgCon.Add(consumers)
	for i := 0; i < consumers; i++ {
		go func(i int) {
			defer wgCon.Done()
			msg, shuttingdown := q.Get()
			if msg.ID.Type == "added after shutdown!" {
				t.Errorf("Got an msg added after shutdown.")
			}
			if shuttingdown {
				t.Logf("Shuting down...")
				return
			}
			t.Logf("Worker %v: begin processing %v\n", i, msg)
			time.Sleep(3 * time.Millisecond)
			t.Logf("Worker %v: done processing %v\n", i, msg)
			q.Done(msg)
		}(i)
	}

	wgProd.Wait()
	q.ShutDown()
	m := Message{
		ID: MessageID{
			Type:    "added after shutdown!",
			SubType: "na",
			Version: 0,
		},
	}
	q.Add(m)
	wgCon.Wait()
}

func TestAddWhileProcessing(t *testing.T) {
	q := New()

	// Start producers
	const producers = 50
	producerWG := sync.WaitGroup{}
	producerWG.Add(producers)
	for i := 0; i < producers; i++ {
		go func(i int) {
			defer producerWG.Done()
			m := Message{
				ID: MessageID{
					Type:    "test",
					SubType: "na",
					Version: i,
				},
			}
			t.Logf("Producer %v: adding: %v\n", i, m)
			q.Add(m)
		}(i)
	}

	// Start consumers
	const consumers = 10
	consumerWG := sync.WaitGroup{}
	consumerWG.Add(consumers)
	// we will add the item we are processing back to the queue
	// this should exercise the dirty map
	for i := 0; i < consumers; i++ {
		go func(i int) {
			defer consumerWG.Done()
			msg, done := q.Get()
			if done {
				return
			}
			count := map[MessageID]int{}
			if count[msg.ID] < 1 {
				q.Add(msg)
				count[msg.ID]++
			}
			t.Logf("Worker %v: begin processing %v\n", i, msg)
			// time.Sleep(3 * time.Millisecond)
			t.Logf("Worker %v: done processing %v\n", i, msg)
			q.Done(msg)

		}(i)
	}

	producerWG.Wait()
	q.ShutDown()
	consumerWG.Wait()
}
