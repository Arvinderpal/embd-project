package l1

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"github.com/Arvinderpal/matra/common"
	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/bpf"
	"github.com/vishvananda/netlink"
)

/*
* L1 is a simple bpf that counts packets recieved by the container.
 The code has these major components:
 1. go code: (this file) is responsible
		i. compiling and starting the bpf which runs in kernel space
		ii. creating the map resource which will be sharred with kernel bpf and userspace. Note that the c program must know the map name which we define as "l1_" + Container ID (see MapName)
		iii. userspace program that can interact with the bpf map
 2. C code: (kernel bpf) this code consists of a .c file which performs the 	necessary function:
		i.	the .c file is in the bpf/networking/l1 directory
		ii. the .h file is generated here (see writeBPFHeader()) and contains various definitions customized to the container. Since the file is different from one container to another, it is kept in the run directory (/var/run/matra/<container id>)
*/

// binary representation for encoding in binary structs
type IPv4 [4]byte

func (v4 IPv4) IP() net.IP {
	return v4[:]
}

func (v4 IPv4) String() string {
	return v4.IP().String()
}

type L1BPFConf struct {
	Mode      string `json:"mode"`      // Clsact mode (e.g. direct).
	Direction string `json:"direction"` // Direction of traffic: ingress, egress, ingress and egress.

	EPNetworkInfo *programapi.EndpointNetworkInfo // Network info of ep.

	// These are set inside the BPF code
	RunDir     string `json:"rundir"`      // where the bpf .o is kept among other things
	LibDir     string `json:"libdir"`      // where the bpf code (.c and .sh)  files are
	DisableJIT bool   `json:"disable-jit"` // JIT is enabled by default.
}

func (c L1BPFConf) ValidateConf() error {
	if c.Mode != "direct" {
		return fmt.Errorf("unknown mode in L1BPFConf: %s", c.Mode)
	}
	if c.Direction != "" && c.Direction != "ingress" && c.Direction != "egress" && c.Direction != "ingress,egress" {
		return fmt.Errorf("invalid direction specified (%s) in config", c.Direction)
	}
	if c.LibDir != "" {
		if exists, _ := common.PathExists(c.LibDir); !exists {
			return fmt.Errorf("specified libdir path (%s) does not exist", c.LibDir)
		}
	}
	return nil
}

type L1BPF struct {
	ContainerID string
	PipelineID  string
	Conf        L1BPFConf
	Map         *bpf.Map
}

const (
	L1MapName       = "l1_"
	L1MapPath       = common.BPFMatraGlobalMaps + "/" + L1MapName // TODO: move out of /tc/globals and into container specific dir
	L1BPFSrcDir     = "networking/l1"
	CHeaderFilePath = "/l1.h"
	CConfigFilePath = "/program_configs.h"
	L1MapNamePrefix = "l1_"
	MaxKeys         = 4 // should be same in bpf .h file
)

func constructMapName(prefix, containerID string) string {
	return prefix + containerID
}

func NewL1BPF(containerID, pipelineID string, conf L1BPFConf) *L1BPF {

	// create l1 map
	l1map := bpf.NewMap(L1MapPath+containerID,
		bpf.MapTypeHash,
		int(unsafe.Sizeof(L1MapKey{})),
		int(unsafe.Sizeof(L1MapValue{})),
		MaxKeys)

	// FIXME (awander): need to agree on some standard location where all sources will be kept: DefaultLibDir = "/usr/lib/matra"
	if conf.LibDir == "" {
		conf.LibDir = "/root/go/src/github.com/Arvinderpal/matra/pkg/programs/bpf"
	}
	conf.RunDir = common.MatraPath + "/" + containerID

	return &L1BPF{
		ContainerID: containerID,
		PipelineID:  pipelineID,
		Conf:        conf,
		Map:         l1map,
	}
}

func (p *L1BPF) Start() error {

	// compile bpf and attach it
	err := p.compileBase()
	if err != nil {
		return fmt.Errorf("Start failed: %s", err)
	}
	return nil
}

