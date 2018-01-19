package controllers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gobot.io/x/gobot"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

// AutonomousDriveConf implements programapi.ProgramConf interface
type AutonomousDriveConf struct {
	//////////////////////////////////////////////////////
	// All controller confs should define the following fields. //
	//////////////////////////////////////////////////////
	MachineID      string   `json:"machine-id"`
	ID             string   `json:"id"`
	ControllerType string   `json:"controller-type"`
	Subscriptions  []string `json:"subscriptions"` // Message Type Subscriptions.

	////////////////////////////////////////////
	// The fields below are controller specific. //
	////////////////////////////////////////////
	LogFilePathname string `json:"log-file-path-name"` // logs will be wirten to this file.
}

func (c AutonomousDriveConf) ValidateConf() error {
	if c.ControllerType == "" {
		return fmt.Errorf("no controller type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no name specified")
	}
	if c.ControllerType != Controller_AutonomousDrive {
		return fmt.Errorf("invalid controller type specified, expected %s, but got %s", Controller_AutonomousDrive, c.ControllerType)
	}
	return nil
}

func (c AutonomousDriveConf) GetType() string {
	return c.ControllerType
}

func (c AutonomousDriveConf) GetID() string {
	return c.ID
}

func (c AutonomousDriveConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c AutonomousDriveConf) NewController(rcvQ *message.Queue, sndQ *message.Queue) (controllerapi.Controller, error) {

	drv := AutonomousDrive{
		State: &autonomousDriveInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}

	return &drv, nil
}

// AutonomousDrive implements the Controller interface
type AutonomousDrive struct {
	mu    sync.RWMutex
	State *autonomousDriveInternal
}

type autonomousDriveInternal struct {
	Conf     AutonomousDriveConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	Shutdown bool `json:"shutdown"`
	killChan chan struct{}

	stateMU                sync.RWMutex
	ultrasonicDataReadings []seguepb.SensorUltraSonicData
}

// Start: starts the controller logic.
func (d *AutonomousDrive) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.work()
	go d.State.commandIssuer()
	return nil
}

// Stop: stops the controller logic.
func (d *AutonomousDrive) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	d.State.Shutdown = true
	close(d.State.killChan)
	return nil
}

// work: generates and consumes messages.
func (d *AutonomousDrive) work() {

	for {
		select {
		case <-d.State.killChan:
			return
		default:
			// NOTE: Get will block this routine until either the controller is stopped or a message arrives.
			msg, shutdown := d.State.rcvQ.Get()
			if shutdown {
				logger.Debugf("stopping worker on controller %s", d.State.Conf.GetID())
				return
			}
			// logger.Debugf("autonomous-driver-controller received msg: %q", msg)
			d.State.messageHandler(msg)
			d.State.rcvQ.Done(msg)
		}
	}
}

func (d *AutonomousDrive) GetConf() controllerapi.ControllerConf {
	return d.State.Conf
}

func (d *AutonomousDrive) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *AutonomousDrive) Copy() controllerapi.Controller {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &AutonomousDrive{
		State: &autonomousDriveInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *AutonomousDrive) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *AutonomousDrive) UnmarshalJSON(data []byte) error {
	d.State = &autonomousDriveInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}

func (s *autonomousDriveInternal) messageHandler(msg message.Message) {
	s.stateMU.Lock()
	defer s.stateMU.Unlock()
	switch msg.ID.Type {
	case seguepb.MessageType_SensorUltraSonic:
		s.ultrasonicDataReadings = append(s.ultrasonicDataReadings, msg.Data.(seguepb.SensorUltraSonicData))
	default:
		logger.Warningf("autonomous-drive unknown message type %s", msg.ID.Type)
	}
	return
}

const (
	maxSpeedValue    = 210
	stopSpeedValue   = 75
	accelerationIncr = 15
)

func (s *autonomousDriveInternal) commandIssuer() {
	var version uint64
	speed := uint32(stopSpeedValue)
	fadeAmount := uint32(15)

	gobot.Every(1000*time.Millisecond, func() {
		s.stateMU.Lock()
		if s.Shutdown {
			s.stateMU.Unlock()
			return
		}

		if len(s.ultrasonicDataReadings) <= 1 {
			speed = stopSpeedValue
			logger.Infof("stopping!")
		} else if len(s.ultrasonicDataReadings) > 5 {
			if speed < maxSpeedValue {
				speed = speed + fadeAmount
			}
			logger.Infof("speeding up %d", speed)
		} else {
			if speed > stopSpeedValue {
				speed = speed - fadeAmount
			}
			logger.Infof("slowing down %d", speed)
		}
		driveMsg := message.Message{
			ID: seguepb.Message_MessageID{
				Type:    seguepb.MessageType_CmdDrive,
				SubType: "forward",
				Version: version,
			},
			Data: seguepb.CmdDriveData{Speed: speed},
		}
		s.sndQ.Add(driveMsg)
		version += 1
		s.ultrasonicDataReadings = nil
		s.stateMU.Unlock()
	})

}
