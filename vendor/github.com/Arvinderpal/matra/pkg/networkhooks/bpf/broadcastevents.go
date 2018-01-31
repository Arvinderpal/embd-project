package bpf

import (
	"github.com/Arvinderpal/matra/common/programapi"
)

// Broadcaster will read events from the internal event chan and broadcast them to all pipelines
func (h *PipelineHandle) Broadcaster(killChan chan struct{}) error {
	logger.Infof("Starting bpf hook broadcast routine")

	bcChan := make(chan programapi.MatraEvent)

	// Main goroutine.
	go func() {
		// Infinite loop that reads incoming events.
		for {
			select {
			case <-killChan:
				logger.Infof("stopping bpf hook broadcast routine (hook kill chan asserted)")
				close(bcChan)
				return
			case <-h.pipeKillChan:
				logger.Infof("stopping bpf hook broadcast routine (pipeline kill chan asserted)")
				close(bcChan)
				return
			case data := <-h.internalEventChan:
				var evt programapi.MatraEvent
				evt.Data = data

				h.mu.RLock() // hold lock while pipeline is busy.
				h.Pipeline.ProcessData(evt)
				// Wait for pipeline to finish before reading next event.
				<-h.Pipeline.GetDoneChan()
				h.mu.RUnlock()
			}
		}
	}()
	return nil
}