func (p *L1BPF) Stop() error {

	action := "stop"
	hostVeth, err := netlink.LinkByIndex(p.Conf.EPNetworkInfo.HostSideIfIndex)
	if err != nil {
		return fmt.Errorf("Error while fetching Link for veth index %v with MAC %s: %s", p.Conf.EPNetworkInfo.HostSideIfIndex, p.Conf.EPNetworkInfo.HostSideMAC, err)
	}
	ifName := hostVeth.Attrs().Name
	args := []string{p.Conf.LibDir, p.Conf.RunDir, action, ifName, p.Conf.Direction}

	out, err := exec.Command(filepath.Join(p.Conf.LibDir, L1BPFSrcDir+"/init.sh"), args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution %s %s failed: %s : command output:\n%s",
			filepath.Join(p.Conf.LibDir, L1BPFSrcDir+"/init.sh"),
			strings.Join(args, " "), err, out)
	}

	// Close the BPF map
	err = p.Map.Close()
	if err != nil {
		return fmt.Errorf("error while closing map: %s", err)
	}

	// TODO: We should delete the bpf map as well. Is it a simple matter of removing the /sys/fs/bpf entry?

	return nil
}

func (p *L1BPF) writeBPFHeader() error {
	headerPath := filepath.Join(p.Conf.RunDir, CHeaderFilePath)
	f, err := os.Create(headerPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %s", headerPath, err)

	}
	defer f.Close()

	fw := bufio.NewWriter(f)

	fmt.Fprint(fw, "/*\n")

	fmt.Fprintf(fw, ""+
		" * Container ID: %s\n"+
		" * Map Name: %s\n"+
		" * MAC: %s\n"+
		" * IPv4 address: %s\n"+
		" * Host Side MAC: %s\n"+
		" * Host Side Interface Index: %q\n"+
		" */\n\n",
		p.ContainerID, constructMapName(L1MapNamePrefix, p.ContainerID),
		p.Conf.EPNetworkInfo.MAC, p.Conf.EPNetworkInfo.IPv4.String(),
		p.Conf.EPNetworkInfo.HostSideMAC, p.Conf.EPNetworkInfo.HostSideIfIndex)

	fmt.Fprintf(fw, "#define CONTAINER_ID %s\n", p.ContainerID)
	fmt.Fprintf(fw, "#define MAP_NAME %s\n", constructMapName(L1MapNamePrefix, p.ContainerID))
	fw.WriteString(common.FmtDefineAddress("CONTAINER_MAC", p.Conf.EPNetworkInfo.MAC))

	/////////////
	// EXAMPLE:
	/////////////
	// /*
	//  * Container ID: 4520cbcde6f2a15d02d456a20a761e7dba8c2b2242ccdf7621f30594dda42b26
	//  * Map Name: ep_l1_4520cbcde6f2a15d02d456a20a761e7dba8c2b2242ccdf7621f30594dda42b26
	//  * MAC:
	//  * IPv4 address: 10.255.1.34
	//  * Host Side MAC: 1e:0c:15:53:92:23
	//  * Host Side Interface Index: '$'
	//  */
	// #define CONTAINER_ID 4520cbcde6f2a15d02d456a20a761e7dba8c2b2242ccdf7621f30594dda42b26
	// #define MAP_NAME ep_l1_4520cbcde6f2a15d02d456a20a761e7dba8c2b2242ccdf7621f30594dda42b26
	// #define CONTAINER_MAC { .addr = {  } }
	// #define CONTAINER_IP_ARRAY { .addr = { 0xa, 0xff, 0x1, 0x22 } }
	// #define CONTAINER_IP_BIGENDIAN 0xaff0122
	// #define CONTAINER_IP 0x2201ff0a
	// #define CONTAINER_HOST_SIDE_MAC { .addr = { 0x1e, 0xc, 0x15, 0x53, 0x92, 0x23 } }
	// #define CONTAINER_HOST_SIDE_IFC_IDX 36

	fw.WriteString(common.FmtDefineAddress("CONTAINER_IP_ARRAY", p.Conf.EPNetworkInfo.IPv4[12:]))
	fmt.Fprintf(fw, "#define CONTAINER_IP_BIGENDIAN %#x\n", binary.BigEndian.Uint32(p.Conf.EPNetworkInfo.IPv4[12:]))
	fmt.Fprintf(fw, "#define CONTAINER_IP %#x\n", binary.LittleEndian.Uint32(p.Conf.EPNetworkInfo.IPv4[12:]))

	fw.WriteString(common.FmtDefineAddress("CONTAINER_HOST_SIDE_MAC", p.Conf.EPNetworkInfo.HostSideMAC))
	fmt.Fprintf(fw, "#define CONTAINER_HOST_SIDE_IFC_IDX %d\n", p.Conf.EPNetworkInfo.HostSideIfIndex)

	// Endpoint options
	// NOTE(awander): good way to pass defines directly from cli to bpf:
	// fw.WriteString(ep.Opts.GetFmtList())

	fw.WriteString("\n")
	fw.Flush()

	// Config file which defines the perf event map among other things...
	configPath := filepath.Join(p.Conf.RunDir, CConfigFilePath)
	fConfig, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %s", configPath, err)

	}
	defer fConfig.Close()

	fwConfig := bufio.NewWriter(fConfig)
	fmt.Fprint(fwConfig, "/*\n")

	fmt.Fprintf(fwConfig, ""+
		" * AUTO GENERATED FILE\n"+
		" */\n\n")
	fwConfig.WriteString("#ifndef __LIB_CONFIGS_H_")
	fwConfig.WriteString("#define __LIB_CONFIGS_H_")
	fwConfig.WriteString("\n")
	fmt.Fprintf(fwConfig, "#define EVENTS_MAP_NAME %s", common.ConstructPerfEventMapName(p.ContainerID, p.PipelineID))
	fwConfig.WriteString("\n")
	fwConfig.WriteString("#endif")
	fwConfig.WriteString("\n")
	fwConfig.Flush()
	return nil
}

