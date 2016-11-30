package main

import (
	"fmt"
	"net"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/urfave/cli.v1"

	log "github.com/Sirupsen/logrus"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/pkg/errors"
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/blueprint/templates/service/server"
	"github.com/wercker/blueprint/templates/service/state"
	"google.golang.org/grpc"
)

var serverCommand = cli.Command{
	Name:   "server",
	Usage:  "start gRPC server",
	Action: serverAction,
	Flags: []cli.Flag{
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

var serverAction = func(c *cli.Context) error {
	log.Info("Starting blueprint server")

	log.Debug("Parsing server options")
	o, err := parseServerOptions(c)
	if err != nil {
		log.WithError(err).Error("Unable to validate arguments")
		return errorExitCode
	}

	store, err := getStore(o)
	if err != nil {
		log.WithError(err).Error("Unable to create state store")
		return errorExitCode
	}
	defer store.Close()

	log.Debug("Creating server")
	server, err := server.New(store)
	if err != nil {
		log.WithError(err).Error("Unable to create server")
		return errorExitCode
	}

	if o.Trace {
		log.Info("Tracing is enabled")
		tracer, err := getTracer(o.TraceEndpoint, "blueprint", fmt.Sprintf(":%d", o.Port), false, false)
		if err != nil {
			log.Println(err.Error())
			return cli.NewExitError("Unable to create tracer", 5)
		}

		store = state.NewTracingStore(store, tracer)
		// intercept = grpcmw.ChainUnaryServer(otgrpc.OpenTracingServerInterceptor(tracer), server)
	}

	s := grpc.NewServer()
	core.RegisterBlueprintServer(s, server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
	if err != nil {
		log.WithField("port", o.Port).WithError(err).Error("Failed to listen")
		return errorExitCode
	}

	log.WithField("port", o.Port).Info("Starting server")
	err = s.Serve(lis)
	if err != nil {
		log.WithError(err).Error("Failed to serve gRPC")
		return errorExitCode
	}

	return nil
}

type serverOptions struct {
	MongoDatabase string
	MongoURI      string
	Port          int
	StateStore    string
}

func parseServerOptions(c *cli.Context) (*serverOptions, error) {
	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	return &serverOptions{
		MongoDatabase: c.String("mongo-database"),
		MongoURI:      c.String("mongo"),
		Port:          port,
		StateStore:    c.String("state-store"),
	}, nil
}

func getTracer(endpoint, serviceName, hostPort string, debug, sameSpan bool) (opentracing.Tracer, error) {
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = fmt.Sprintf("http://%s", endpoint)
	}

	if strings.Count(endpoint, ":") == 1 {
		endpoint = fmt.Sprintf("%s:9411", endpoint)
	}

	collector, err := zipkin.NewHTTPCollector(fmt.Sprintf("%s/api/v1/spans", endpoint))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Zipkin HTTP collector")
	}

	// create recorder.
	recorder := zipkin.NewRecorder(collector, debug, hostPort, serviceName)

	// create tracer.
	tracer, err := zipkin.NewTracer(recorder, zipkin.ClientServerSameSpan(sameSpan))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Zipkin tracer")
	}
	return tracer, nil
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
