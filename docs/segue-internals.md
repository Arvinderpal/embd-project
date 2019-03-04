Segue is written entirely in go and makes extensive use of go routines and go chans. It's fairly straight forward to extend segue to support your needs. Understanding the concepts below will give you a good start!

## Segue Machine

Segue defines a Machine object to encapsulate all the details for a particular machine (e.g. autonomous car):

	type Machine struct {
		mutex       sync.RWMutex
		MachineID   string                 `json:"machine-id"`     // Machine ID.
		Adaptors    []*AdaptorWrapper      `json:"adaptors"`       // Machine adaptor for communicating with hardware.
		Drivers     []*DriverWrapper       `json:"drivers"`        // Drivers that are part of this machine.
		Controllers []*ControllerWrapper   `json:"controllers"`    // Controllers that are part of this machine.
		MsgRouter   *message.MessageRouter `json:"message-router"` // Route messages between drivers/controllers.
		Opts        *option.BoolOptions    `json:"options"`        // Machine options.
		Status      *MachineStatus         `json:"status,omitempty"`
	}

## Segue Messages

Segue makes use of protobufs for message definition and processing. The segue Message is of the following form (see common/seguepb/segue.proto):

	message Message {
	message MessageID {
		MessageType Type				= 1;
		string SubType 					= 2;
		uint64 Version 					= 3;
		string Qualifier				= 4;
	}
	MessageID 		ID        			= 1;
	bytes      		Data  				= 2; 
	}

It essentailly consists of an ID and Data. The ID identifies the message type and Data is type specific data packed into a byte slice. Here are some message types that have been defined so far: 

	
	enum MessageType {
		UnitTest							= 0;
		SensorUltraSonic					= 1;
		CmdDrive							= 2;
		CmdLEDSwitch						= 3;
		LIRCEvent							= 4;
		RF24NetworkNodeHeartbeat			= 100;
		// NOTE: Any changes here should also be reflected in message.go
	}

Let's look at an example of a specific message:

	syntax = "proto3";
	package seguepb;
	message CmdDriveData {
		uint32 		speed 				= 1;
	}

