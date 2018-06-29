package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	"github.com/Arvinderpal/embd-project/common/controllerapi"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/prometheus/common/log"
)

type TestPluginConf struct {
	CEPin             uint16 `json:"ce-pin"`              // CE pin on device
	CSPin             uint16 `json:"cs-pin"`              // CS pin on device
	Channel           byte   `json:"channel"`             // RF Channel to use [0...255]
	Address           uint16 `json:"address"`             // Node address in Octal
	Master            bool   `json:"master"`              // Is this a master node?
	PollInterval      int    `json:"poll-interval"`       // We poll RF module every poll interval [units: Millisecond]
	HeartbeatInterval int    `json:"heartbeat-interval"`  // Child nodes send periodic heartbeats
	RouterWorkerCount int    `json:"router-worker-count"` // MASTER: number of workers to use per receive queue in the router
}

func main() {
	fmt.Printf("starting testplugin\n")
	reader := bufio.NewReader(os.Stdin)

	// meta conf
	text, _ := reader.ReadString('\n')
	metaConf := &controllerapi.MPIMetaConf{}
	err := json.Unmarshal([]byte(text), metaConf)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	fmt.Printf("MetaConf: %v \n", metaConf)

	text, _ = reader.ReadString('\n')
	fmt.Printf("test plugin conf dump: %s", text)
	conf := &TestPluginConf{}
	err = json.Unmarshal([]byte(text), conf)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	fmt.Printf("TestPluginConf: %v \n", conf)

	// GRPC conf
	text, _ = reader.ReadString('\n')
	fmt.Printf("grpc conf dump: %s\n", text)

	grpcConf := &controllerapi.GRPCConf{}
	err = json.Unmarshal([]byte(text), grpcConf)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	fmt.Printf("GRPCConf: %v \n", grpcConf)

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
	sendReceiveTestMessages(client)

	<-stop
	fmt.Printf("Shutting down the plugin...")
}

// sendTestMessage sends messages to segue via GRPC.
func sendReceiveTestMessages(client seguepb.MessengerClient) {
	// We send a series of envelopes where each envelope contains a single message. Of course, we can add multiple messages to an envelope as well

	stream, err := client.Messenger(context.Background())
	if err != nil {
		panic(fmt.Sprintf("%v.Messenger(_) = _, %v", client, err))
	}
	waitc := make(chan struct{})
	go func() {
		for {
			msgEnv, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				panic(fmt.Sprintf("Failed to receive a message : %v", err))
			}
			log.Infof("Got %d messages", len(msgEnv.Messages))
			for i, eMsg := range msgEnv.Messages {
				iMsg, err := message.ConvertToInternalFormat(eMsg)
				if err != nil {
					log.Errorf("grpc: messge convertion error: %s", err)
				}
				log.Infof("%d: %v", i+1, iMsg)
			}
		}
	}()

	var cmd *seguepb.MessageEnvelope
	cmdCh := make(chan *seguepb.MessageEnvelope, 10)
	go func() {
		for {
			select {
			case cmd = <-cmdCh:
				if err := stream.Send(cmd); err != nil {
					fmt.Errorf(fmt.Sprintf("failed to send message: %v", err))
				}
			}
		}
	}()

	// We'll send a test message every 2 seconds.
	version := 0
	tickChan := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-tickChan:
			var eMsgs []*seguepb.Message
			iMsg := message.Message{
				ID: seguepb.Message_MessageID{
					Type:    seguepb.MessageType_UnitTest,
					SubType: "",
					Version: uint64(version),
				},
				Data: &seguepb.UnitTestData{TestMessage: fmt.Sprintf("msg: %d", version)},
			}
			eMsg, err := message.ConvertToExternalFormat(iMsg)
			if err != nil {
				fmt.Errorf(fmt.Sprintf("error marshaling data: %s", err))
			} else {
				eMsgs = append(eMsgs, eMsg)
			}
			msgEnv := &seguepb.MessageEnvelope{eMsgs}
			version += 1
			cmdCh <- msgEnv
			break
		default:
			break
		}
	}
	stream.CloseSend()
	<-waitc
}
