package server

import (
	"github.com/wercker/{{package .Name}}/core"
	"github.com/wercker/{{package .Name}}/state"
)

func New(store state.Store) (*{{class .Name}}Server, error) {
	return &{{class .Name}}Server{
		store:        store,
	}, nil
}

type {{class .Name}}Server struct {
	store        state.Store
}

// Make sure that {{.Name}}Server implements the core.{{.Name}}Service interface.
var _ core.{{class .Name}}Server = &{{class .Name}}Server{}
