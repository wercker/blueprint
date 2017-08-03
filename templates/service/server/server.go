package server

import (
	"github.com/wercker/blueprint/templates/service/blue_print"
	"github.com/wercker/blueprint/templates/service/state"

	"golang.org/x/net/context"
)

// New Creates a new BlueprintServer which implements blue_print.BlueprintServer.
func New(store state.Store) (*BlueprintServer, error) {
	return &BlueprintServer{
		store: store,
	}, nil
}

// BlueprintServer implements blue_print.BlueprintServer.
type BlueprintServer struct {
	store state.Store
}

// Action is a example implementation and should be replaced with an actual
// implementation.
func (s *BlueprintServer) Action(ctx context.Context, req *blue_print.ActionRequest) (*blue_print.ActionResponse, error) {
	return &blue_print.ActionResponse{}, nil
}

// Make sure that BlueprintServer implements the blue_print.BlueprintService interface.
var _ blue_print.BlueprintServer = &BlueprintServer{}
