package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

// GRPCMessageServerConf implements ControllerConf interface
type GRPCMessageServerConf struct {
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

	HostAddress string `json:"host-address"` // IP address or just localhost
	Port        int    `json:"port"`         // Port to listen on

	TLSEnabled bool   `json:"tls-enabled"` // Will enable TLS on server
	CertFile   string `json:"cert-file"`   // TLS certfile (optional)
	KeyFile    string `json:"key-file"`    // TLS Key (optional)
}

func (c GRPCMessageServerConf) ValidateConf() error {
	if c.ControllerType == "" {
		return fmt.Errorf("no controller type specified")
	}
	if c.ID == "" {
		return fmt.Errorf("no id specified")
	}
	if c.ControllerType != Controller_GRPCMessageServer {
		return fmt.Errorf("invalid controller type specified, expected %s, but got %s", Controller_GRPCMessageServer, c.ControllerType)
	}
	if c.HostAddress == "" {
		return fmt.Errorf("no hostname specified in configuration")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d specified in configuration", c.Port)
	}
	if c.TLSEnabled {
		if c.CertFile == "" || c.KeyFile == "" {
			return fmt.Errorf("must specifiy both a cert file and a key file when enabling TLS")
		}
	}
	return nil
}

func (c GRPCMessageServerConf) GetType() string {
	return c.ControllerType
}

func (c GRPCMessageServerConf) GetID() string {
	return c.ID
}

func (c GRPCMessageServerConf) GetSubscriptions() []string {
	return c.Subscriptions
}

func (c GRPCMessageServerConf) NewController(rcvQ *message.Queue, sndQ *message.Queue) (controllerapi.Controller, error) {

	drv := GRPCMessageServer{
		State: &grpcMessageServerInternal{
			Conf:     c,
			rcvQ:     rcvQ,
			sndQ:     sndQ,
			killChan: make(chan struct{}),
		},
	}
	return &drv, nil
}

// GRPCMessageServer implements the Controller interface
type GRPCMessageServer struct {
	mu    sync.RWMutex
	State *grpcMessageServerInternal
}

type grpcMessageServerInternal struct {
	Conf     GRPCMessageServerConf `json:"conf"`
	rcvQ     *message.Queue
	sndQ     *message.Queue
	killChan chan struct{}

	grpcServer *grpc.Server
	streamQs   []*message.Queue // per stream queue for outgoing messages
	sQCount    int              // used to identify each stream queue
}

// Start: starts the controller logic.
func (d *GRPCMessageServer) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	go d.work()
	return nil
}

// Stop: stops the controller logic.
func (d *GRPCMessageServer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.State.rcvQ.ShutDown()
	d.State.sndQ.ShutDown()
	close(d.State.killChan)
	if d.State.grpcServer != nil {
		d.State.grpcServer.Stop()
	}
	return nil
}

