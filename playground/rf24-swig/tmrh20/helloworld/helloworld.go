package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/Arvinderpal/RF24"
	"github.com/Arvinderpal/RF24Network"
)

// *************** NOTE *************************
// This is the golang version of the gettingstarted.cpp that comes with RF24

const (
	CEPIN             = 25 // RPI_V2_GPIO_P1_22     = 25,  /*!< Version 2, Pin P1-22 */
	CSPIN             = 0  // BCM2835_SPI_CS0 = 0,     /*!< Chip Select 0 */
	CHANNEL           = 90
	TX_NODE           = 0x01
	RX_NODE           = 0x00
	TRANSMIT_INTERVAL = 1 // in seconds
)

func main() {

	var (
		cepin uint16
		cspin uint16
	)

	cepin = CEPIN
	cspin = CSPIN
	radio := RF24.NewRF24(cepin, cspin)

	network := RF24Network.NewRF24Network(radio)

	fmt.Printf("starting radio...\n")

	// Setup and configure rf radio
	ok := radio.Begin()
	if !ok {
		fmt.Printf("error in radio.Begin() - likely cause is no response from module")
		return
	}

	time.Sleep(5 * time.Millisecond)
	// Dump the confguration of the rf unit for debugging
	radio.PrintDetails()

	fmt.Printf("\n ************ Role Setup ***********\n")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choose a role: Enter 0 for TX (default), 1 for RX (CTRL+C to exit): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	if input == "1" {
		fmt.Printf("Role: RX\n")
		rx(network)
	} else {
		fmt.Printf("Role: TX\n")
		tx(network)
	}

}

func tx(network RF24Network.RF24Network) {
	var (
		this_node  uint16
		other_node uint16
	)
	this_node = TX_NODE
	other_node = RX_NODE

	interval := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(TRANSMIT_INTERVAL * time.Second)
			interval <- true
		}
	}()

	network.Begin( /*channel*/ byte(CHANNEL) /*node address*/, this_node)

	for {
		select {
		case <-interval:
			fmt.Printf(".\n")
			network.Update()
			var payload int64
			payload = time.Now().UnixNano()
			ptr := (uintptr)(unsafe.Pointer(&payload))
			header := RF24Network.NewRF24NetworkHeader(other_node)
			ok := network.Write(header, uintptr(ptr), uint16(8))
			if !ok {
				fmt.Printf("write failed.\n")
			}
		}
	}

}

func rx(network RF24Network.RF24Network) {
	var (
		this_node uint16
		// other_node uint16
	)

	this_node = RX_NODE
	// other_node = TX_NODE
	network.Begin( /*channel*/ byte(CHANNEL) /*node address*/, this_node)

	for {
		network.Update()
	INNER_LOOP:
		for {
			fmt.Printf(".\n")
			if network.Available() { // Is there anything ready for us?
				var payload int64
				ptr := (uintptr)(unsafe.Pointer(&payload))
				var header RF24Network.RF24NetworkHeader
				network.Read(header, uintptr(ptr), uint16(8))
				fmt.Printf("received payload %x (%d)\n", payload)
				break INNER_LOOP
			}
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(TRANSMIT_INTERVAL * time.Second)
	}
}
