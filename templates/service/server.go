package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcmw "github.com/mwitkow/go-grpc-middleware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/blueprint/templates/service/server"
	"github.com/wercker/blueprint/templates/service/state"
	"github.com/wercker/pkg/conf"
	"github.com/wercker/pkg/health"
	"github.com/wercker/pkg/log"
	"github.com/wercker/pkg/trace"
	"google.golang.org/grpc"
	mgo "gopkg.in/mgo.v2"
	cli "gopkg.in/urfave/cli.v1"
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
		Value:  6666,
		EnvVar: "PORT",
	},
	cli.IntFlag{
		Name:   "health-port",
		Value:  6668,
		EnvVar: "HEALTH_PORT",
	},
	cli.IntFlag{
		Name:   "metrics-port",
		Value:  6669,
		EnvVar: "METRICS_PORT",
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

	healthService := health.New()

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
	healthService.RegisterProbe("store", store)

	store = state.NewTraceStore(store, tracer)
	store = state.NewMetricsStore(store)

	log.Debug("Creating server")
	server, err := server.New(store)
	if err != nil {
		log.WithError(err).Error("Unable to create server")
		return errorExitCode
	}

	// The following interceptors will be called in order (ie. top to bottom)
	interceptors := []grpc.UnaryServerInterceptor{
		trace.Interceptor(tracer),              // opentracing + expose trace ID
		grpc_prometheus.UnaryServerInterceptor, // prometheus
	}

	s := grpc.NewServer(grpcmw.WithUnaryServerChain(interceptors...))
	core.RegisterBlueprintServer(s, server)
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
	if err != nil {
		log.WithField("port", o.Port).WithError(err).Error("Failed to listen")
		return errorExitCode
	}

	errc := make(chan error, 4)

	// Shutdown on SIGINT, SIGTERM
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Start gRPC server
	go func() {
		log.WithField("port", o.Port).Info("Starting server")
		err := s.Serve(lis)
		errc <- errors.Wrap(err, "server returned an error")
	}()

	// Start health server
	go func() {
		log.WithField("port", o.HealthPort).Info("Starting health service")
		err := healthService.ListenAndServe(fmt.Sprintf(":%d", o.HealthPort))
		errc <- errors.Wrap(err, "health service returned an error")
	}()

	// Start metrics server
	go func() {
		log.WithField("port", o.HealthPort).Info("Starting metrics server")
		http.Handle("/metrics", prometheus.Handler())
		errc <- http.ListenAndServe(fmt.Sprintf(":%d", o.MetricsPort), nil)
	}()

	err = <-errc
	log.WithError(err).Info("Shutting down")

	// Gracefully shutdown the health server
	healthService.Shutdown(context.Background())

	// Gracefully shutdown the gRPC server
	s.GracefulStop()

	return nil
}

type serverOptions struct {
	*conf.TraceOptions

	MongoDatabase string
	MongoURI      string
	Port          int
	HealthPort    int
	MetricsPort   int
	StateStore    string
}

func parseServerOptions(c *cli.Context) (*serverOptions, error) {
	traceOptions := conf.ParseTraceOptions(c)

	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("invalid port number: %d", port)
	}

	healthPort := c.Int("health-port")
	if !validPortNumber(healthPort) {
		return nil, fmt.Errorf("invalid health-port number: %d", healthPort)
	}

	if healthPort == port {
		return nil, errors.New("health-port and port cannot be the same")
	}

	metricsPort := c.Int("metrics-port")
	if !validPortNumber(metricsPort) {
		return nil, fmt.Errorf("invalid metrics port number: %d", metricsPort)
	}

	if metricsPort == port {
		return nil, errors.New("metrics-port and port cannot be the same")
	}

	if metricsPort == healthPort {
		return nil, errors.New("metrics-port and health-port cannot be the same")
	}

	return &serverOptions{
		TraceOptions: traceOptions,

		MongoDatabase: c.String("mongo-database"),
		MongoURI:      c.String("mongo"),
		Port:          port,
		HealthPort:    healthPort,
		MetricsPort:   metricsPort,
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
