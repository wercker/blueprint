package main

import (
	"os"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()

	app.Name = "Blueprint"
	app.Copyright = "(c) 1996 Wercker Holding BV"
	app.Usage = "TiVo for VRML"

	//app.Version = version.Version
	//app.Compiled = version.CompiledAt

	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		//clientCommand,
		gatewayCommand,
		serverCommand,
	}

	app.Run(os.Args)
}

type globalOptions struct{}

func parseGlobalOptions(c *cli.Context) (*globalOptions, error) {
	return &globalOptions{}, nil
}

var errorExitCode = cli.NewExitError("", 1)