For each message type, protobufs will generated the appropriate marshal/unmarshal code for us. So, all we need to do is call the appropriate marshal/unmarshal func based on the message type. This is done in common/message.go where we can easily convert message between their external (seguepb.Message) and internal format:

	func ConvertToInternalFormat(m *seguepb.Message) (Message, error) {
		switch m.GetID().GetType() {
		case seguepb.MessageType_CmdDrive:
			data := &seguepb.CmdDriveData{}
			err := proto.Unmarshal(m.GetData(), data)
			if err != nil {
				return Message{}, err
			}
			return Message{
				ID:   *m.ID,
				Data: data,
			}, nil
			...
	func ConvertToExternalFormat(iMsg Message) (*seguepb.Message, error) {
	switch iMsg.ID.GetType() {
	case seguepb.MessageType_CmdDrive:
		data, err := proto.Marshal(iMsg.Data.(*seguepb.CmdDriveData))
		if err != nil {
			return nil, err
		}
		return &seguepb.Message{
			ID:   &iMsg.ID,
			Data: data,
		}, nil

## Message Pub/Sub Model and Internal Routing 

Within sugue, controller and drivers can publish and subscribe to specific message types (we saw an example of this in the autonomous car example). Segue will ensure that messages are routed according to these subscriptions. This functionality is implemented in MessageRouter: 

	type MessageRouter struct {
		mu       sync.RWMutex
		RouteMap map[seguepb.MessageType][]*Queue
	}

When a component wants to subscribe to a particular message type, it first creates a queue on which to receive the messages on. MessageRouter maintains a set of queues for each message type; all message generated/arriving in segue will be sent to the MessageRouter which will then forward the messages to the appropriate listeners. 

## Drivers

Segue abstraction for the underlying hardware is the Driver object. Current set of implemented drivers can be found `pkg/drivers`. All drivers implement the Driver interface: 


	type Driver interface {
		Start() error
		Stop() error
		GetConf() DriverConf
		Copy() Driver
		String() string
	}

Let's look at the `dualmotors.go` Driver as an example. 

	type DualMotors struct {
		mu    sync.RWMutex
		State *dualMotorsInternal
	}
	type dualMotorsInternal struct {
		Conf       DualMotorsConf    `json:"conf"`
		RightMotor *gpio.MotorDriver `json:"right-motor"`
		LeftMotor  *gpio.MotorDriver `json:"left-motor"`
		robot      *driverapi.Robot
		Running    bool `json:"running"`
		killChan   chan struct{}
		rcvQ       *message.Queue
		sndQ       *message.Queue
	}

First, the RightMotor, LeftMotor and Robot fields are for the gobot logic that implements the underlying hardware interface. Second, two queues are defined, rcvQ and sndQ, corresponing to the Message receive and send queues for the driver. Lastly, we define a killChan; all drivers run inside their own go routine and may start additional sub routines; the killChan is used to stop all associated go routines. 

Dual-motors work func is nothing more than a go routine that waits for CmdDrive messages to arrive and maps those messages to the appropriate commands in gobot. 

	func (d *DualMotors) work() {
	for {
		select {
		case <-d.State.killChan:
			return
		default:
			// NOTE: Get will block this routine until either the controller is stopped or a message arrives.
			msg, shutdown := d.State.rcvQ.Get()
			if shutdown {
				logger.Debugf("stopping worker on driver %s", d.State.Conf.GetID())
				return
			}
			if msg.ID.Qualifier == d.State.Conf.Qualifier {
				logger.Debugf("dualmotors: received msg: %q", msg)
				switch msg.ID.Type {
				case seguepb.MessageType_CmdDrive:
					d.ProcessDriveCmd(msg)
				default:
					logger.Errorf("dualmotors: unknown message type %s", msg.ID.Type)
				}
			}
			d.State.rcvQ.Done(msg)
		}
	}
	
All Drivers must also implement the configuration interface below:

	type DriverConf interface {
		NewDriver(adaptorapi.Adaptor, *message.Queue, *message.Queue) (Driver, error)
		ValidateConf() error
		GetType() string
		GetID() string
		GetAdaptorID() string
		GetSubscriptions() []string
	}

## Controllers 

Controlllers follow a pattern similar to Drivers. The following interface (`common/controllerapi/controllerapi.go`) must be implemented by all controllers:
	
	type Controller interface {
		Start() error
		Stop() error
		GetConf() ControllerConf
		Copy() Controller
		String() string
	}



### Machine-Plugin Interface (MPI)

Developers can make use of the MPI (machine plugin interface) to develop functions outside of segue and still have them interact with the segue framework through GRPC. For example, the RF24Network plugin (see plugins/rf24networknode.go) makes use of MPI. MPI Controller is responsible for starting/stoping the external plugin. Let's look at the internals of the MPI controller:
	
	type MPI struct {
		mu    sync.RWMutex
		State *MPIInternal
	}
	type MPIInternal struct {
		Conf     MPIConf `json:"conf"`
		rcvQ     *message.Queue
		sndQ     *message.Queue
		killChan chan struct{}
		procMaster *procwatcher.ProcMaster   // Pointer to the one and only master.
		proc       procwatcher.ProcContainer // proc is the plugin process.
		grpcServer *grpc.Server
	}

MPI controller is very similar to other controllers, but with a few exceptions. First, MPI implements process control functions; that is, MPI is capable of starting/stoping the processes of the underlying plugins it sets up. Second, it runs a GRPC server which forms the communication path between segue and the plugin process. 

Let's look at the RF24Network plugin as an example (`plugins/rf24networknode`). The base RF24Network code is written in C/C++; however, this specific plugin is written go and makes use of cgo/swig for integration with the base RF24Network C/C++ code. Note that we could have just as well done the plugin in C/C++. The following JSON configuration is used to create an MPI controller which manages the the RF24Network plugin:

	{
	"machine-id": "mh1",
	"confs": 
	[
		{
			"type": "mpi",
			"id": "mpi-controller-rf24network-master-mh1",
			"conf": {
				"plugin-name": "rf24networknode",
				"keepalive": false,
			    "plugin-conf": {
				    "ce-pin": 25,
				    "cs-pin": 0,
				    "channel": 90,
				    "address": 0,
				    "master": true,
				    "poll-interval": 100,
				    "router-worker-count": 2
				},
				"grpc-conf": {
			    	"host-address": "localhost",
			    	"port": 30000,
			    	"tls-enabled": false
			    }
			},
			"subscriptions":
			[
				"CmdDrive",
				"UnitTest"
			]
		}
	]
	}

First, note the plugin name "rf24networknode" is used to identify the process executable (defaults to `plugins` directory). Second, additional plugin specific configuration is part of the "plugin-conf" field. Third, the "grpc-conf" is used to setup a GRPC on the MPI controller side and a client on the plugin executable side. Lastly, we have the subscriptions that the plugin wishes to receive. 

## Adaptors

In segue, adaptors provide the abstration needed to support different embedded platforms (e.g. Arduino/Firmata, RaspberryPI) . All adaptors must implement the following interface: 

	type Adaptor interface {
		Attach() error
		Detach() error
		DigitalWrite(pin string, level byte) (err error)
		PwmWrite(string, byte) (err error)
		ServoWrite(string, byte) (err error)
		DigitalRead(string) (val int, err error)
		AnalogRead(string) (val int, err error)
		GetConf() AdaptorConf
		Copy() Adaptor
		String() string
	}

