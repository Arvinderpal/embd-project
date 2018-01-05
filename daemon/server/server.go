package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"

	"github.com/Arvinderpal/embd-project/common"
	"github.com/Arvinderpal/embd-project/daemon/daemon"

	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("segue-server")
	format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)
)

// Server listens for HTTP requests and sends them to our router.
type Server interface {
	Start() error
	Stop() error
}

type server struct {
	listener   net.Listener
	socketPath string
	router     Router
}

// NewServer returns a new Server that listens for requests in socketPath and
// sends them to daemon.
func NewServer(socketPath string, daemon *daemon.Daemon) (Server, error) {
	socketDir := path.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create '%s' directory: %s", socketDir, err)
	}

	if err := os.Remove(socketPath); !os.IsNotExist(err) && err != nil {
		return nil, fmt.Errorf("failed to remove older listener socket: %s", err)
	}

	router := NewRouter(daemon)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create listen socket: %s", err)
	}
	if os.Getuid() == 0 {
		gid, err := common.GetGroupIDByName(common.SegueGroupName)
		if err == nil {
			if err := os.Chown(socketPath, 0, gid); err != nil {
				return nil, fmt.Errorf("failed while setting up %s's group ID in %q: %s", common.SegueGroupName, socketPath, err)
			}
		} else {
			logger.Warningf("Group %s not found: %s", common.SegueGroupName, err)
		}
		if err := os.Chmod(socketPath, 0660); err != nil {
			return nil, fmt.Errorf("failed while setting up %s's file permissions in %q: %s", common.SegueGroupName, socketPath, err)
		}
	}

	return server{listener, socketPath, router}, nil
}

// Start starts the server and blocks to server HTTP requests.
func (d server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	srv := &http.Server{Handler: d.router}
	go func() {
		logger.Infof("Listening on %q", d.socketPath)
		if err := srv.Serve(d.listener); err != nil {
			logger.Fatal(err)
		}
	}()

	<-stop
	logger.Infof("Shutting down the server...")
	if err := srv.Shutdown(context.Background()); err != nil {
		return err
	}
	logger.Infof("Server gracefully stopped.")
	return nil
	// return http.Serve(d.listener, d.router)
}

// Stop stops the HTTP listener.
func (d server) Stop() error {
	return d.listener.Close()
}
