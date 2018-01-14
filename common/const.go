package common

var (
	// Version number needs to be var since we override the value when building
	Version = "dev"
)

const (
	// SeguePath is the path where segue operational files are running.
	SeguePath     = "/var/run/segue"
	DefaultLibDir = "/usr/lib/segue"
	// SegueSock is the socket for the communication between the daemon and client.
	SegueSock = SeguePath + "/segue.sock"

	// GroupFilePath is the unix group file path.
	GroupFilePath = "/etc/group"
	// segue's unix group name.
	SegueGroupName = "segue"

	// RFC3339Milli is the RFC3339 with milliseconds for the default timestamp format
	// log files.
	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

	OptionDebug          = "Debug"
	OptionDisableRestore = "DisableRestore"
	OptionTestMode       = "TestMode"

	// Restore logic
	SnapshotFileName = "machine-state.json"
)
const (
	LED = "13"
)

const (
	Message_UltraSonic = "ultrasonic"
)
