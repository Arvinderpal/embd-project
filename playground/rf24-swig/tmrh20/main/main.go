package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/Arvinderpal/RF24"
)

// *************** NOTE *************************
// This is the golang version of the gettingstarted.cpp that comes with RF24

const (
	CEPIN     = 25 // RPI_V2_GPIO_P1_22     = 25,  /*!< Version 2, Pin P1-22 */
	CSPIN     = 0  // BCM2835_SPI_CS0 = 0,     /*!< Chip Select 0 */
	PING_OUT  = true
	PING_BACK = false
)

func main() {

	node1Str := "1Node\n"
	node2Str := "2Node\n"
	var node1 [6]byte
	var node2 [6]byte
	copy(node1[:], node1Str)
	copy(node2[:], node2Str)
	// Radio pipe addresses for the 2 nodes to communicate.
	pipes := [2][6]byte{node1, node2}

	var (
		cepin uint16
		cspin uint16
	)

	cepin = CEPIN
	cspin = CSPIN
	radio := RF24.NewRF24(cepin, cspin)

	// bool role_ping_out = true, role_pong_back = false;
	// bool role = role_pong_back;

	fmt.Printf("starting radio...\n")

	// Setup and configure rf radio
	ok := radio.Begin()
	if !ok {
		fmt.Printf("error in radio.Begin() - likely cause is no response from module")
		return
	}

	// optionally, increase the delay between retries & # of retries
	radio.SetRetries(15, 15)
	// Dump the confguration of the rf unit for debugging
	radio.PrintDetails()

	role := PING_BACK
	fmt.Printf("\n ************ Role Setup ***********\n")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choose a role: Enter 0 for pong_back (default), 1 for ping_out (CTRL+C to exit): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	if input == "1" {
		fmt.Printf("Role: Ping Out, starting transmission\n")
		role = PING_OUT
	} else {
		fmt.Printf("Role: Pong Back, awaiting transmission\n")
		role = PING_BACK
	}

	fmt.Printf("\n ************ Select Node ***********\n")
	fmt.Print("Which node is this?: 1 '1Node' (default), 2 for '2Node' (CTRL+C to exit): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	if input == "2" {
		fmt.Printf("Write pipe: %s Read pipe: %s\n", string(pipes[1][:]), string(pipes[0][:]))
		radio.OpenWritingPipe(&pipes[1][0])
		radio.OpenReadingPipe(byte(1), &pipes[0][0])
	} else {
		fmt.Printf("Write pipe: %s Read pipe: %s\n", string(pipes[0][:]), string(pipes[1][:]))
		radio.OpenWritingPipe(&pipes[0][0])
		radio.OpenReadingPipe(byte(1), &pipes[1][0])
	}

	radio.StartListening()
	for {
		if role == PING_OUT {
			// First, stop listening so we can talk.
			radio.StopListening()
			fmt.Printf("Now sending...\n")
			// unsigned long time = millis();
			var curtime uint64
			curtime = 0x1234
			ptr := (uintptr)(unsafe.Pointer(&curtime))
			ok := radio.Write(uintptr(ptr), byte(8))
			if !ok {
				fmt.Printf("write failed.\n")
				// return
			}

			fmt.Printf("Response sent, waiting for reply\n")
			radio.StartListening()

			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(1 * time.Second)
				timeout <- true
			}()
		INNER_PING_OUT:
			for {
				select {
				case <-timeout:
					// the read from ch has timed out
					fmt.Printf("Failed, response timed out.\n")
					break INNER_PING_OUT
				default:
					if radio.Available() {
						// Grab the response, compare, and send to debugging spew
						var gottime uint64
						ptr := (uintptr)(unsafe.Pointer(&gottime))
						radio.Read(uintptr(ptr), byte(8))
						fmt.Printf("Response: %x\n", gottime)
						// Spew it
						// printf("Got response %lu, round-trip delay: %lu\n",got_time,millis()-got_time);
						break INNER_PING_OUT
					}
				}
			}

		} else { // Pong back role.  Receive each packet, dump it out, and send it back
			// if there is data ready
			if radio.Available() {
				var gottime uint64
			INNER_PINGBACK:
				for {
					if radio.Available() {
						ptr := (uintptr)(unsafe.Pointer(&gottime))
						radio.Read(uintptr(ptr), byte(8))
					} else {
						break INNER_PINGBACK
					}
				}

				radio.StopListening()

				var curtime uint64
				curtime = 0x9876
				ptr := (uintptr)(unsafe.Pointer(&curtime))
				ok := radio.Write(uintptr(ptr), byte(8))
				if !ok {
					fmt.Printf("write failed.\n")
				}

				// Now, resume listening so we catch the next packets.
				radio.StartListening()

				fmt.Printf("Response: %x\n", gottime)
			}
		}
		time.Sleep(1)
	}
}
