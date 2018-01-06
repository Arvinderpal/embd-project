package daemon

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"

	common "github.com/Arvinderpal/embd-project/common"
	mclient "github.com/Arvinderpal/embd-project/common/client"
	"github.com/Arvinderpal/embd-project/daemon/daemon"
	s "github.com/Arvinderpal/embd-project/daemon/server"
	"github.com/Arvinderpal/embd-project/pkg/machine"
	"github.com/Arvinderpal/embd-project/pkg/option"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

var (
	config = daemon.NewConfig()

	// Arguments variables keep in alphabetical order
	socketPath string
	nodeAddr   string

	log = logging.MustGetLogger("segue-daemon")

	// CliCommand is the command that will be used in main program.
	CliCommand cli.Command
)

func init() {
	CliCommand = cli.Command{
		Name: "daemon",
		// Keep Destination alphabetical order
		Subcommands: []cli.Command{
			{
				Name:   "run",
				Usage:  "Run the daemon",
				Before: initEnv,
				Action: run,
				Flags: []cli.Flag{
					cli.StringFlag{
						Destination: &config.LibDir,
						Name:        "D",
						Value:       common.DefaultLibDir,
						Usage:       "library directory",
					},
					cli.StringFlag{
						Destination: &config.RunDir,
						Name:        "R",
						Value:       common.SeguePath,
						Usage:       "Runtime data directory",
					},
					cli.StringFlag{
						Destination: &socketPath,
						Name:        "s",
						Value:       common.SegueSock,
						Usage:       "Sets the socket path to listen for connections",
					},
					cli.StringFlag{
						Destination: &nodeAddr,
						Name:        "n, node-address",
						Value:       "123.123.123.123",
						Usage:       "IPv4 address of node, must be in correct format",
					},
					cli.BoolFlag{
						Name:  "debug",
						Usage: "Enable debug messages",
					},
					cli.BoolFlag{
						Name:  "dr, disable-restore",
						Usage: "Disable restore at startup",
					},
				},
			},
			{
				Name:      "config",
				Usage:     "Manage daemon configuration",
				Action:    configDaemon,
				ArgsUsage: "[<option>=(enable|disable) ...]",
			},
			{
				Name:   "status",
				Usage:  "Returns the daemon current status",
				Action: statusDaemon,
			},
			{
				Name:      "machine",
				Usage:     "Join,Leave,Get,Get-All Machines",
				Action:    machineUpdate,
				ArgsUsage: "<join JSON-file>/<leave/get machine-id>/<get-all>",
			},
			{
				Name:      "adaptor",
				Usage:     "Attach and Detach Adaptor",
				Action:    adaptorUpdate,
				ArgsUsage: "<attach JSON file>/<detach machine-id adaptor-type adaptor-id>",
			},
			{
				Name:      "driver",
				Usage:     "Start and Stop Driver",
				Action:    driverUpdate,
				ArgsUsage: "<start JSON file>/<stop machine-id driver-type driver-id>",
			},
		},
	}
}

func statusDaemon(ctx *cli.Context) {

	var (
		client *mclient.Client
		err    error
	)
	if host := ctx.GlobalString("host"); host == "" {
		client, err = mclient.NewDefaultClient()
	} else {
		client, err = mclient.NewClient(host, nil)
	}
	if err != nil {
		log.Errorf("Error while creating client: %s\n", err)
		return
	}

	if sr, err := client.GlobalStatus(); err != nil {
		log.Errorf("Status: ERROR - Unable to reach out daemon: %s\n", err)
		return
	} else {
		// w := tabwriter.NewWriter(ctx.App.Writer, 2, 0, 3, ' ', 0)
		// fmt.Fprintf(w, "Status:\t%s\n", sr)
		// w.Flush()
		log.Infof("Status: %s\n", sr)
		return
	}

}

