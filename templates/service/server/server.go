package server

import (
	"github.com/wercker/blueprint/templates/service/core"
	"github.com/wercker/blueprint/templates/service/state"

	"golang.org/x/net/context"
)

// New Creates a new BlueprintServer which implements core.BlueprintServer.
func New(store state.Store) (*BlueprintServer, error) {
	return &BlueprintServer{
		store: store,
	}, nil
}

// BlueprintServer implements core.BlueprintServer.
type BlueprintServer struct {
	store state.Store
}

// Action is a example implementation and should be replaced with an actual
// implementation.
func (s *BlueprintServer) Action(ctx context.Context, req *core.ActionRequest) (*core.ActionResponse, error) {
	return &core.ActionResponse{}, nil
}

// Make sure that BlueprintServer implements the core.BlueprintService interface.
var _ core.BlueprintServer = &BlueprintServer{}
