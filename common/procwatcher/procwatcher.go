package procwatcher

// Builds on top of: github.com/topfreegames/apm

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Arvinderpal/embd-project/common"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("proc-watcher")

func init() {
	common.SetupLOG(logger, "DEBUG", nil)
}

type ProcContainer interface {
	Start() error
	Stop() error
	ForceStop() error
	GracefullyStop() error
	Restart() error
	Delete() error
	IsAlive() bool
	Identifier() string
	ShouldKeepAlive() bool
	AddRestart()
	SetStatus(status string)
	GetPid() int
	GetStatus() *ProcStatus
	Watch() *watcherProcStatus
	WatchNonChild() *watcherProcStatus
	release()
	Restore(*os.Process)
}

// Proc is a os.Process wrapper with Status and more info
type Proc struct {
	Name string
	Cmd  string
	Args []string
	// Path      string
	Pidfile   string
	Infile    string
	Outfile   string
	Errfile   string
	KeepAlive bool
	Pid       int
	Status    *ProcStatus
	process   *os.Process
	Sys       *syscall.SysProcAttr
}

// ProcStatus is a wrapper with the process current status.
type ProcStatus struct {
	Status   string
	Restarts int
}

// SetStatus will set the process string status.
func (proc_status *ProcStatus) SetStatus(status string) {
	proc_status.Status = status
}

// IncrRestart will add one restart to the process status.
func (proc_status *ProcStatus) IncrRestart() {
	proc_status.Restarts++
}

// Restore will simply copy the os.Process values into the proc
// When matra crashes/restarts, Restore is used to link snort hooks back to
// their snort processes.
func (proc *Proc) Restore(p *os.Process) {
	proc.Pid = p.Pid
	proc.process = p
	proc.Status.SetStatus("restored")
}

// Start will execute the command Cmd that should run the process. It will also create an out, err and pidfile in case they do not exist yet.
// Returns an error in case there's any.
func (proc *Proc) Start() error {
	inFile, err := common.GetFile(proc.Infile)
	if err != nil {
		return err
	}
	outFile, err := common.GetFile(proc.Outfile)
	if err != nil {
		return err
	}
	errFile, err := common.GetFile(proc.Errfile)
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	procAtr := &os.ProcAttr{
		Dir: wd,
		Env: os.Environ(),
		Files: []*os.File{
			// os.Stdin,
			inFile,
			outFile,
			errFile,
		},
		Sys: proc.Sys,
	}
	args := append([]string{proc.Name}, proc.Args...)
	process, err := os.StartProcess(proc.Cmd, args, procAtr)
	if err != nil {
		return err
	}
	proc.process = process
	proc.Pid = proc.process.Pid
	err = common.WriteFile(proc.Pidfile, []byte(strconv.Itoa(proc.process.Pid)))
	if err != nil {
		return err
	}

	proc.Status.SetStatus("started")
	return nil
}

// ForceStop will forcefully send a SIGKILL signal to process killing it instantly.
// Returns an error in case there's any.
func (proc *Proc) ForceStop() error {
	if proc.process != nil {
		err := proc.process.Signal(syscall.SIGKILL)
		proc.Status.SetStatus("stopped")
		proc.release()
		return err
	}
	return errors.New("Process does not exist.")
}

// GracefullyStop will send a SIGTERM signal asking the process to terminate.
// The process may choose to die gracefully or ignore this signal completely. In that case
// the process will keep running unless you call ForceStop()
// Returns an error in case there's any.
func (proc *Proc) GracefullyStop() error {
	if proc.process != nil {
		err := proc.process.Signal(syscall.SIGTERM)
		proc.Status.SetStatus("asked to stop")
		return err
	}
	return errors.New("Process does not exist.")
}

