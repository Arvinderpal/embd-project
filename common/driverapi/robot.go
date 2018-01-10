package driverapi

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"

	"gobot.io/x/gobot"

	multierror "github.com/hashicorp/go-multierror"
)

var robotCount uint

// Robot is a named entity that manages a collection of devices.
// It contains its own work routine and a collection of
// custom commands to control a robot remotely via the Gobot api.
type Robot struct {
	Name    string
	Work    func()
	devices *gobot.Devices
	trap    func(chan os.Signal)
	AutoRun bool
	running atomic.Value
	done    chan bool
}

// NewRobot returns a new Robot. It supports the following optional params:
//
//		name:	string with the name of the Robot. A name will be automatically generated if no name is supplied.
//		[]Device: Devices which are automatically started and stopped with the robot
//		func(): The work routine the robot will execute once all devices have been initialized and started
//
func NewRobot(v ...interface{}) *Robot {
	r := &Robot{
		Name:    fmt.Sprintf("%X", robotCount),
		devices: &gobot.Devices{},
		done:    make(chan bool, 1),
		trap: func(c chan os.Signal) {
			signal.Notify(c, os.Interrupt)
		},
		AutoRun: true,
		Work:    nil,
	}
	robotCount += 1

	for i := range v {
		switch v[i].(type) {
		case string:
			r.Name = v[i].(string)
		case []gobot.Device:
			log.Println("Initializing devices...")
			for _, device := range v[i].([]gobot.Device) {
				d := r.AddDevice(device)
				log.Println("Initializing device", d.Name(), "...")
			}
		case func():
			r.Work = v[i].(func())
		}
	}

	r.running.Store(false)
	log.Println("Robot", r.Name, "initialized.")

	return r
}

// Start a Robot's Devices, and work.
func (r *Robot) Start(args ...interface{}) (err error) {
	if len(args) > 0 && args[0] != nil {
		r.AutoRun = args[0].(bool)
	}
	log.Println("Starting Robot", r.Name, "...")
	if derr := r.Devices().Start(); derr != nil {
		err = multierror.Append(err, derr)
		log.Println(err)
		return
	}
	if r.Work == nil {
		r.Work = func() {}
	}

	log.Println("Starting work...")
	go func() {
		r.Work()
		<-r.done
	}()

	r.running.Store(true)
	if r.AutoRun {
		c := make(chan os.Signal, 1)
		r.trap(c)

		// waiting for interrupt coming on the channel
		<-c

		// Stop calls the Stop method on itself, if we are "auto-running".
		r.Stop()
	}

	return
}

// Stop stops a Robot's Devices
func (r *Robot) Stop() error {
	var result error
	log.Println("Stopping Robot", r.Name, "...")
	err := r.Devices().Halt()
	if err != nil {
		result = multierror.Append(result, err)
	}

	r.done <- true
	r.running.Store(false)
	return result
}

// Running returns if the Robot is currently started or not
func (r *Robot) Running() bool {
	return r.running.Load().(bool)
}

// Devices returns all devices associated with this Robot.
func (r *Robot) Devices() *gobot.Devices {
	return r.devices
}

// AddDevice adds a new Device to the robots collection of devices. Returns the
// added device.
func (r *Robot) AddDevice(d gobot.Device) gobot.Device {
	*r.devices = append(*r.Devices(), d)
	return d
}

// Device returns a device given a name. Returns nil if the Device does not exist.
func (r *Robot) Device(name string) gobot.Device {
	if r == nil {
		return nil
	}
	for _, device := range *r.devices {
		if device.Name() == name {
			return device
		}
	}
	return nil
}
