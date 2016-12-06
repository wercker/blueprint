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

	app.Version = Version()
	app.Compiled = CompiledAt()

	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		//clientCommand,
		gatewayCommand,
		serverCommand,
	}

	app.Run(os.Args)
}