func (p *L1BPF) compileBase() error {
	var args []string
	var ifName string

	if err := p.writeBPFHeader(); err != nil {
		return fmt.Errorf("Unable to create BPF header file: %s", err)
	}

	hostVeth, err := netlink.LinkByIndex(p.Conf.EPNetworkInfo.HostSideIfIndex)
	if err != nil {
		return fmt.Errorf("Error while fetching Link for veth index %v with MAC %s: %s", p.Conf.EPNetworkInfo.HostSideIfIndex, p.Conf.EPNetworkInfo.HostSideMAC, err)
	}
	ifName = hostVeth.Attrs().Name

	args = []string{p.Conf.LibDir, p.Conf.RunDir, p.Conf.Mode, ifName, p.Conf.Direction}

	//./init.sh /usr/lib/matra /var/run/matra direct eth1 ingress
	out, err := exec.Command(filepath.Join(p.Conf.LibDir, L1BPFSrcDir+"/init.sh"), args...).CombinedOutput()
	if err != nil {
		fmt.Errorf("Command execution %s %s failed: %s : Command output:\n%s",
			filepath.Join(p.Conf.LibDir, L1BPFSrcDir+"/init.sh"),
			strings.Join(args, " "), err, out)
		return err
	}
	return nil
}

// NOTE(awander): Must match 'struct lb4_key' in "bpf/networking/l1.h"
// implements: bpf.MapKey
type L1MapKey struct {
	Address IPv4
}

// func (k L1MapKey) Map() *bpf.Map              { return prog }
func (k L1MapKey) NewValue() bpf.MapValue     { return &L1MapValue{} }
func (k *L1MapKey) GetKeyPtr() unsafe.Pointer { return unsafe.Pointer(k) }

func (k *L1MapKey) String() string {
	return fmt.Sprintf("%s", k.Address)
}

// Convert between host byte order and map byte order
func (k *L1MapKey) Convert() *L1MapKey {
	n := *k
	// n.Port = common.Swab16(n.Port)
	return &n
}

