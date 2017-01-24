package main

import (
	"os"

	"github.com/wercker/pkg/log"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()

	app.Name = "Blueprint"
	app.Copyright = "(c) 1996 Wercker Holding BV"
	app.Usage = "TiVo for VRML"

	app.Version = Version()
	app.Compiled = CompiledAt()
	app.Before = log.SetupLogging
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug logging",
		},
	}
	app.Commands = []cli.Command{
		//clientCommand,
		gatewayCommand,
		serverCommand,
	}

	app.Run(os.Args)
}