func configDaemon(ctx *cli.Context) {

	var (
		client *mclient.Client
		err    error
	)

	first := ctx.Args().First()

	if first == "list" {
		for k, s := range daemon.DaemonOptionLibrary {
			log.Infof("%-24s %s\n", k, s.Description)
		}
		return
	}

	if host := ctx.GlobalString("host"); host == "" {
		client, err = mclient.NewDefaultClient()
	} else {
		client, err = mclient.NewClient(host, nil)
	}

	if err != nil {
		log.Errorf("Error while creating segue-client: %s\n", err)
		return
	}

	res, err := client.Ping()
	if err != nil {
		log.Errorf("Unable to reach daemon: %s\n", err)
		return
	}

	if res == nil {
		log.Errorf("Empty response from daemon\n")
		return
	}

	opts := ctx.Args()

	if len(opts) == 0 {
		res.Opts.Dump()
		return
	}

	dOpts := make(option.OptionMap, len(opts))

	for k := range opts {
		name, value, err := option.ParseOption(opts[k], &daemon.DaemonOptionLibrary)
		if err != nil {
			log.Errorf("%s\n", err)
			return
		}

		dOpts[name] = value

		err = client.Update(dOpts)
		if err != nil {
			log.Errorf("Unable to update daemon: %s\n", err)
			return
		}
	}
}

func machineUpdate(ctx *cli.Context) {
	var (
		client *mclient.Client
		err    error
	)

	if host := ctx.GlobalString("host"); host == "" {
		client, err = mclient.NewDefaultClient()
	} else {
		client, err = mclient.NewClient(host, nil)
	}

	if err != nil {
		log.Errorf("Error while creating segue-client: %s\n", err)
		return
	}

	args := ctx.Args()
	if len(args) < 1 {
		log.Errorf("Insufficient arguments provided.\n")
		return
	}

	operation := ctx.Args().Get(0)

	switch strings.ToLower(operation) {
	case "join":
		args := ctx.Args()
		if len(args) < 2 {
			log.Errorf("Insufficient arguments provided to %s.\n", operation)
			return
		}
		dataPathname := ctx.Args().Get(1)
		var mh machine.Machine
		mhFile, err := os.Open(dataPathname)
		defer mhFile.Close()
		if err != nil {
			log.Errorf("Error opening machine file: %s", err.Error())
			return
		}
		jsonParser := json.NewDecoder(mhFile)
		if err = jsonParser.Decode(&mh); err != nil {
			log.Errorf("Error parsing machine file: %s", err.Error())
			return
		}

		err = client.MachineJoin(mh)
		if err != nil {
			log.Errorf("Unable to %s machine {%s}: %s.\n", operation, mh.MachineID, err)
			return
		}
	case "leave":
		args := ctx.Args()
		if len(args) < 2 {
			log.Errorf("Insufficient arguments provided to %s.\n", operation)
			return
		}
		machineID := ctx.Args().Get(1)
		err = client.MachineLeave(machineID)
		if err != nil {
			log.Errorf("Unable to %s machine {%s}: %s\n", operation, machineID, err)
			return
		}
	case "get":
		args := ctx.Args()
		if len(args) < 2 {
			log.Errorf("Insufficient arguments provided to %s.\n", operation)
			return
		}
		machineID := ctx.Args().Get(1)
		mh, err := client.MachineGet(machineID)
		if err != nil {
			log.Errorf("Unable to %s machine {%s}: %s\n", operation, machineID, err)
			return
		}
		if mh != nil {
			log.Infof("%s\n", mh.PrettyPrint())
			// if mh.Status != nil {
			// 	log.Infof("Machine status: \n%s\n", mh.Status.DumpLog())
			// }
		}
	case "get-all":
		mhs, err := client.MachinesGet()
		if err != nil {
			log.Errorf("Unable to %s machines: %s\n", operation, err)
			return
		}
		for _, mh := range mhs {
			log.Infof("%+v\n", mh)
		}
	default:
		log.Errorf("Unknown Operation %s", operation)
		return
	}
}