// Stop can terminate a process depending on how it was started. If setpgid as used, we need to kill the processes using -1*PID; all other processes can be terminated normally.
func (proc *Proc) Stop() error {
	if proc.process != nil {
		if proc.Sys != nil && proc.Sys.Setpgid == true {
			// We use Setpgid to allow snort processes to run even if matra exists or restarts:
			// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
			// Internally this will cause Go to call setpgid(2) between fork(2) and execve(2), to assign the child process a new PGID identical to its PID. This allows us to kill all processes in the process group by sending a KILL to -PID of the process, which is the same as -PGID. Assuming that the child process did not use setpgid(2) when spawning its own child, this should kill the child along with all of its children on any *Nix systems.
			syscall.Kill(-proc.GetPid(), syscall.SIGKILL)
			proc.Status.SetStatus("killed (via setpgid)")
			return nil
		}
		err := proc.GracefullyStop()
		if err != nil {
			return err
		}

	}
	return errors.New("Process does not exist.")
}

// Restart will try to gracefully stop the process and then Start it again.
// Returns an error in case there's any.
func (proc *Proc) Restart() error {
	if proc.IsAlive() {
		err := proc.Stop()
		if err != nil {
			return err
		}
	}
	return proc.Start()
}

// Delete will delete everything created by this process, including the out, err and pid file.
// Returns an error in case there's any.
func (proc *Proc) Delete() error {
	proc.release()
	err := common.DeleteFile(proc.Infile)
	if err != nil {
		return err
	}
	err = common.DeleteFile(proc.Outfile)
	if err != nil {
		return err
	}
	err = common.DeleteFile(proc.Errfile)
	if err != nil {
		return err
	}
	return nil //os.RemoveAll(proc.Path)
}

// IsAlive will check if the process is alive or not.
func (proc *Proc) IsAlive() bool {
	dir := fmt.Sprintf("/proc/%d", proc.GetPid())
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		logger.Errorf("Error while checking process state: %s", err)
		return false
	}
	return true
}

// IsAlive will check if the process is alive or not.
// Returns true if the process is alive or false otherwise.
// func (proc *Proc) IsAlive() bool {
// 	p, err := os.FindProcess(proc.Pid) // this is useles on Unix since it always returns a proc -- see source of this func.
// 	if err != nil {
// 		return false
// 	}
// 	return p.Signal(syscall.Signal(0)) == nil
// }

// Watch will stop execution and wait until the process change its state. Usually changing state, means that the process died.
// Returns a tuple with the new process state and an error in case there's any.
func (proc *Proc) Watch() *watcherProcStatus {
	state, err := proc.process.Wait()
	return &watcherProcStatus{state.String(), err}
}

// WatchNonChild uses a simple poll and sleep pattern to monitor a process
// https://stackoverflow.com/questions/1157700/how-to-wait-for-exit-of-non-children-processes
func (proc *Proc) WatchNonChild() *watcherProcStatus {
	for {
		if !proc.IsAlive() {
			return &watcherProcStatus{state: "existed (non-child)"}
		}
		time.Sleep(1 * time.Second)
	}
}

// Will release the process and remove its PID file
func (proc *Proc) release() {
	if proc.process != nil {
		proc.process.Release()
	}
	common.DeleteFile(proc.Pidfile)
}

// Add one restart to proc status
func (proc *Proc) AddRestart() {
	proc.Status.IncrRestart()
}

// Return proc current PID
func (proc *Proc) GetPid() int {
	return proc.Pid
}

// Return proc current status
func (proc *Proc) GetStatus() *ProcStatus {
	return proc.Status
}

// Set proc status
func (proc *Proc) SetStatus(status string) {
	proc.Status.SetStatus(status)
}

// Proc identifier that will be used by watcher to keep track of its processes
func (proc *Proc) Identifier() string {
	return proc.Name
}

// Returns true if the process should be kept alive or not
func (proc *Proc) ShouldKeepAlive() bool {
	return proc.KeepAlive
}

// watcherProcStatus is a wrapper with the process state and an error in case there's any.
type watcherProcStatus struct {
	state string
	err   error
}

