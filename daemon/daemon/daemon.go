package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	common "github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/pkg/machine"
	"github.com/Arvinderpal/embd-project/pkg/option"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("segue-daemon")

type Daemon struct {
	machinesMU sync.RWMutex
	machines   map[string]*machine.Machine
	conf       *Config
}

func (d *Daemon) init() (err error) {

	if err = os.MkdirAll(common.SeguePluginsPath, 0777); err != nil {
		logger.Fatalf("Could not create runtime directory %s: %s", common.SeguePluginsPath, err)
	}
	globalsDir := filepath.Join(d.conf.RunDir, "globals")
	if err = os.MkdirAll(globalsDir, 0755); err != nil {
		logger.Fatalf("Could not create runtime directory %s: %s", globalsDir, err)
	}
	if err = os.Chdir(d.conf.RunDir); err != nil {
		logger.Fatalf("Could not change to runtime directory %s: \"%s\"",
			d.conf.RunDir, err)
	}

	return nil
}

// NewDaemon creates and returns a new Daemon with the parameters set in c.
func NewDaemon(c *Config) (*Daemon, error) {
	if c == nil {
		return nil, fmt.Errorf("Configuration is nil")
	}

	d := Daemon{
		conf:     c,
		machines: make(map[string]*machine.Machine),
	}

	err := d.init()
	if err != nil {
		logger.Fatalf("Error while initializing daemon: %s\n", err)
	}

	if !c.Opts.Opts[common.OptionDisableRestore] {
		err = d.restore()
		if err != nil {
			logger.Fatalf("Error while restoring machines: %s\n", err)
		}
	}

	return &d, err
}

func changedOption(key string, value bool, data interface{}) {
}

func (d *Daemon) Update(opts option.OptionMap) error {
	d.conf.OptsMU.Lock()
	defer d.conf.OptsMU.Unlock()

	if err := d.conf.Opts.Validate(opts); err != nil {
		return err
	}

	changes := d.conf.Opts.Apply(opts, changedOption, d)
	if changes == 0 {
		logger.Warningf("No options modified\n")
	}

	return nil
}
