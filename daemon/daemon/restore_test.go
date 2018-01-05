package daemon

import (
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Arvinderpal/matra/common"
	"github.com/Arvinderpal/matra/common/networkhookapi"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/endpoint"
)

func TestDaemon_SnapshotRestore(t *testing.T) {
	t.Logf("Working dir: %s", workingDir)
	runDir := filepath.Join(workingDir, "test/snapshot/run")
	libDir := filepath.Join(workingDir, "test/snapshot/lib")

	config := createConfig(runDir, libDir)
	disableRestore := true
	config.Opts.Set(common.OptionDisableRestore, disableRestore)
	d, err := NewDaemon(config)
	if err != nil {
		t.Fatalf("Error while creating daemon: %s", err)
		return
	}

	setupEP1(t, d) // ep1: just an empty ep
	setupEP2(t, d) // ep2: ep with pmmap hook
	setupEP3(t, d) // ep3: ep with pmmap hook and sample pipeline

	// We'll start another daemon process that will do the restore of the above eps
	disableRestore = false
	config.Opts.Set(common.OptionDisableRestore, disableRestore)
	d, err = NewDaemon(config)
	if err != nil {
		t.Fatalf("Error while creating daemon: %s", err)
		return
	}

	checkEP1(t, d)
	checkEP2(t, d)
	checkEP3(t, d)

}

