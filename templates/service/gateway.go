package main

import (
	"fmt"
	"net/http"

	"gopkg.in/urfave/cli.v1"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/wercker/auth/middleware"
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/pkg/log"
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
	err = core.RegisterBlueprintHandlerFromEndpoint(ctx, mux, o.Host, opts)
	if err != nil {
		log.WithError(err).Error("Unable to register handler from Endpoint")
		return errorExitCode
	}

	authMiddleware := middleware.AuthTokenMiddleware(mux)

	log.Printf("Listening on port %d", o.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", o.Port), authMiddleware)

	return nil
}

func parseGatewayOptions(c *cli.Context) (*gatewayOptions, error) {
	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	return &gatewayOptions{
		Port: port,
		Host: c.String("host"),
	}, nil
}

type gatewayOptions struct {
	Port int
	Host string
}
