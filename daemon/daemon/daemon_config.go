package daemon

import (
	"sync"

	"github.com/Arvinderpal/embd-project/pkg/option"
)

const (
	OptionSample = "SampleOption"
)

var (
	OptionSpecSample = option.Option{
		Description: "This is just a sample boolean option",
	}

	DaemonOptionLibrary = option.OptionLibrary{
		OptionSample: &OptionSpecSample,
	}
)

func init() {
}

// Config is the configuration used by Daemon.
type Config struct {
	LibDir string // library directory
	RunDir string // runtime directory

	NodeAddress string // Node IPv4 Address in CIDR notation

	// Options changeable at runtime
	Opts   *option.BoolOptions
	OptsMU sync.RWMutex
}

func NewConfig() *Config {
	return &Config{
		Opts: option.NewBoolOptions(&DaemonOptionLibrary),
	}
}
