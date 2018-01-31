package programs

import (
	"os"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// Open and create a pcap output file
func openPcap(pcapOut string) (*os.File, *pcapgo.Writer, error) {
	if pcapOut != "" {
		f, err := os.Create(pcapOut)
		w := pcapgo.NewWriter(f)

		// Write the PCAP global header
		w.WriteFileHeader(65536, layers.LinkTypeEthernet)

		return f, w, err
	}

	return nil, nil, nil
}
