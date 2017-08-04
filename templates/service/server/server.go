package server

import (
	"github.com/wercker/blueprint/templates/service/blue_printpb"
	"github.com/wercker/blueprint/templates/service/state"

	"golang.org/x/net/context"
)

// New Creates a new BlueprintServer which implements blue_printpb.BlueprintServer.
func New(store state.Store) (*BlueprintServer, error) {
	return &BlueprintServer{
		store: store,
	}, nil
}

// BlueprintServer implements blue_printpb.BlueprintServer.
type BlueprintServer struct {
	store state.Store
}

// Action is a example implementation and should be replaced with an actual
// implementation.
func (s *BlueprintServer) Action(ctx context.Context, req *blue_printpb.ActionRequest) (*blue_printpb.ActionResponse, error) {
	return &blue_printpb.ActionResponse{}, nil
}

// Make sure that BlueprintServer implements the blue_printpb.BlueprintService interface.
var _ blue_printpb.BlueprintServer = &BlueprintServer{}
