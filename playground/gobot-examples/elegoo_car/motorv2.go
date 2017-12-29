// +build example
//
// Do not build by default.

package main

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
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
	firmataAdaptor := firmata.NewAdaptor("/dev/ttyACM0")
	rightmotor := gpio.NewMotorDriver(firmataAdaptor, ENB)
	rightmotor.ForwardPin = IN4
	rightmotor.BackwardPin = IN3

	work := func() {
		speed := byte(0)
		fadeAmount := byte(15)

		rightmotor.Forward(speed)
		gobot.Every(500*time.Millisecond, func() {
			rightmotor.Speed(speed)
			speed = speed + fadeAmount
			if speed == 0 || speed == 255 {
				fadeAmount = -fadeAmount
			}
			fmt.Printf("%d, ", speed)
		})
	}

	robot := gobot.NewRobot("rightmotorBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{rightmotor},
		work,
	)

	robot.Start()
}
