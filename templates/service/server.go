package main

import (
	"fmt"
	"net"

	"gopkg.in/mgo.v2"
	"gopkg.in/urfave/cli.v1"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/wercker/{{package .Name}}/core"
	"github.com/wercker/{{package .Name}}/server"
	"github.com/wercker/{{package .Name}}/state"
	"google.golang.org/grpc"
)

var serverCommand = cli.Command{
	Name:   "server",
	Usage:  "start gRPC server",
	Action: serverAction,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "port",
			Value: {{ .Port }},
			EnvVar: "PORT",
		},
		cli.StringFlag{
			Name:  "mongo",
			Value: "mongodb://localhost:27017/{{package .Name }}",
		},
		cli.StringFlag{
			Name:  "state-store",
			Usage: "storage driver, currently supported [mongo]",
			Value: "mongo",
		},
	},
}

var serverAction = func(c *cli.Context) error {
	o, err := ParseServerOptions(c)
	if err != nil {
		log.WithError(err).Error("Unable to validate arguments")
		return ErrorExitCode
	}

	store, err := getStore(o)
	if err != nil {
		log.WithError(err).Error("Unable to create state store")
		return ErrorExitCode
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
		return ErrorExitCode
	}

	s := grpc.NewServer()
	core.Register{{method .Name }}Server(s, server)

	log.WithField("port", o.Port).Info("Starting server on port")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
	if err != nil {
		log.WithField("port", o.Port).WithError(err).Error("Failed to listen")
		return ErrorExitCode
	}

	err = s.Serve(lis)
	if err != nil {
		log.WithError(err).Error("Failed to serve gRPC")
		return ErrorExitCode
	}

	return nil
}

func getStore(o *ServerOptions) (state.Store, error) {
	switch o.StateStore {
	case "mongo":
		return getMongoStore(o)
	default:
		return nil, fmt.Errorf("Invalid state driver: %s", o.StateStore)
	}
}

func getMongoStore(o *ServerOptions) (*state.MongoStore, error) {
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

// ServerOptions are the options available for the server command.
type ServerOptions struct {
	*GlobalOptions
	Port         int    
	HealthPort   int   
	StateStore   string 
	MongoURI     string 
	CookieSecret string 
}

// ParseServerOptions will parse the options that apply to the server command.
func ParseServerOptions(c *cli.Context) (*ServerOptions, error) {
	globalOptions, err := ParseGlobalOptions(c)
	if err != nil {
		return nil, err
	}

	port := c.Int("port")
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("Invalid port number: %d", port)
	}

	stateStore := c.String("state-store")
	mongoURI := c.String("mongo")

	return &ServerOptions{
		GlobalOptions: globalOptions,
		Port:          port,
		StateStore:    stateStore,
		MongoURI:      mongoURI,
	}, nil
}