// func (k *L1MapKey) MapDelete() error {
// 	return k.Map().Delete(k)
// }

func NewKey(ip net.IP) *L1MapKey {
	key := L1MapKey{}
	copy(key.Address[:], ip.To4())
	return &key
}

// TODO(awander): Must match 'struct lb4_service' in "bpf/networking/l1.h"
type L1MapValue struct {
	TxCount uint16
	RxCount uint16
}

func NewL1MapValue(txcount, rxcount uint16) *L1MapValue {
	l1 := L1MapValue{
		TxCount: txcount,
		RxCount: rxcount,
	}
	return &l1
}

func (s *L1MapValue) GetValuePtr() unsafe.Pointer {
	return unsafe.Pointer(s)
}

func (v *L1MapValue) Convert() *L1MapValue {
	n := *v
	return &n
}

func (v *L1MapValue) String() string {
	// TODO(awander): should make this JSON format
	return fmt.Sprintf("%#x %#x", v.TxCount, v.RxCount)
}

func (p *L1BPF) UpdateElement(k, v, mapID string) error {

	var ip net.IP
	var err error
	var value uint64

	ip = net.ParseIP(k)
	if ip == nil {
		return fmt.Errorf("Unable to parsekey: %v", k)
	}
	value, err = strconv.ParseUint(v, 10, 16)
	if err != nil {
		return fmt.Errorf("Can't parse value: %s: %s", v, err)
	}

	l1key := NewKey(ip)
	l1value := l1key.NewValue().(*L1MapValue)
	l1value.TxCount = uint16(value)
	l1value.RxCount = uint16(value)

	if err = p.updateElement(l1key, l1value); err != nil {
		return fmt.Errorf("Map update failed for key=%s: %s", ip, err)
	}
	return nil
}

func (p *L1BPF) updateElement(key *L1MapKey, value *L1MapValue) error {
	if _, err := p.Map.OpenOrCreate(); err != nil {
		return err
	}

	return p.Map.Update(key.Convert(), value.Convert())
}

func (p *L1BPF) DeleteElement(k, mapID string) error {

	var ip net.IP
	var err error

	ip = net.ParseIP(k)
	if ip == nil {
		return fmt.Errorf("Unable to parsekey: %v", k)
	}

	l1key := NewKey(ip)

	if err = p.deleteElement(l1key); err != nil {
		return fmt.Errorf("Map delete failed for key=%s: %s", ip, err)
	}
	return nil
}

func (p *L1BPF) deleteElement(key *L1MapKey) error {
	return p.Map.Delete(key.Convert())
}

func (p *L1BPF) LookupElement(k, mapID string) (string, error) {
	var ip net.IP

	ip = net.ParseIP(k)
	if ip == nil {
		return "", fmt.Errorf("Unable to parsekey: %v", k)
	}

	l1key := NewKey(ip)

	val, err := p.lookupElement(l1key)
	if err != nil {
		return "", fmt.Errorf("Map lookup failed for key=%s: %s", ip, err)
	}
	return val.String(), nil

}

func (p *L1BPF) lookupElement(key *L1MapKey) (*L1MapValue, error) {
	var elem *L1MapValue

	val, err := p.Map.Lookup(key.Convert())
	if err != nil {
		return nil, err
	}

	elem = val.(*L1MapValue)

	return elem.Convert(), nil
}

// Dump2String dumps the entire map object in string format
// Note that input - mapID - is ignored since we only have a single map
func (p *L1BPF) Dump2String(mapID string) (string, error) {

	var dump string

	dumpit := func(key []byte, value []byte) (bpf.MapKey, bpf.MapValue, error) {
		// fmt.Printf("Key:\n%sValue:\n%s\n", hex.Dump(key), hex.Dump(value))
		dump = dump + fmt.Sprintf("Key:\n%sValue:\n%s\n", hex.Dump(key), hex.Dump(value))
		return nil, nil, nil
	}

	err := p.Map.Dump(dumpit, nil)
	if err != nil {
		return "", fmt.Errorf("Unable to dump map %s: %s\n", err)
	}
	return dump, nil
}
