package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
	"github.com/gogo/protobuf/proto"
	termbox "github.com/nsf/termbox-go"
	logging "github.com/op/go-logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

var (
	log = logging.MustGetLogger("remotecar")
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containning the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
	logFilePath        = flag.String("log_file_path", "/tmp/remotecar.s", "All logs are written to this file. These logs include the messages received")
)

var flg = 0

var mu sync.Mutex

var out = ""

func keyEventLoop(kch chan termbox.Event) {
	for {
		kch <- termbox.PollEvent()
	}
}

func noInputTimerLoop(ich chan bool) {
	for {
		ich <- true
		time.Sleep(time.Duration(600) * time.Millisecond)
	}
}

func cleanup() {
	termbox.Close()
	fmt.Println(out)
}

func marshallData(msgIn *seguepb.CmdDriveData) []byte {
	mhData, err := proto.Marshal(msgIn)
	if err != nil {
		panic(fmt.Sprintf("error marshaling data: %s", err))
	}
	return mhData
}

// runCarControl sends CmdDrive message to remote segue instance.
func runCarControl(client seguepb.MessengerClient, keyCh chan termbox.Event, noInputTimerCh chan bool) {
	// We send a series of envelopes where each envelope contains a single message. Of course, we can add multiple messages to an envelope as well.
	// NOTE: a better approach would have been to create Messages of with the internal format and then call ConvertToExternalFormat().
	forwardCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "forward",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(200)}),
			},
		},
	}
	backwardCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "backward",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(200)}),
			},
		},
	}
	leftCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "left",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(250)}),
			},
		},
	}
	rightCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "right",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(250)}),
			},
		},
	}
	stopCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "stop",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(0)}),
			},
		},
	}
	forwardRightCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "forward-right",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(200)}),
			},
		},
	}
	forwardLeftCmd := &seguepb.MessageEnvelope{
		Messages: []*seguepb.Message{
			&seguepb.Message{
				ID: &seguepb.Message_MessageID{
					Type:    seguepb.MessageType_CmdDrive,
					SubType: "forward-left",
					Version: 1,
				},
				Data: marshallData(&seguepb.CmdDriveData{Speed: uint32(200)}),
			},
		},
	}
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
					panic(fmt.Sprintf("failed to send message: %v", err))
				}
			}
		}
	}()

	fmt.Println("Control [forward: ->][backward: <-][esc: escape]")
	noInputDetected := true
OUTTTER_LOOP:
	for {
		if flg == 1 {
			break OUTTTER_LOOP
		}
		select {
		case key := <-keyCh:
			mu.Lock()
			switch {
			case key.Key == termbox.KeyEsc || key.Key == termbox.KeyCtrlC: //exit
				mu.Unlock()
				break OUTTTER_LOOP
			case key.Key == termbox.KeyArrowLeft || key.Ch == 's': //left
				cmdCh <- leftCmd
				break
			case key.Key == termbox.KeyArrowRight || key.Ch == 'd': //right
				cmdCh <- rightCmd
				break
			case key.Ch == 'a': // forward left
				cmdCh <- forwardLeftCmd
				break
			case key.Ch == 'f': // forward right
				cmdCh <- forwardRightCmd
				break
			case key.Key == termbox.KeyArrowUp || key.Ch == 'e': //up
				cmdCh <- forwardCmd
				break
			case key.Key == termbox.KeyArrowDown || key.Ch == 'w': //down
				cmdCh <- backwardCmd
				break
			default:
				break
			}
			noInputDetected = false
			mu.Unlock()
			break
		case <-noInputTimerCh:
			mu.Lock()
			if noInputDetected {
				// sndCmd = stopCmd
				cmdCh <- stopCmd
			}
			noInputDetected = true
			mu.Unlock()
			break
		default:
			break
		}
	}
	stream.CloseSend()
	<-waitc
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption

	// Logger setup:
	f, err := os.Create(*logFilePath)
	if err != nil {
		panic(err)
	}
	logWriter := bufio.NewWriter(f)
	common.SetupLOG(log, "DEBUG", logWriter)
	defer func() {
		logWriter.Flush()
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	// Options:
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			panic(fmt.Sprintf("Failed to create TLS credentials %v", err))
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Connect to remote peer
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		panic(fmt.Sprintf("fail to dial: %v", err))
	}
	defer conn.Close()

	// Create a new client:
	client := seguepb.NewMessengerClient(conn)

	err = termbox.Init()
	defer cleanup()
	if err != nil {
		panic(err)
	}

	keyCh := make(chan termbox.Event)
	noInputTimerCh := make(chan bool)

	go keyEventLoop(keyCh)
	go noInputTimerLoop(noInputTimerCh)

	if !terminal.IsTerminal(0) {
		go func() {
			stdin, _ := ioutil.ReadAll(os.Stdin)
			out = string(stdin)

			time.Sleep(1000 * time.Millisecond)
			flg = 1
		}()
	}

	// FIXME ignore output by other process excepting stdout e.g. git clone
	termbox.SetCursor(0, 0)
	runCarControl(client, keyCh, noInputTimerCh)

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}
