package packetmmap

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
)

type AfpacketSnifferConf struct {
	// IFace is the specific interface to bind to.
	Iface afpacket.OptInterface `json:"iface"`
	// FrameSize is TPacket's tp_frame_size.
	FrameSize afpacket.OptFrameSize `json:"frame-size"`
	// BlockSize is TPacket's tp_block_size.
	BlockSize afpacket.OptBlockSize `json:"block-size"`
	// NumBlocks is TPacket's tp_block_nr.
	NumBlocks afpacket.OptNumBlocks `json:"num-blocks"`
	// BlockTimeout is TPacket v3's tp_retire_blk_tov.
	BlockTimeout int64 `json:"block-timeout"`
	// PollTimeout number of milliseconds that poll() should block waiting
	// for a file descriptor to become ready.
	PollTimeout int64 `json:"poll-timeout"`
}

func (c AfpacketSnifferConf) ValidateConf() error {
	// if string(c.Iface) == "" {
	// 	return fmt.Errorf("no network interface specified in configuration")
	// }
	if c.FrameSize < afpacket.OptFrameSize(0) {
		return fmt.Errorf("invalid frame size: %v", c.FrameSize)
	}
	if c.BlockSize < afpacket.OptBlockSize(0) {
		return fmt.Errorf("invalid block size: %v", c.BlockSize)
	}
	if c.NumBlocks < afpacket.OptNumBlocks(0) {
		return fmt.Errorf("invalid number of blocks: %v", c.NumBlocks)
	}
	if c.BlockTimeout < int64(0) {
		return fmt.Errorf("invalid block timeout: %v", c.BlockTimeout)
	}
	if c.PollTimeout < int64(0) {
		return fmt.Errorf("invalid poll timeout: %v", c.PollTimeout)
	}
	return nil
}

// afpacket version of the gopacket library
type AfpacketSniffer struct {
	handle *afpacket.TPacket
	Conf   AfpacketSnifferConf
}

func NewAfpacketSniffer(conf AfpacketSnifferConf) (*AfpacketSniffer, error) {

	// Configure the afpacket ring and bind it to the interface
	var tPacket *afpacket.TPacket
	opts := buildOptions(conf)
	tPacket, err := afpacket.NewTPacket(opts...)
	if err != nil {
		return nil, fmt.Errorf("Error opening afpacket interface: %s", err)
	}
	sniffer := &AfpacketSniffer{
		Conf:   conf,
		handle: tPacket,
	}

	return sniffer, nil
}

func (s *AfpacketSniffer) Close() {
	s.handle.Close()
}

func (s *AfpacketSniffer) ReadPacket() (data []byte, ci gopacket.CaptureInfo, err error) {
	return s.handle.ZeroCopyReadPacketData()
}

func buildOptions(conf AfpacketSnifferConf) []interface{} {
	var opts []interface{}
	opts = append(opts, conf.Iface)

	if conf.FrameSize > afpacket.OptFrameSize(0) {
		opts = append(opts, conf.FrameSize)
	}
	if conf.BlockSize > afpacket.OptBlockSize(0) {
		opts = append(opts, conf.BlockSize)
	}
	if conf.NumBlocks > afpacket.OptNumBlocks(0) {
		opts = append(opts, conf.NumBlocks)
	}
	if conf.BlockTimeout > int64(0) {
		opts = append(opts, afpacket.OptBlockTimeout(time.Duration(conf.BlockTimeout)*time.Millisecond))
	}
	if conf.PollTimeout > int64(0) {
		opts = append(opts, afpacket.OptPollTimeout(time.Duration(conf.PollTimeout)*time.Millisecond))
	}
	return opts
}

// Computes the block_size and the num_blocks in such a way that the
// allocated mmap buffer is close to but smaller than target_size_mb.
// The restriction is that the block_size must be divisible by both the
// frame size and page size. From PacketBeat.
func afpacketComputeSize(target_size_mb int, snaplen int, page_size int) (
	frame_size int, block_size int, num_blocks int, err error) {

	if snaplen < page_size {
		frame_size = page_size / (page_size / snaplen)
	} else {
		frame_size = (snaplen/page_size + 1) * page_size
	}

	// 128 is the default from the gopacket library so just use that
	block_size = frame_size * 128
	num_blocks = (target_size_mb * 1024 * 1024) / block_size

	if num_blocks == 0 {
		return 0, 0, 0, fmt.Errorf("Buffer size too small")
	}

	return frame_size, block_size, num_blocks, nil
}
