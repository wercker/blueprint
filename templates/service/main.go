package main

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()

	app.Name = "Blueprint"
	app.Copyright = "(c) 1996 Wercker Holding BV"
	app.Usage = "Tivo for VRML"

	//app.Version = version.Version
	//app.Compiled = version.CompiledAt

	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		//clientCommand,
		//gatewayCommand,
		serverCommand,
	}

	app.Run(os.Args)
}

type GlobalOptions struct{}

func ParseGlobalOptions(c *cli.Context) (*GlobalOptions, error) {
	return &GlobalOptions{}, nil
}

var ErrorExitCode = cli.NewExitError("", 1)
