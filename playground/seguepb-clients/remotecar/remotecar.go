package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/gogo/protobuf/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containning the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func marshallData(msgIn *seguepb.CmdDriveData) []byte {
	mhData, err := proto.Marshal(msgIn)
	if err != nil {
		panic(fmt.Sprintf("error marshaling data: %s", err))
	}
	return mhData
}

// runCarControl sends CmdDrive message to remote segue instance.
func runCarControl(client seguepb.MessengerClient) {
	// We send a series of envelopes where each envelope contains a single message. Of course, we can add multiple messages to an envelope as well.
	envelopes := []*seguepb.MessageEnvelope{
		{
			Messages: []*seguepb.Message{
				&seguepb.Message{
					ID: &seguepb.Message_MessageID{
						Type:    seguepb.MessageType_CmdDrive,
						SubType: "forward",
						Version: 1,
					},
					Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(150)}),
				},
			},
		},
		{
			Messages: []*seguepb.Message{
				&seguepb.Message{
					ID: &seguepb.Message_MessageID{
						Type:    seguepb.MessageType_CmdDrive,
						SubType: "backward",
						Version: 2,
					},
					Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(150)}),
				},
			},
		},
	}
	stream, err := client.Messenger(context.Background())
	if err != nil {
		log.Fatalf("%v.Messenger(_) = _, %v", client, err)
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
				log.Fatalf("Failed to receive a message : %v", err)
			}
			log.Printf("Got %d messages", len(msgEnv.Messages))
		}
	}()
	for _, env := range envelopes {
		if err := stream.Send(env); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	stream.CloseSend()
	<-waitc
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := seguepb.NewMessengerClient(conn)

	// RouteChat
	runCarControl(client)
}
