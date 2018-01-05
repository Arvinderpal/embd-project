package main

import (
	"os"

	"github.com/Arvinderpal/embd-project/common"
	daemon "github.com/Arvinderpal/embd-project/daemon"

	"github.com/codegangsta/cli"
	l "github.com/op/go-logging"
)

var (
	log = l.MustGetLogger("segue-cli")
)

func newSegueApp() *cli.App {
	app := cli.NewApp()
	app.Name = "segue"
	app.Usage = "segue"
	app.Version = common.Version
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D", // NOTE:
			Usage: "Enable debug messages",
		},
		cli.StringFlag{
			Name:  "host, H",
			Usage: "Daemon host to connect to",
		},
	}
	app.Commands = []cli.Command{
		daemon.CliCommand,
	}
	return app
}

func main() {
	app := newSegueApp()
	app.Before = initEnv
	app.Run(os.Args)
}

func initEnv(ctx *cli.Context) error {

	if ctx.Bool("debug") {
		common.SetupLOG(log, "DEBUG", nil)
	} else {
		common.SetupLOG(log, "INFO", nil)
	}
	return nil
}
