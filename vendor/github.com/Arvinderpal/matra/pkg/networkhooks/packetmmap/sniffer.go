package packetmmap

import (
	"sync"

	"github.com/Arvinderpal/matra/common/programapi"
)

// Listen in an infinite loop for new packets
func (h *PacketMMAP) Listen() error {
	logger.Infof("Starting packet-mmap listener on %s", h.State.Conf.AFpacketConf.Iface)

	var (
		err    error
		wg     sync.WaitGroup
		bcChan chan programapi.MatraEvent
	)
	bcChan = make(chan programapi.MatraEvent)

	// Main Listener goroutine.
	go func() {
		// Infinite loop that reads incoming packets.
		for {
			select {
			case <-h.State.killChan:
				// This go routine terminates only if killChan is closed.
				close(bcChan)
				h.State.sniffer.Close()
				return
			default:
				var pkt programapi.MatraEvent
				pkt.Data, pkt.CI, err = h.State.sniffer.ReadPacket()
				if err != nil {
					logger.Errorf("Error getting packet (continuing): %v", err)
					// TODO: if the error is a block timeout, then we could add
					// sleep here to avoid spinning on an idle channel.
					continue
				}

				h.mu.RLock()  // hold lock while pipeline is busy.
				bcChan <- pkt // Send pkt for broadcast to all pipelines.
				// Wait for all pipelines to finish before reading next pkt.
				pipesDone := func(d <-chan programapi.MatraEvent) {
					defer wg.Done()
					for {
						select {
						case <-d:
							return
						}
					}
				}
				wg.Add(len(h.State.Pipelines))
				for _, pipe := range h.State.Pipelines {
					go pipesDone(pipe.GetDoneChan())
				}
				wg.Wait()
				h.mu.RUnlock()
			}
		}
	}()

	// Main Dispatcher goroutine.
	go func() {
		for data := range bcChan {
			h.mu.RLock()
			// Using -race will flag the above read of h.Pipelines as a race with the write to h.Pipelines in startPipeline(). At first glance, we may think that this RLock is not necessary since this read is only done when bcChan has an event (see above code) and it already has holds an RLock(). However, in the rarest of cases, it may be that this go routine stops after the call to the final ProcessData() and does not get to run again until the pipelines have finished and the RLock in the top goroutine has been release. The subsequent read of h.State.Pipelines will be racy!
			for _, pipe := range h.State.Pipelines {
				pipe.ProcessData(data)
			}
			h.mu.RUnlock()
		}
	}()

	return nil
}
