package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Arvinderpal/embd-project/common"
)

var (
	workingDir string
)

func init() {
	workingDir, _ = os.Getwd()

}

func createConfig(runDir, libDir string) *Config {
	config := NewConfig()
	config.RunDir = runDir
	config.LibDir = libDir
	config.NodeAddress = "123.123.123.123/16"
	config.Opts.Set(common.OptionDebug, true)
	return config
}

func TestDaemon_Simple(t *testing.T) {
	runDir := filepath.Join(workingDir, "test/daemon/run")
	libDir := filepath.Join(workingDir, "test/daemon/lib")
	config := createConfig(runDir, libDir)
	config.Opts.Set(common.OptionDisableRestore, true)
	_, err := NewDaemon(config)
	if err != nil {
		t.Fatalf("Error while creating daemon: %s", err)
		return
	}
}

func TestMain(m *testing.M) {

	code := m.Run()
	os.Exit(code)

}
