package main

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/urfave/cli.v1"

	log "github.com/Sirupsen/logrus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	othttp "github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/wercker/auth/middleware"
	"github.com/wercker/blueprint/templates/service/core"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var gatewayCommand = cli.Command{
	Name:   "gateway",
	Usage:  "Start gRPC gateway",
	Action: gatewayAction,
	Flags: []cli.Flag{
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
		cli.BoolFlag{
			Name:   "trace",
			EnvVar: "TRACING_ENABLED",
			Usage:  "Enable tracing",
		},
		cli.StringFlag{
			Name:   "trace-endpoint",
			EnvVar: "TRACING_ENDPOINT",
			Usage:  "Endpoint for the tracing data",
		},
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

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	handler := middleware.AuthTokenMiddleware(mux)

	if o.Trace {
		log.Info("Tracing is enabled")
		tracer, err := getTracer(o.TraceEndpoint, "blueprint-gw", fmt.Sprintf(":%d", o.Port), false, false)
		if err != nil {
			log.WithError(err).Error("Unable to create a tracer")
			return errorExitCode
		}
		handler = othttp.Middleware(tracer, handler)
		opts = append(opts, grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(tracer)))
	}

	err = core.RegisterBlueprintHandlerFromEndpoint(ctx, mux, o.Host, opts)
	if err != nil {
		log.WithError(err).Error("Unable to register handler from Endpoint")
		return errorExitCode
	}

	authMiddleware := middleware.AuthTokenMiddleware(mux)

	log.Printf("Listening on port %d", o.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", o.Port), handler)

	return nil
}

func parseGatewayOptions(c *cli.Context) (*gatewayOptions, error) {
	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	trace := c.Bool("trace")
	traceEndpoint := c.String("trace-endpoint")
	if trace && traceEndpoint == "" {
		return nil, errors.New("Trace endpoint is required")
	}

	return &gatewayOptions{
		Port:          port,
		Host:          c.String("host"),
		Trace:         trace,
		TraceEndpoint: traceEndpoint,
	}, nil
}

type gatewayOptions struct {
	Port          int
	Host          string
	Trace         bool
	TraceEndpoint string
}
