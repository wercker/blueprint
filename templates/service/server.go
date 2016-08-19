package main

import (
	"fmt"
	"net"

	"gopkg.in/mgo.v2"
	"gopkg.in/urfave/cli.v1"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/blueprint/templates/service/server"
	"github.com/wercker/blueprint/templates/service/state"
	"google.golang.org/grpc"
)

// ErrInvalidPortNumber occurs if passing in a port that is not valid
var ErrInvalidPortNumber = errors.New("Invalid port number")

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
			Value:  "mongodb://localhost:27017/blueprint",
			EnvVar: "MONGODB_URI",
		},
		cli.StringFlag{
			Name:  "state-store",
			Usage: "storage driver, currently supported [mongo]",
			Value: "mongo",
		},
	},
}

var serverAction = func(c *cli.Context) error {
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

	//go func() {
	//log.Printf("Starting health server on port: %d", o.HealthPort)

	//checker := health.New()

	//// Add store if it supports a Probe
	//if storeProbe, ok := interface{}(store).(health.Probe); ok {
	//checker.RegisterProbe("store", storeProbe)
	//}

	//log.Printf("Health server stopped: %+v", checker.ListenAndServe(fmt.Sprintf(":%d", o.HealthPort)))
	//}()

	log.Info("Creating server")
	server, err := server.New(store)
	if err != nil {
		log.WithError(err).Error("Unable to create server")
		return errorExitCode
	}

	s := grpc.NewServer()
	core.RegisterBlueprintServer(s, server)

	log.WithField("port", o.Port).Info("Starting server on port")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
	if err != nil {
		log.WithField("port", o.Port).WithError(err).Error("Failed to listen")
		return errorExitCode
	}

	err = s.Serve(lis)
	if err != nil {
		log.WithError(err).Error("Failed to serve gRPC")
		return errorExitCode
	}

	return nil
}

func getStore(o *serverOptions) (state.Store, error) {
	switch o.StateStore {
	case "mongo":
		return getMongoStore(o)
	default:
		return nil, fmt.Errorf("Invalid state driver: %s", o.StateStore)
	}
}

func getMongoStore(o *serverOptions) (*state.MongoStore, error) {
	log.WithField("", o.MongoURI).Debug("Creating MongoDB store")

	log.Debug("Dialing the MongoDB cluster")
	session, err := mgo.Dial(o.MongoURI)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing the MongoDB cluster failed")
	}

	store, err := state.NewMongoStore(session, "")
	if err != nil {
		return nil, errors.Wrap(err, "Creating MongoDB store failed")
	}

	return store, nil
}

type serverOptions struct {
	*globalOptions
	Port         int
	HealthPort   int
	StateStore   string
	MongoURI     string
	CookieSecret string
}

func parseServerOptions(c *cli.Context) (*serverOptions, error) {
	gopts, err := parseGlobalOptions(c)
	if err != nil {
		return nil, err
	}

	port := c.Int("port")
	if !validPortNumber(port) {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	stateStore := c.String("state-store")
	mongoURI := c.String("mongo")

	return &serverOptions{
		globalOptions: gopts,
		Port:          port,
		StateStore:    stateStore,
		MongoURI:      mongoURI,
	}, nil
}

func validPortNumber(port int) bool {
	return port > 0 && port < 65535
}
