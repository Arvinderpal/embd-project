Segue (aka embd-project) is extensible golang based software framework for embedded systems. Key features:
 * event-driven -- support custom message definitions via protobufs/GRPC and a pub/sub model among the components. 
 * controller, driver, adaptor and machine abstractions -- allow for a flexible means by which to easily add new functionality. 
 * machine plugin interface (MPI) -- allows external processes to communicate with the segue daemon.
 * REST API provides the control-plane over which to query and update the state of the segue daemon. 
 * supported plantforms include for arduino and raspberry pi but can be easily extended to others.
 * a wide range of hardware devices are supported via the [gobot](https://github.com/hybridgroup/gobot) project
 * snapshot and automatic restore of internal state when segue crashes or restarts
 * Wifi, RF24Network, LIRC support for communication among segue daemons and other entities (Zigbee support under works). 


Before getting into the details, let's consider the following example of an autonomous car project:
Segue is running on a raspberry pi. It contains controller(s) which form the "brain" of the car and are responsible for processing a wide range of messages generated from various sources in the system. Segue has an arduino adaptor layer that allows it to communicate with a USB attached arduino -- which serves as the interface to car hardware (e.g. motors, ultra-sonic sensor, etc). For each of these hardware components, we define drivers that map messages from segue into hardware commands and conversely map hardware events into messages for segue to process. The Pi may also be directly attached to several hardware components (e.g. an IR transceiver, LED, etc) and has appropriate drivers defined to handle the interactions.  The controller(s) may send/receives messages from other processes/agents located in the system and/or on the Internet. 

## Autonomous Car Model

![Alt text](docs/autonomous-car.png?raw=true "Autonomous Car Model")

We can create the above model, with the following commands. We first need to define our machine: 

	sudo ./segue daemon machine join ../scripts/configs/mh-1.json; 

This is is nothing more than an id of the machine we which to create:

	{
    "machine-id": "my-autonomous-car"
	}

Next, let's define our arduino adaptor. In this example, arduino is attached to usb port "/dev/ttyACM0"

	sudo ./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json

	{
	"machine-id": "my-autonomous-car",
	"confs": 
	[
		{
			"type": "adaptor_firmata_serial",
			"id": "ttyACM0",
			"conf": {
			    "address": "/dev/ttyACM0"
			}
		}
	]
	}

Now, we can start our hardware drivers. Let's start with our motors: 

	sudo ./segue daemon driver start ../scripts/configs/dualmotors-mh1.json

	{
	"machine-id": "my-autonomous-car",
	"confs": 
	[
		{
			"type": "driver_dualmotors",
			"id": "dualmotors-1",
			"adaptor-id": "ttyACM0",
			"conf": {
			    "right-motor": {
			    	"forward-pin": "11",
			    	"backward-pin": "9",
			    	"speed-pin": "6"
			    },
			    "left-motor": {
			    	"forward-pin": "7",
			    	"backward-pin": "8",
			    	"speed-pin": "5"
			    }
			},
			"subscriptions":
			[
				"CmdDrive"
			]
		}
	]
	}

There are few important things to note here. First, the adapotor-id corresponds to the arduino adaptor we created earlier. Second, we have a dual-motor driver and for each motor we must specify the pin number to which the functions are mapped to. Lastly, we define the message subscription "CmdDrive"; this will ensure that messages of this type are routed to our driver. These messages are generated by the controller we will define below.

Here is another driver we will need. You can find the details in the JSON file. 

	sudo ./segue daemon driver start ../scripts/configs/ultrasonic-mh1.json;

Finally, we need to define the "brains" of the machine -- our autonomous car controller:

	{
	"machine-id": "my-autonomous-car",
	"confs": 
	[
		{
			"type": "autonomous-drive",
			"id": "autonomous-drive-1",
			"conf": {
			    "log-file-path-name": "/tmp/autonomous-drive-1.log"
			},
			"subscriptions":
			[
				"SensorUltraSonic",
				"LIRCEvent"
			]
		}
	]
	}

The controller subscribes to two events -- those generated by the ultra-sonic driver and also IR events (we didn't show this above but you're welcome to look at the details yourself). The autonomous car controller is extremely simple, it uses the ultrasonic readings to events for the motors with the objective of avoiding objects. Additionally, the LIRC interface allows remote control of the car via an IR transmitter. 

## Segue Internals

It's fairly straight forward to extend segue to support new platforms and add controllers that meet your needs. Understanding the concepts [here](./docs/segue-internals.md) will give you a good start!
