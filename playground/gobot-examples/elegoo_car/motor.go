// +build example
//
// Do not build by default.

package main

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/api"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/firmata"
)

const (
	LED = "13"
	ENB = "6"
	IN3 = "9"
	IN4 = "11"
)

func main() {
	master := gobot.NewMaster()
	a := api.NewAPI(master)
	a.Start()

	firmataAdaptor := firmata.NewAdaptor("/dev/ttyACM0")
	enb := gpio.NewDirectPinDriver(firmataAdaptor, ENB)
	in3 := gpio.NewDirectPinDriver(firmataAdaptor, IN3)
	in4 := gpio.NewDirectPinDriver(firmataAdaptor, IN4)

	run := true
	work_motor := func() {
		enb.DigitalWrite(byte(1))
		gobot.Every(1*time.Second, func() {
			if run {
				in3.Off()
				in4.On() // Right wheel turning forwards.
			} else {
				in3.Off()
				in4.Off() // Right wheel stoped.
			}
			run = !run
		})
	}

	robot := gobot.NewRobot("motor",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{enb, in3, in4},
		work_motor,
	)

	master.AddRobot(robot)

	led := gpio.NewLedDriver(firmataAdaptor, LED)

	work_led := func() {
		gobot.Every(2*time.Second, func() {
			led.Toggle()
		})
	}

	robot_led := gobot.NewRobot("led",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{led},
		work_led,
	)

	master.AddRobot(robot_led)

	master.Start()
}
