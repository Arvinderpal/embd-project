package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	"github.com/Arvinderpal/RF24"
	"github.com/Arvinderpal/RF24Network"
	logging "github.com/op/go-logging"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/Arvinderpal/embd-project/pkg/controllers/rf24networknodebackend"
)

var (
	logger = logging.MustGetLogger("plugin-rf24networknode")
)

// RF24NetworkNodeConf implements ControllerConf interface
type RF24NetworkNodeConf struct {
	MachineID     string   `json:"machine-id"`
	ID            string   `json:"id"`            // IMPORTANT: ID should be unique for across all nodes in the RF24Network
	Subscriptions []string `json:"subscriptions"` // Message Type Subscriptions.

	CEPin             uint16 `json:"ce-pin"`              // CE pin on device
	CSPin             uint16 `json:"cs-pin"`              // CS pin on device
	Channel           byte   `json:"channel"`             // RF Channel to use [0...255]
	Address           uint16 `json:"address"`             // Node address in Octal
	Master            bool   `json:"master"`              // Is this a master node?
	PollInterval      int    `json:"poll-interval"`       // We poll RF module every poll interval [units: Millisecond]
	HeartbeatInterval int    `json:"heartbeat-interval"`  // Child nodes send periodic heartbeats
	RouterWorkerCount int    `json:"router-worker-count"` // MASTER: number of workers to use per receive queue in the router
}

func (c RF24NetworkNodeConf) ValidateConf() error {
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.Channel < 0 || c.Channel > 255 {
		return fmt.Errorf("invalid channel %d specified in configuration", c.Channel)
	}
	if c.Master {
		if c.Address != 00 {
			return fmt.Errorf("master node must have address 00 (octal), but found %#0", c.Address)
		}
	}
	if !c.Master {
		if c.HeartbeatInterval <= 0 {
			return fmt.Errorf("heartbeat interval must be > 0: %d", c.HeartbeatInterval)
		}
	}
	// These checks are based on http://tmrh20.github.io/RF24NetworkNode/
	// The largest node address is 05555, so 3,125 nodes are allowed on a single channel. I believe this comes from the fact that nRF24L01(+) radio can listen actively on up to 6 radios at once. Thus, the max number of children a node can have is capped from 1...5. That's 5 children and 1 parent.
	if (c.Address&00007) > 00005 || (c.Address&00070) > 00050 || (c.Address&00700) > 00500 || (c.Address&07000) > 05000 {
		return fmt.Errorf("invalid address %#0. address must be less than 05555 (octal)", c.Address)
	}

	return nil
}

func (c RF24NetworkNodeConf) GetID() string {
	return c.ID
}

func (c RF24NetworkNodeConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func NewRF24NetworkNode(conf RF24NetworkNodeConf) *RF24NetworkNode {

	drv := RF24NetworkNode{
		State: &RF24NetworkNodeInternal{
			Conf:     conf,
			rcvQ:     message.NewQueue(conf.ID),
			sndQ:     message.NewQueue(conf.ID),
			killChan: make(chan struct{}),
		},
	}
	return &drv
}

// RF24NetworkNode implements the Controller interface
type RF24NetworkNode struct {
	mu    sync.RWMutex
	State *RF24NetworkNodeInternal
}

type RF24NetworkNodeInternal struct {
	Conf     RF24NetworkNodeConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}

	backend rf24networknodebackend.RF24NetworkNodeBackend
	radio   RF24.RF24
	network RF24Network.RF24Network
}

// Start: starts the RF24NetworkNode.
func (d *RF24NetworkNode) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.State.radio = RF24.NewRF24(d.State.Conf.CEPin, d.State.Conf.CSPin)
	d.State.network = RF24Network.NewRF24Network(d.State.radio)

	logger.Infof("starting rf24 radio...\n")

	ok := d.State.radio.Begin() // Setup and configure rf radio
	if !ok {
		return fmt.Errorf("error in radio.Begin() - likely cause is no response from module")
	}

	time.Sleep(5 * time.Millisecond)
	// Dump the confguration of the rf unit for debugging
	d.State.radio.PrintDetails() // TODO: we should send these to logger.

	logger.Infof("starting rf24 network...\n")
	d.State.network.Begin(d.State.Conf.Channel, d.State.Conf.Address)

	// We convert subscriptions from []string to []int32 (i.e.MessageType value). We do this to minimize payload size of the outgoing packet.
	var subs []int32
	for _, s := range d.State.Conf.Subscriptions {
		subs = append(subs, seguepb.MessageType_value[s])
	}

	if d.State.Conf.Master {
		logger.Infof("creating master node for rf24 network...\n")
		d.State.backend = rf24networknodebackend.NewRF24NetworkNodeMaster(d.State.Conf.ID, d.State.Conf.Address, subs, d.State.network, d.State.Conf.PollInterval, d.State.sndQ, d.State.rcvQ, d.State.Conf.RouterWorkerCount)
	} else {
		logger.Infof("creating child node for rf24 network...\n")
		d.State.backend = rf24networknodebackend.NewRF24NetworkNodeChild(d.State.Conf.ID, d.State.Conf.Address, subs, d.State.network, d.State.Conf.PollInterval, d.State.Conf.HeartbeatInterval, d.State.sndQ, d.State.rcvQ)
	}

	go d.work()
	return nil
}

// Stop: stops the RF24NetworkNode.
func (d *RF24NetworkNode) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)
	d.State.backend.Stop()

	// We need to cleanup RF24 and RF24Network objects
	RF24Network.DeleteRF24Network(d.State.network)
	d.State.radio.PowerDown() // not sure if ths necessary
	RF24.DeleteRF24(d.State.radio)
	return nil
}

