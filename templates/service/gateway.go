package main

import (
	"fmt"
	"net/http"

	"gopkg.in/urfave/cli.v1"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/wercker/auth/middleware"
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/pkg/conf"
	"github.com/wercker/pkg/log"
	"github.com/wercker/pkg/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var gatewayCommand = cli.Command{
	Name:   "gateway",
	Usage:  "Start gRPC gateway",
	Action: gatewayAction,
	Flags:  append(gatewayFlags, conf.TraceFlags()...),
}

var gatewayFlags = []cli.Flag{
	cli.IntFlag{
		Name:   "port",
		Value:  667,
		EnvVar: "HTTP_PORT",
	},
	cli.StringFlag{
		Name:   "host",
		Value:  "localhost:666",
		EnvVar: "GRPC_HOST",
	},
}

var gatewayAction = func(c *cli.Context) error {
	log.Info("Starting blueprint gateway")

	log.Debug("Parsing gateway options")
	o, err := parseGatewayOptions(c)
	if err != nil {
		log.WithError(err).Error("Unable to validate arguments")
		return errorExitCode
	}

	tracer, err := getTracer(o.TraceOptions, "blueprint-gw", o.Port)
	if err != nil {
		log.WithError(err).Error("Unable to create a tracer")
		return errorExitCode
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux() // grpc-gateway

	// The following handlers will be called in reversed order (ie. bottom to top)
	var handler http.Handler
	handler = middleware.AuthTokenMiddleware(mux)   // authentication middleware
	handler = trace.HTTPMiddleware(handler, tracer) // opentracing + expose trace ID

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)), // opentracing (outgoing)
	}

	err = core.RegisterBlueprintHandlerFromEndpoint(ctx, mux, o.Host, opts)
	if err != nil {
		log.WithError(err).Error("Unable to register handler from Endpoint")
		return errorExitCode
	}

	log.Printf("Listening on port %d", o.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", o.Port), handler)

	return nil
}

type gatewayOptions struct {
	*conf.TraceOptions

	Port int
	Host string
}

func parseGatewayOptions(c *cli.Context) (*gatewayOptions, error) {
	traceOptions := conf.ParseTraceOptions(c)

	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	return &gatewayOptions{
		TraceOptions: traceOptions,

		Port: port,
		Host: c.String("host"),
	}, nil
}
