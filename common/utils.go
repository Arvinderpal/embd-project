package common

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	l "github.com/op/go-logging"
)

func ConstructPerfEventMapName(containerID, pipelineID string) string {
	return "matra_events_" + pipelineID + "_" + containerID
}

// goArray2C transforms a byte slice into its hexadecimal string representation.
// Example:
// array := []byte{0x12, 0xFF, 0x0, 0x01}
// fmt.Print(GoArray2C(array)) // "{ 0x12, 0xff, 0x0, 0x1 }"
func goArray2C(array []byte) string {
	ret := "{ "
	for i, e := range array {
		if i == 0 {
			ret = ret + fmt.Sprintf("%#x", e)
		} else {
			ret = ret + fmt.Sprintf(", %#x", e)
		}
	}

	return ret + " }"
}

// FmtDefineAddress returns the a define string from the given name and addr.
// Example:
// fmt.Print(FmtDefineAddress("foo", []byte{1, 2, 3})) // "#define foo { .addr = { 0x1, 0x2, 0x3 } }\n"
func FmtDefineAddress(name string, addr []byte) string {
	return fmt.Sprintf("#define %s { .addr = %s }\n", name, goArray2C(addr))
}

// FmtDefineArray returns the a define string from the given name and array.
// Example:
// fmt.Print(FmtDefineArray("foo", []byte{1, 2, 3})) // "#define foo { 0x1, 0x2, 0x3 }\n"
func FmtDefineArray(name string, array []byte) string {
	return fmt.Sprintf("#define %s %s\n", name, goArray2C(array))
}

// SetupLOG sets up logger with the correct parameters
func SetupLOG(logger *l.Logger, logLevel string, out io.Writer) {
	hostname, _ := os.Hostname()
	// fileFormat := l.MustStringFormatter(
	// 	`%{time:` + RFC3339Milli + `} ` + hostname +
	// 		` %{level:.4s} %{id:03x} %{shortfunc} > %{message}`,
	// )
	fileFormat := l.MustStringFormatter(
		`%{color}%{time:15:04:05.000} ` + hostname +
			` %{longfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	level, err := l.LogLevel(logLevel)
	if err != nil {
		logger.Fatal(err)
	}
	var backend *l.LogBackend
	if out != nil {
		backend = l.NewLogBackend(out, "", 0)
	} else {
		backend = l.NewLogBackend(os.Stderr, "", 0)
	}

	oBF := l.NewBackendFormatter(backend, fileFormat)

	backendLeveled := l.SetBackend(oBF)
	backendLeveled.SetLevel(level, "")
	logger.SetBackend(backendLeveled)
}

// GetGroupIDByName returns the group ID for the given grpName.
func GetGroupIDByName(grpName string) (int, error) {
	f, err := os.Open(GroupFilePath)
	if err != nil {
		return -1, err
	}
	defer f.Close()
	br := bufio.NewReader(f)
	for {
		s, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return -1, err
		}
		p := strings.Split(s, ":")
		if len(p) >= 3 && p[0] == grpName {
			return strconv.Atoi(p[2])
		}
	}
	return -1, fmt.Errorf("group %q not found", grpName)
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// PathExists returns whether the given file or directory exists or not
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// IsAlive will check if the process is alive or not.
// Returns true if the process is alive or false otherwise.
func IsAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

func GetProcess(pid int) *os.Process {
	if !IsAlive(pid) {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	return p
}

// GetFile will open filepath.
// Returns a tuple with a file and an error in case there's any.
func GetFile(filepath string) (*os.File, error) {
	return os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
}

// WriteFile will write the info on array of bytes b to filepath. It will set
// the file permission mode to 0660
// Returns an error in case there's any.
func WriteFile(filepath string, b []byte) error {
	return ioutil.WriteFile(filepath, b, 0660)
}

// DeleteFile will delete filepath permanently.
// Returns an error in case there's any.
func DeleteFile(filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	err = os.Remove(filepath)
	return err
}