// work: starts a grpc server and processes requests and responses.
func (d *GRPCMessageServer) work() {

	// Get internal messages from rcvQ and send them to all currently active streams.
	go func() {
		for {
			select {
			case <-d.State.killChan:
				return
			default:
				msg, shutdown := d.State.rcvQ.Get()
				if shutdown {
					logger.Debugf("stopping send routine in worker on controller %s", d.State.Conf.GetID())
					return
				}

				d.mu.RLock()
				logger.Debugf("grpc: (internal) message received (# of receivers %d): %v", len(d.State.streamQs), msg)
				for _, sq := range d.State.streamQs {
					sq.Add(msg)
				}
				d.State.rcvQ.Done(msg)
				d.mu.RUnlock()
			}
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", d.State.Conf.HostAddress, d.State.Conf.Port))
	if err != nil {
		logger.Errorf("failed to listen (%s:%d): %v", d.State.Conf.HostAddress, d.State.Conf.Port, err)
		return
	}
	var opts []grpc.ServerOption
	if d.State.Conf.TLSEnabled {
		logger.Infof("grpc: starting tls server")
		creds, err := credentials.NewServerTLSFromFile(d.State.Conf.CertFile, d.State.Conf.KeyFile)
		if err != nil {
			logger.Errorf("Failed to generate credentials %v", err)
			return
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	d.State.grpcServer = grpc.NewServer(opts...)
	seguepb.RegisterMessengerServer(d.State.grpcServer, d)
	d.State.grpcServer.Serve(lis) // Stop will ensure this go routine exists
}

func (d *GRPCMessageServer) GetConf() controllerapi.ControllerConf {
	return d.State.Conf
}

func (d *GRPCMessageServer) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return fmt.Sprintf("%#v", d)
}

func (d *GRPCMessageServer) Copy() controllerapi.Controller {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cpy := &GRPCMessageServer{
		State: &grpcMessageServerInternal{
			Conf: d.State.Conf,
		},
	}
	return cpy
}

func (d *GRPCMessageServer) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return json.Marshal(d.State)
}

func (d *GRPCMessageServer) UnmarshalJSON(data []byte) error {
	d.State = &grpcMessageServerInternal{}
	err := json.Unmarshal(data, d.State)
	return err
}

func (d *GRPCMessageServer) Messenger(stream seguepb.Messenger_MessengerServer) error {
	// Add a new queue on which we will receive messages from the work() func; these messages will be then be sent on the stream.
	d.mu.Lock()
	id := fmt.Sprintf("%d", d.State.sQCount)
	d.State.sQCount += 1
	sq := message.NewQueue(id)
	d.State.streamQs = append(d.State.streamQs, sq)
	d.mu.Unlock()
	defer func() {
		logger.Debugf("grpc: stoping messenger")
		sq.ShutDown()
		d.mu.Lock()
		for i, q := range d.State.streamQs {
			if q.ID() == id {
				d.State.streamQs = append(d.State.streamQs[:i], d.State.streamQs[i+1:]...)
				break
			}
		}
		d.mu.Unlock()
	}()
	// Send routine
	go func() {
		defer func() {
			logger.Debugf("grpc: stoping messenger-send routine")
		}()

		logger.Debugf("grpc: starting messenger-send routine")
		// Read all the messages in the streams send queue every X
		// milliseconds. All the messages will be put in the envelope.
		// The idea being that we'll minimize network stack overhead.
		// However, this approach introduces latency; alternatives include
		// reducing the interval or just sending a message as soon as it arrives.
		tickChan := time.NewTicker(time.Millisecond * 100).C
		for {
			select {
			case <-d.State.killChan:
				return
			case <-tickChan:
				if sq.IsShuttingDown() {
					return
				}
				var eMsgs []*seguepb.Message
				// NOTE: this assumes only a single consumer of the queue.
				sqLen := sq.Len() // TODO: may want to cap this to say 25 msgs
				if sqLen == 0 {
					continue
				}
				logger.Debugf("grpc: sending %d messages", sqLen)
				for i := 0; i < sqLen; i++ {
					iMsg, shutdown := sq.Get()
					if shutdown {
						logger.Debugf("stopping sender routine for grpc-stream-send-queue-%s", id)
						return
					}
					eMsg, err := message.ConvertToExternalFormat(iMsg)
					if err != nil {
						logger.Errorf(fmt.Sprintf("error marshaling data in routine for grpc-stream-send-queue-%s: %s", id, err))
					} else {
						eMsgs = append(eMsgs, eMsg)
					}
					sq.Done(iMsg)
				}
				msgEnv := &seguepb.MessageEnvelope{eMsgs}
				if err := stream.Send(msgEnv); err != nil {
					logger.Errorf("error while sending on grpc-stream-send-queue-%s (exiting send routine): %s", id, err)
					return
				}
			}
		}
	}()
	logger.Debugf("grpc: starting messenger-receive routine")
	for {
		select {
		case <-d.State.killChan:
			return nil
		default:
			msgPBEnv, err := stream.Recv()
			if err == io.EOF {
				logger.Debugf("grpc: exiting Messanger() EOF recieved")
				return nil
			}
			if err != nil {
				logger.Errorf("grpc: error: %s", err)
				return err
			}
			for _, msgPB := range msgPBEnv.Messages {
				// We go through the messages in the envelope, converting them into
				// the internal message format and sending them to the internal
				// routing system.
				msg, err := message.ConvertToInternalFormat(msgPB)
				if err != nil {
					logger.Errorf("grpc: messge convertion error: %s", err)
				} else {
					logger.Debugf("grpc: received msg %v", msg)
					d.State.sndQ.Add(msg)
				}
			}
		}
	}
}
