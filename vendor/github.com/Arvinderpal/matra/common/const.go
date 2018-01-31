package common

var (
	// Version number needs to be var since we override the value when building
	Version = "dev"
)

const (
	// MatraPath is the path where matra operational files are running.
	MatraPath     = "/var/run/matra"
	DefaultLibDir = "/usr/lib/matra"
	// MatraSock is the socket for the communication between the daemon and client.
	MatraSock = MatraPath + "/matra.sock"
	// DefaultContainerMAC represents a dummy MAC address for the containers.
	DefaultContainerMAC = "AA:BB:CC:DD:EE:FF"

	// GroupFilePath is the unix group file path.
	GroupFilePath = "/etc/group"
	// matra's unix group name.
	MatraGroupName = "matra"

	// BPFMap is the file that contains the BPF Map for the host.
	BPFMapRoot         = "/sys/fs/bpf"
	BPFMatraGlobalMaps = BPFMapRoot + "/tc/globals"
	// Debug Events (Global)
	BPFMatraGlobalEventMap = BPFMapRoot + "/tc/globals/matra_events"

	// RFC3339Milli is the RFC3339 with milliseconds for the default timestamp format
	// log files.
	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

	OptionDebug          = "Debug"
	OptionDisableRestore = "DisableRestore"
	OptionTestMode       = "TestMode"

	// Restore logic
	SnapshotFileName = "ep-state.json"
)

// NFQ constants
const (
	AF_INET = 2

	NF_DROP   uint = 0
	NF_ACCEPT uint = 1
	NF_STOLEN uint = 2
	NF_QUEUE  uint = 3
	NF_REPEAT uint = 4
	NF_STOP   uint = 5

	NF_DEFAULT_PACKET_SIZE uint32 = 0xffff

	VERDICT_UNKOWN uint = 42
)
