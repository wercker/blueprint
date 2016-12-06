package main

import "gopkg.in/urfave/cli.v1"

var (
	// errorExitCode returns a urfave decorated error which indicates a exit
	// code 1. To be return from a urfave action.
	errorExitCode = cli.NewExitError("", 1)
)

// validPortNumber returns true if port is between 0 and 65535.
func validPortNumber(port int) bool {
	return port > 0 && port < 65535
}
