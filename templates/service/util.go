package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/wercker/pkg/conf"
	"github.com/wercker/pkg/trace"
	"gopkg.in/urfave/cli.v1"
)

var (
	// errorExitCode returns a urfave decorated error which indicates a exit
	// code 1. To be returned from a urfave action.
	errorExitCode = cli.NewExitError("", 1)
)

// validPortNumber returns true if port is between 0 and 65535.
func validPortNumber(port int) bool {
	return port > 0 && port < 65535
}

func getTracer(o *conf.TraceOptions, serviceName string, port int) (opentracing.Tracer, error) {
	if o.Trace {
		return trace.NewZipkinTracer(o.TraceEndpoint, serviceName, port)
	}

	return trace.NewNoopTracer()
}
