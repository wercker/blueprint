//-----------------------------------------------------------------------------
// Copyright (c) 2017 Oracle and/or its affiliates.  All rights reserved.
// This program is free software: you can modify it and/or redistribute it
// under the terms of:
//
// (i)  the Universal Permissive License v 1.0 or at your option, any
//      later version (http://oss.oracle.com/licenses/upl); and/or
//
// (ii) the Apache License v 2.0. (http://www.apache.org/licenses/LICENSE-2.0)
//-----------------------------------------------------------------------------

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