// ProcWatcher is a wrapper that act as a object that watches a process.
type ProcWatcher struct {
	procStatus  chan *watcherProcStatus
	proc        ProcContainer
	stopWatcher chan bool
}

// Watcher is responsible for watching a list of processes and report to
// Master in case the process dies at some point.
type Watcher struct {
	sync.Mutex
	restartProc chan ProcContainer
	watchProcs  map[string]*ProcWatcher
}

// InitWatcher will create a Watcher instance.
// Returns a Watcher instance.
func InitWatcher() *Watcher {
	watcher := &Watcher{
		restartProc: make(chan ProcContainer),
		watchProcs:  make(map[string]*ProcWatcher),
	}
	return watcher
}

// RestartProc is a wrapper to export the channel restartProc. It basically keeps track of
// all the processes that died and need to be restarted.
// Returns a channel with the dead processes that need to be restarted.
func (watcher *Watcher) RestartProc() chan ProcContainer {
	return watcher.restartProc
}

// AddProcWatcher will add a watcher on proc.
// If notChild is true, we cannot use the proc.Wait() method since it
// can only wait on child processes. The notChild approach uses polling.
// It's required in the case of restore, where a process (e.g. snort) was
// started by an earlier instance of matra and therefore is not a child
// of the current instance.
func (watcher *Watcher) AddProcWatcher(proc ProcContainer, notChild bool) {
	watcher.Lock()
	defer watcher.Unlock()
	if _, ok := watcher.watchProcs[proc.Identifier()]; ok {
		logger.Debugf("A watcher for this process already exists.")
		return
	}
	procWatcher := &ProcWatcher{
		procStatus:  make(chan *watcherProcStatus, 1),
		proc:        proc,
		stopWatcher: make(chan bool, 1),
	}
	watcher.watchProcs[proc.Identifier()] = procWatcher
	go func() {
		var status *watcherProcStatus
		if notChild {
			logger.Debugf("Starting watcher on (non-child) proc %s", proc.Identifier())
			status = proc.WatchNonChild()
		} else {
			logger.Debugf("Starting watcher on proc %s", proc.Identifier())
			status = proc.Watch()
		}
		procWatcher.procStatus <- status
	}()
	go func() {
		defer delete(watcher.watchProcs, procWatcher.proc.Identifier())
		select {
		case procStatus := <-procWatcher.procStatus:
			logger.Debugf("Proc %s is dead, advising...", procWatcher.proc.Identifier())
			logger.Debugf("State is %s", procStatus.state)
			watcher.restartProc <- procWatcher.proc
			break
		case <-procWatcher.stopWatcher:
			break
		}
	}()
}

// StopWatcher will stop a running watcher on a process with 'identifier'
// Returns a channel that will be populated when the watcher is finally done.
func (watcher *Watcher) StopWatcher(identifier string) chan bool {
	if watcher, ok := watcher.watchProcs[identifier]; ok {
		logger.Debugf("Stopping watcher on proc %s", identifier)
		watcher.stopWatcher <- true
		waitStop := make(chan bool, 1)
		go func() {
			<-watcher.procStatus
			waitStop <- true
		}()
		return waitStop
	}
	return nil
}

// ProcMaster is the main module that keeps everything in place and execute
// the necessary actions to keep the process running as they should be.
type ProcMaster struct {
	sync.Mutex

	Watcher *Watcher // Watcher is a watcher instance.

	Procs map[string]ProcContainer // Procs is a map containing all procs started on APM.
}

var master *ProcMaster
var once sync.Once

func GetProcMaster() *ProcMaster {
	once.Do(func() {
		master = &ProcMaster{
			Watcher: InitWatcher(),
			Procs:   make(map[string]ProcContainer),
		}
		go master.WatchProcs()
	})
	return master
}

