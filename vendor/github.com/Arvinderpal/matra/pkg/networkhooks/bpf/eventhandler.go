package bpf

/*
#cgo CFLAGS: -I../../bpf/include
#include <linux/perf_event.h>
*/
import "C"

import (
	"syscall"
	"time"

	"github.com/Arvinderpal/matra/common"
	"github.com/Arvinderpal/matra/pkg/bpf"
)

func NewPerfConfig(path string, cpus, pages int) *bpf.PerfEventConfig {
	return &bpf.PerfEventConfig{
		MapPath:      path,
		Type:         C.PERF_TYPE_SOFTWARE,
		Config:       C.PERF_COUNT_SW_BPF_OUTPUT,
		SampleType:   C.PERF_SAMPLE_RAW,
		WakeupEvents: 1,
		NumPages:     pages,
		NumCpus:      cpus,
	}
}

func (h *PipelineHandle) lostEvent(lost *bpf.PerfEventLost, cpu int) {
	logger.Infof("Lost %d events\n", lost.Lost)
}

func (h *PipelineHandle) receiveEvent(msg *bpf.PerfEventSample, cpu int) {
	// logger.Debugf("sending event on event chan...")
	// NOTE: DataDirect() would create memory leaks since the Read() func calling this func will free the ring buffer as soon as this func returns.
	data := msg.DataCopy()
	h.internalEventChan <- data

	// prefix := fmt.Sprintf("CPU %02d:", cpu)
	// if data[0] == bpf.MATRA_NOTIFY_DBG_MSG {
	// 	dm := bpf.DebugMsg{}
	// 	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &dm); err != nil {
	// 		logger.Warningf("Error while parsing debug message: %s\n", err)
	// 	} else {
	// 		dm.Dump(data, prefix)
	// 	}
	// } else if data[0] == bpf.MATRA_NOTIFY_DBG_CAPTURE {
	// 	dc := bpf.DebugCapture{}
	// 	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &dc); err != nil {
	// 		logger.Warningf("Error while parsing debug capture message: %s\n", err)
	// 	}
	// 	dc.Dump(dissect, data, prefix)
	// } else {
	// 	fmt.Printf("%s Unknonwn event: %+v\n", prefix, msg)
	// }
}

func (h *PipelineHandle) eventListener(killChan chan struct{}) error {

	logger.Infof("Starting bpf hook perf event listener...")
	var events *bpf.PerCpuEvents
	var err error

	go func() {
		// The first loops waits for the event map to be created by the bpf programs associated with the pipelines. NewPerCpuEvents() does not create this map; it is the responsibility of the bpfs to create and write to the map the events they wish to send to userspace.
	WaitLoop:
		for {
			select {
			case <-h.pipeKillChan:
				return
			case <-killChan:
				return
			default:
				exists, err := common.PathExists(h.perfConfig.MapPath)
				if err != nil {
					logger.Warningf("error while checking existence of bpf event map (ignoring): %s", err)
				}
				if exists {
					break WaitLoop
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		events, err = bpf.NewPerCpuEvents(h.perfConfig)
		if err != nil {
			logger.Errorf("error while openning bpf event map with PerfEventConfig (%+v): %s", h.perfConfig, err)
			return
		}

		logger.Infof("New bpf event map opened...")
		// The primary loop which reads events.
		for {
			select {
			case <-h.pipeKillChan:
				stopEventListener(events)
				return
			case <-killChan:
				stopEventListener(events)
				return
			default:
				todo, err := events.Poll(5000)
				if err != nil && err != syscall.EINTR {
					logger.Errorf("%s", err)
				}
				if todo > 0 {
					err = events.ReadAll(h.receiveEvent, h.lostEvent)
					if err != nil {
						logger.Errorf("error received while reading from perf buffer: %s", err)
					}
				}
			}
		}
	}()
	return nil
}

func stopEventListener(events *bpf.PerCpuEvents) {
	logger.Infof("Stopping bpf hook perf event listener")
	// This go routine terminates only if killChan is closed.
	if events != nil {
		lost, unknown := events.Stats()
		if lost != 0 || unknown != 0 {
			logger.Warningf("%d events lost, %d unknonwn notifications", lost, unknown)
		}
		err := events.CloseAll()
		if err != nil {
			logger.Errorf("%s", err)
		}
	}
	return
}