// work: processes messages to be sent out and listens for incoming messages on RF.
func (d *RF24NetworkNode) work() {

	err := d.State.backend.Run()
	if err != nil {
		logger.Errorf("Listening stopped with an error: %s", err)
	}
}

func (d *RF24NetworkNode) GetConf() RF24NetworkNodeConf {
	return d.State.Conf
}

func (d *RF24NetworkNode) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *RF24NetworkNode) Copy() *RF24NetworkNode {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &RF24NetworkNode{
		State: &RF24NetworkNodeInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *RF24NetworkNode) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *RF24NetworkNode) UnmarshalJSON(data []byte) error {
	d.State = &RF24NetworkNodeInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}

func main() {
	logger.Infof("Starting rf24networknode")
	reader := bufio.NewReader(os.Stdin)

	// meta conf
	text, _ := reader.ReadString('\n')
	metaConf := &controllerapi.MPIMetaConf{}
	err := json.Unmarshal([]byte(text), metaConf)
	if err != nil {
		logger.Errorf("%s", err)
		return
	}
	logger.Infof("MetaConf: %v \n", metaConf)

	// rf24 conf
	text, _ = reader.ReadString('\n')
	conf := &RF24NetworkNodeConf{}
	err = json.Unmarshal([]byte(text), conf)
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
	// We put the metaConf data in rf24 conf:
	conf.MachineID = metaConf.MachineID
	conf.ID = metaConf.ID
	conf.Subscriptions = metaConf.Subscriptions
	logger.Infof("RF24NetworkNodeConf: %v \n", conf)
	if err := (*conf).ValidateConf(); err != nil {
		panic(fmt.Sprintf("Could not validate conf: %s", err))
	}

	// GRPC conf
	text, _ = reader.ReadString('\n')
	grpcConf := &controllerapi.GRPCConf{}
	err = json.Unmarshal([]byte(text), grpcConf)
	if err != nil {
		logger.Errorf("%s", err)
		return
	}
	logger.Infof("GRPCConf: %v \n", grpcConf)

	// Setup GRPC Client:
	var opts []grpc.DialOption
	if grpcConf.TLSEnabled {
		var caFile string
		if grpcConf.CertFile == "" {
			caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, "x.test.youtube.com" /*serverHostOverride*/)
		if err != nil {
			panic(fmt.Sprintf("Failed to create TLS credentials %v", err))
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Connect to remote peer
	serverAddr := fmt.Sprintf("%s:%d", grpcConf.HostAddress, grpcConf.Port)
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		panic(fmt.Sprintf("fail to dial: %v", err))
	}
	defer conn.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Create a new client:
	client := seguepb.NewMessengerClient(conn)
	// Create rf24network node
	node := NewRF24NetworkNode(*conf)
	err = node.Start()
	if err != nil {
		panic(fmt.Sprintf("node start failed: %s", err))
	}

	listen(client, node)

	<-stop

	node.Stop()
	logger.Infof("Shutting down the plugin...")
}

// listen: listens for message from segue side and places those messages in
// the rcvQ and also listens on the rf network and places those message in the
// sndQ.
func listen(client seguepb.MessengerClient, node *RF24NetworkNode) {

	stream, err := client.Messenger(context.Background())
	if err != nil {
		panic(fmt.Sprintf("%v.Messenger(_) = _, %v", client, err))
	}
	waitc := make(chan struct{})

	// Messages received from Segue daemon (via grpc):
	go func() {
		for {
			select {
			case <-node.State.killChan:
				return
			default:
				msgEnv, err := stream.Recv()
				if err == io.EOF {
					// read done.
					close(waitc)
					return
				}
				if err != nil {
					logger.Errorf("Failed to receive a message : %v", err)
					return
				}
				logger.Debugf("Got %d messages", len(msgEnv.Messages))
				for i, eMsg := range msgEnv.Messages {
					iMsg, err := message.ConvertToInternalFormat(eMsg)
					if err != nil {
						logger.Errorf("grpc: messge convertion error: %s", err)
					}
					logger.Debugf("%d: %v", i+1, iMsg)
					node.State.rcvQ.Add(iMsg)
				}
			}
		}
	}()

	// Messages received from rf network to be sent to segue (via grpc):
	const maxMessagesPerEnv = 5
	go func() {
		for {
			select {
			case <-node.State.killChan:
				return
			default:
				if node.State.sndQ.IsShuttingDown() {
					return
				}
				var eMsgs []*seguepb.Message
				// IMPORTANT: this assumes only a single consumer of the queue.
				for i := 0; i < maxMessagesPerEnv; i++ {
					iMsg, shutdown := node.State.sndQ.Get()
					if shutdown {
						logger.Debugf("stopping rf24 network message receive routine")
						return
					}
					eMsg, err := message.ConvertToExternalFormat(iMsg)
					if err != nil {
						logger.Errorf(fmt.Sprintf("error converting message to external format: %s", err))
					} else {
						eMsgs = append(eMsgs, eMsg)
					}
					node.State.sndQ.Done(iMsg)
					if node.State.sndQ.Len() == 0 {
						// If the queue is empty, we don't want to block on
						// the Get(), so we'll just send the messages we have.
						break
					}
				}
				msgEnv := &seguepb.MessageEnvelope{eMsgs}
				if err := stream.Send(msgEnv); err != nil {
					logger.Errorf("error while sending to segue (continuing anyways): %s", err)
				}
			}
		}
	}()
}