func driverUpdate(ctx *cli.Context) {

	var (
		client *mclient.Client
		err    error
	)

	if host := ctx.GlobalString("host"); host == "" {
		client, err = mclient.NewDefaultClient()
	} else {
		client, err = mclient.NewClient(host, nil)
	}

	if err != nil {
		log.Errorf("Error while creating segue-client: %s\n", err)
		return
	}

	args := ctx.Args()
	operation := ctx.Args().Get(0)

	switch strings.ToLower(operation) {
	case "start":
		if len(args) < 2 {
			log.Errorf("Insufficient arguments provided\n")
			return
		}
		confPath := ctx.Args().Get(1)
		bytes, err := ioutil.ReadFile(confPath)
		if err != nil {
			log.Errorf("Error reading driver file: %s", err.Error())
			return
		}
		err = client.StartDrivers(bytes)
		if err != nil {
			log.Errorf("Unable to %s driver(s) on machine: %s.\n", operation, err)
			return
		}
	case "stop":
		if len(args) < 4 {
			log.Errorf("Insufficient arguments provided\n")
			return
		}
		machineID := ctx.Args().Get(1)
		driverType := ctx.Args().Get(2)
		driverID := ctx.Args().Get(3)
		err = client.StopDriver(machineID, driverType, driverID)
		if err != nil {
			log.Errorf("Unable to %s driver %s on machine %s: %s.\n", operation, driverID, machineID, err)
			return
		}
	default:
		log.Errorf("Unknown Operation %s", operation)
		return
	}
}

func adaptorUpdate(ctx *cli.Context) {

	var (
		client *mclient.Client
		err    error
	)

	if host := ctx.GlobalString("host"); host == "" {
		client, err = mclient.NewDefaultClient()
	} else {
		client, err = mclient.NewClient(host, nil)
	}

	if err != nil {
		log.Errorf("Error while creating segue-client: %s\n", err)
		return
	}

	args := ctx.Args()
	operation := ctx.Args().Get(0)

	switch strings.ToLower(operation) {
	case "attach":
		if len(args) < 2 {
			log.Errorf("Insufficient arguments provided\n")
			return
		}
		confPath := ctx.Args().Get(1)
		bytes, err := ioutil.ReadFile(confPath)
		if err != nil {
			log.Errorf("Error reading adaptor file: %s", err.Error())
			return
		}
		err = client.AttachAdaptors(bytes)
		if err != nil {
			log.Errorf("Unable to %s adaptor(s) on machine: %s.\n", operation, err)
			return
		}
	case "detach":
		if len(args) < 4 {
			log.Errorf("Insufficient arguments provided\n")
			return
		}
		machineID := ctx.Args().Get(1)
		adaptorType := ctx.Args().Get(2)
		adaptorID := ctx.Args().Get(3)
		err = client.DetachAdaptor(machineID, adaptorType, adaptorID)
		if err != nil {
			log.Errorf("Unable to %s adaptor %s on machine %s: %s.\n", operation, adaptorID, machineID, err)
			return
		}
	default:
		log.Errorf("Unknown Operation %s", operation)
		return
	}
}

func setupLogger(ctx *cli.Context) {
	config.OptsMU.RLock()
	defer config.OptsMU.RUnlock()

	var logWriter io.Writer
	if config.Opts.Opts[common.OptionTestMode] {
		logWriter = ctx.App.Writer
		if logWriter == nil {
			log.Warningf("Test Mode enabled without log writer. Default log writer (stderr) will be used.")
		}
		log.Info("Test mode is enabled.")
	} else {
		logWriter = os.Stderr
	}

	if config.Opts.Opts[common.OptionDebug] {
		common.SetupLOG(log, "DEBUG", logWriter)
		log.Info("Debuging is enabled")
	} else {
		common.SetupLOG(log, "INFO", logWriter)
	}
}

func initEnv(ctx *cli.Context) error {
	config.OptsMU.Lock()
	defer config.OptsMU.Unlock()

	if ctx.GlobalBool("testmode") {
		config.Opts.Set(common.OptionTestMode, true)
	}

	if ctx.GlobalBool("debug") {
		config.Opts.Set(common.OptionDebug, true)
	}

	return nil
}

func run(ctx *cli.Context) {

	setupLogger(ctx)

	if ctx.Bool("disable-restore") {
		config.Opts.Set(common.OptionDisableRestore, true)
	}

	ipaddr := net.ParseIP(nodeAddr)
	if ipaddr == nil {
		log.Errorf("Unable to parse node address %s", nodeAddr)
	}
	config.NodeAddress = nodeAddr

	d, err := daemon.NewDaemon(config)
	if err != nil {
		log.Errorf("Error while creating daemon: %s", err)
		return
	}

	server, err := s.NewServer(socketPath, d)
	if err != nil {
		log.Errorf("Error while creating daemon: %s", err)
	}
	defer server.Stop()
	server.Start()
}