func setupEP1(t *testing.T, d *Daemon) {
	t.Logf("Joining endpoint")
	ep1 := endpoint.Endpoint{
		ContainerID: "ep1",
		NetworkInfo: &programapi.EndpointNetworkInfo{
			IfName: "eth0",
			IPv4:   net.IPv4(192, 168, 1, 2),
		},
	}
	err := d.EndpointJoin(ep1)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func setupEP2(t *testing.T, d *Daemon) {
	t.Logf("Joining endpoint")
	ep2 := endpoint.Endpoint{
		ContainerID: "ep2",
		NetworkInfo: &programapi.EndpointNetworkInfo{
			IfName: "eth0",
			IPv4:   net.IPv4(192, 168, 1, 2),
		},
	}
	err := d.EndpointJoin(ep2)
	if err != nil {
		t.Fatalf("%s", err)
	}

	const pmmapConf = `
{
	"type": "packet_mmap",
	"conf": {
	    "container-id": "ep2",
	    "afpacket-conf": {
		    "iface": "eth0",
		    "frame-size": 4096,
		    "block-size": 524288, 
		    "num-blocks": 128,
		    "block-timeout": 10
		}
	}
}
`
	confB, err := ioutil.ReadAll(strings.NewReader(pmmapConf))
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = d.AttachNetworkHook(confB)
	if err != nil {
		t.Fatalf("%s", err)
	}

}

func setupEP3(t *testing.T, d *Daemon) {
	t.Logf("Joining endpoint")
	ep3 := endpoint.Endpoint{
		ContainerID: "ep3",
		NetworkInfo: &programapi.EndpointNetworkInfo{
			IfName: "eth0",
			IPv4:   net.IPv4(192, 168, 1, 2),
		},
	}
	err := d.EndpointJoin(ep3)
	if err != nil {
		t.Fatalf("%s", err)
	}

	const pmmapConf = `
{
	"type": "packet_mmap",
	"conf": {
	    "container-id": "ep3",
	    "afpacket-conf": {
		    "iface": "eth0",
		    "frame-size": 4096,
		    "block-size": 524288, 
		    "num-blocks": 128,
		    "block-timeout": 10
		}
	}
}
`
	confB, err := ioutil.ReadAll(strings.NewReader(pmmapConf))
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = d.AttachNetworkHook(confB)
	if err != nil {
		t.Fatalf("%s", err)
	}

	const pipeConf = `
{
	"id": "ep3-pmmap-pipe1",
	"type": "packet_mmap",
	"container": "ep3",
	"confs": 
	[
		{
			"type": "unittest_program",
			"conf": {
				"name": "unittest-program-1"
			}
		},
		{
			"type": "unittest_program",
			"conf": {
			    "name": "unittest-program-2"
			}
		}
	]
}

`
	confB, err = ioutil.ReadAll(strings.NewReader(pipeConf))
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = d.StartPipeline(confB)
	if err != nil {
		t.Fatalf("%s", err)
	}

}

// ep1 is just simple endpoint with no hooks attached.
func checkEP1(t *testing.T, d *Daemon) {
	epName := "ep1"
	iface := "eth0"
	ep, err := d.EndpointGet(epName)
	if err != nil {
		t.Fatalf("Error while getting ep: %s", err)
	}
	if ep == nil {
		t.Fatalf("Expected %s to be restored. Not the case.", epName)
	}
	if ep.ContainerID != epName {
		t.Fatalf("Expected %s to be restored. Got %s.", epName, ep.ContainerID)
	}
	if ep.NetworkInfo.IfName != iface {
		t.Fatalf("Expected interface %s after restore. Got: %s", iface, ep.NetworkInfo.IfName)
	}
}

// ep2 has the packet mmap hook attached.
func checkEP2(t *testing.T, d *Daemon) {
	epName := "ep2"
	iface := "eth0"
	ep, err := d.EndpointGet(epName)
	if err != nil {
		t.Fatalf("Error while getting ep: %s", err)
	}
	if ep == nil {
		t.Fatalf("Expected %s to be restored. Not the case.", epName)
	}
	if ep.ContainerID != epName {
		t.Fatalf("Expected %s to be restored. Got %s.", epName, ep.ContainerID)
	}
	if ep.NetworkInfo.IfName != iface {
		t.Fatalf("Expected interface %s after restore. Got: %s", iface, ep.NetworkInfo.IfName)
	}
	if len(ep.Hooks) == 0 {
		t.Fatalf("Expected hook to be restored. Found %d hooks.", len(ep.Hooks))
	}
	if ep.Hooks[0].GetConf().GetType() != networkhookapi.NetworkHookType_PacketMMAP {
		t.Fatalf("Expected %s to be restored. Got: %s.", networkhookapi.NetworkHookType_PacketMMAP, ep.Hooks[0].GetConf().GetType())

	}
}

// ep3 has packet mmap and a pipeline attached.
func checkEP3(t *testing.T, d *Daemon) {
	epName := "ep3"
	iface := "eth0"
	ep, err := d.EndpointGet(epName)
	if err != nil {
		t.Fatalf("Error while getting ep: %s", err)
	}
	if ep == nil {
		t.Fatalf("Expected %s to be restored. Not the case.", epName)
	}
	if ep.ContainerID != epName {
		t.Fatalf("Expected %s to be restored. Got %s.", epName, ep.ContainerID)
	}
	if ep.NetworkInfo.IfName != iface {
		t.Fatalf("Expected interface %s after restore. Got: %s", iface, ep.NetworkInfo.IfName)
	}
	if len(ep.Hooks) == 0 {
		t.Fatalf("Expected hook to be restored. Found %d hooks.", len(ep.Hooks))
	}
	hook := ep.Hooks[0]
	if hook.GetConf().GetType() != networkhookapi.NetworkHookType_PacketMMAP {
		t.Fatalf("Expected %s to be restored. Got: %s.", networkhookapi.NetworkHookType_PacketMMAP, hook.GetConf().GetType())

	}
	if len(hook.GetPipelines()) == 0 {
		t.Fatalf("Expected pipeline to be restored. Found %d pipelines.", len(hook.GetPipelines()))
	}
	pipeID := "ep3-pmmap-pipe1"
	pipe := hook.LookupPipeline(pipeID)
	if pipe == nil {
		t.Fatalf("Expected pipeline %s to be restored. Not the case.", pipeID)
	}
	if len(pipe.Programs) < 2 {
		t.Fatalf("Expected 2 program to be restored. Found %d.", len(pipe.Programs))
	}
	prog1 := pipe.Programs[0]
	prog2 := pipe.Programs[1]
	if prog1.GetConf().GetName() != "unittest-program-1" {
		t.Fatalf("Expected program %s to be restored. Got: %s", "unittest-program-1", prog1.GetConf().GetName())

	}
	if prog2.GetConf().GetName() != "unittest-program-2" {
		t.Fatalf("Expected program %s to be restored. Got: %s", "unittest-program-2", prog2.GetConf().GetName())
	}
}

func copyGoldenSnapshot(t *testing.T, runDir, libDir, epName string) {

	golden := filepath.Join(runDir, epName, testFileName+".golden")
	testCopy := filepath.Join(runDir, epName, testFileName)
	// copy golden file to the file from which restore happens
	if err := common.CopyFile(golden, testCopy); err != nil {
		t.Fatalf("Error while copying golden file: %s", err)
	}
}
