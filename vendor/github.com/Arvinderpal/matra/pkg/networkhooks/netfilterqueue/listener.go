package netfilterqueue

import (
	"sync"

	"github.com/Arvinderpal/matra/common/programapi"
)

// Listen in an infinite loop for new packets
func (h *NetfilterQueue) Listen() error {
	logger.Infof("Starting netfilter-queue listener on iface %s with queue id %d", h.State.Conf.NFQConf.Iface, h.State.Conf.NFQConf.QueueId)

	var (
		wg     sync.WaitGroup
		bcChan chan programapi.MatraEvent
	)
	bcChan = make(chan programapi.MatraEvent)
	packetsChan := h.State.nfqueue.GetPackets()

	// Main Listener goroutine.
	go func() {
		// Infinite loop that reads incoming packets.
		for {
			select {
			case <-h.State.killChan:
				// This go routine terminates only if killChan is closed.
				close(bcChan)
				h.State.nfqueue.Close()
				return
			case p := <-packetsChan:
				var pkt programapi.MatraEvent
				pkt.Data = p.Packet.Data()
				if p.Packet.Metadata() != nil {
					pkt.CI = p.Packet.Metadata().CaptureInfo
				}
				pkt.VerdictChannel = p.VerdictChannel

				h.RLock()     // hold lock while pipeline is busy.
				bcChan <- pkt // Send pkt for broadcast to all pipelines.
				// Wait for all pipelines to finish before reading next pkt.
				pipesDone := func(d <-chan programapi.MatraEvent) {
					defer wg.Done()
					for {
						select {
						case <-d:
							// logger.Infof("Verdict: %d", result.Verdict)
							// p.SetVerdict(Verdict(result.Verdict))
							return
						}
					}
				}
				wg.Add(len(h.State.Pipelines))
				for _, pipe := range h.State.Pipelines {
					go pipesDone(pipe.GetDoneChan())
				}
				wg.Wait()
				h.RUnlock()
			}
		}
	}()

	// Main Dispatcher goroutine.
	go func() {
		for data := range bcChan {
			h.RLock()
			for _, pipe := range h.State.Pipelines {
				pipe.ProcessData(data)
				// pipe.InChan <- data
			}
			h.RUnlock()
		}
	}()

	return nil
}