// WatchProcs will keep the procs running forever.
func (master *ProcMaster) WatchProcs() {
	for proc := range master.Watcher.RestartProc() {
		if !proc.ShouldKeepAlive() {
			master.Lock()
			master.updateStatus(proc)
			master.Unlock()
			logger.Debugf("Proc %s does not have keep alive set. Will not be restarted.", proc.Identifier())
			continue
		}
		logger.Debugf("Restarting proc %s.", proc.Identifier())
		if proc.IsAlive() {
			logger.Warningf("Proc %s was supposed to be dead, but it is alive.", proc.Identifier())
		}
		master.Lock()
		proc.AddRestart()
		err := master.restart(proc)
		master.Unlock()
		if err != nil {
			logger.Warningf("Could not restart process %s due to %s.", proc.Identifier(), err)
		}
	}
}

func (master *ProcMaster) updateStatus(proc ProcContainer) {
	if proc.IsAlive() {
		proc.SetStatus("running")
	} else {
		proc.SetStatus("stopped")
	}
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *ProcMaster) restart(proc ProcContainer) error {
	err := master.stop(proc)
	if err != nil {
		return err
	}
	return master.start(proc)
}

// StopProcess will stop a process with the given name.
func (master *ProcMaster) StopProcess(name string) error {
	master.Lock()
	defer master.Unlock()
	if proc, ok := master.Procs[name]; ok {
		err := master.stop(proc)
		if err != nil {
			return err
		}
		delete(master.Procs, proc.Identifier())
		return nil
	}
	return errors.New("Unknown process.")
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *ProcMaster) stop(proc ProcContainer) error {
	if proc.IsAlive() {
		waitStop := master.Watcher.StopWatcher(proc.Identifier())
		err := proc.Stop()
		if err != nil {
			return err
		}
		if waitStop != nil {
			<-waitStop
			proc.SetStatus("stopped")
		}
		err = proc.Delete()
		if err != nil {
			logger.Warning("Error while deleting %s (%d) process files: %s", proc.Identifier(), proc.GetPid(), err)
			// We don't return after this error. It's not super critical.
		}
	} else {
		// If the process is not longer active, we can remove any files it has sitting around.
		err := proc.Delete()
		if err != nil {
			logger.Warning("Error while deleting %s (%d) process files: %s", proc.Identifier(), proc.GetPid(), err)
		}
	}
	logger.Infof("Proc %s successfully stopped.", proc.Identifier())
	return nil
}

func (master *ProcMaster) StartProcess(proc ProcContainer) error {
	master.Lock()
	defer master.Unlock()
	if _, ok := master.Procs[proc.Identifier()]; ok {
		logger.Warningf("Proc %s already exist.", proc.Identifier())
		return errors.New("Trying to start a process that already exist.")
	}

	err := master.start(proc)
	if err != nil {
		return err
	}
	master.Procs[proc.Identifier()] = proc
	return nil
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *ProcMaster) start(proc ProcContainer) error {
	if !proc.IsAlive() {
		err := proc.Start()
		if err != nil {
			return err
		}
		master.Watcher.AddProcWatcher(proc, false)
		proc.SetStatus("running")
	}
	return nil
}

// Restore will add an already running proc to the master's internal map.
func (master *ProcMaster) Restore(proc ProcContainer) error {
	master.Lock()
	defer master.Unlock()
	if _, ok := master.Procs[proc.Identifier()]; ok {
		logger.Warningf("Proc %s already exist.", proc.Identifier())
		return errors.New("Trying to start a process that already exist.")
	}

	err := master.restore(proc)
	if err != nil {
		return err
	}
	master.Procs[proc.Identifier()] = proc
	return nil
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *ProcMaster) restore(proc ProcContainer) error {
	if proc.IsAlive() {
		master.Watcher.AddProcWatcher(proc, true /*non-child*/)
		proc.SetStatus("restored")
		return nil
	}
	return errors.New(fmt.Sprintf("Restore called on %s but process not running.", proc.Identifier()))
}
