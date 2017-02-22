package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/mgo.v2"
	"gopkg.in/urfave/cli.v1"

	grpcmw "github.com/mwitkow/go-grpc-middleware"
	"github.com/pkg/errors"
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/blueprint/templates/service/server"
	"github.com/wercker/blueprint/templates/service/state"
	"github.com/wercker/pkg/conf"
	"github.com/wercker/pkg/log"
	"github.com/wercker/pkg/trace"
	"google.golang.org/grpc"
)

var serverCommand = cli.Command{
	Name:   "server",
	Usage:  "start gRPC server",
	Action: serverAction,
	Flags:  append(serverFlags, conf.TraceFlags()...),
}

var serverFlags = []cli.Flag{
	cli.IntFlag{
		Name:   "port",
		Value:  666,
		EnvVar: "PORT",
	},
	cli.StringFlag{
		Name:   "mongo",
		Value:  "mongodb://localhost:27017",
		EnvVar: "MONGODB_URI",
	},
	cli.StringFlag{
		Name:  "mongo-database",
		Value: "blueprint",
	},
	cli.StringFlag{
		Name:  "state-store",
		Usage: "storage driver, currently supported [mongo]",
		Value: "mongo",
	},
}

var serverAction = func(c *cli.Context) error {
	log.Info("Starting blueprint server")

	log.Debug("Parsing server options")
	o, err := parseServerOptions(c)
	if err != nil {
		log.WithError(err).Error("Unable to validate arguments")
		return errorExitCode
	}

	tracer, err := getTracer(o.TraceOptions, "blueprint", o.Port)
	if err != nil {
		log.WithError(err).Error("Unable to create a tracer")
		return errorExitCode
	}

	store, err := getStore(o)
	if err != nil {
		log.WithError(err).Error("Unable to create state store")
		return errorExitCode
	}
	defer store.Close()

	store = state.NewTraceStore(store, tracer)

	log.Debug("Creating server")
	server, err := server.New(store)
	if err != nil {
		log.WithError(err).Error("Unable to create server")
		return errorExitCode
	}

	// The following interceptors will be called in order (ie. top to bottom)
	interceptors := []grpc.UnaryServerInterceptor{
		trace.Interceptor(tracer), // opentracing + expose trace ID
	}

	s := grpc.NewServer(grpcmw.WithUnaryServerChain(interceptors...))
	core.RegisterBlueprintServer(s, server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
	if err != nil {
		log.WithField("port", o.Port).WithError(err).Error("Failed to listen")
		return errorExitCode
	}

	errc := make(chan error, 2)

	// Shutdown on SIGINT, SIGTERM
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Start gRPC server in separate goroutine
	go func() {
		log.WithField("port", o.Port).Info("Starting server")
		errc <- s.Serve(lis)
	}()

	err = <-errc
	log.WithError(err).Info("Shutting down")

	// Graceful shutdown the gRPC server
	s.GracefulStop()

	return nil
}

type serverOptions struct {
	*conf.TraceOptions

	MongoDatabase string
	MongoURI      string
	Port          int
	StateStore    string
}

func parseServerOptions(c *cli.Context) (*serverOptions, error) {
	traceOptions := conf.ParseTraceOptions(c)

	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	return &serverOptions{
		TraceOptions: traceOptions,

		MongoDatabase: c.String("mongo-database"),
		MongoURI:      c.String("mongo"),
		Port:          port,
		StateStore:    c.String("state-store"),
	}, nil
}

func getStore(o *serverOptions) (state.Store, error) {
	switch o.StateStore {
	case "mongo":
		return getMongoStore(o)
	default:
		return nil, fmt.Errorf("Invalid store: %s", o.StateStore)
	}
}

func getMongoStore(o *serverOptions) (*state.MongoStore, error) {
	log.Info("Creating MongoDB store")

	log.WithField("MongoURI", o.MongoURI).Debug("Dialing the MongoDB cluster")
	session, err := mgo.Dial(o.MongoURI)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing the MongoDB cluster failed")
	}

	log.WithField("MongoDatabase", o.MongoDatabase).Debug("Creating MongoDB store")
	store, err := state.NewMongoStore(session, o.MongoDatabase)
	if err != nil {
		return nil, errors.Wrap(err, "Creating MongoDB store failed")
	}

	return store, nil
}
