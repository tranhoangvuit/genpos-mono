package handler

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
)

// Server implements GenposServiceHandler.
type Server struct {
	genposv1connect.UnimplementedGenposServiceHandler
	logger *slog.Logger
}

// NewServer creates a Server with all required dependencies.
func NewServer(logger *slog.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Ping(
	_ context.Context,
	_ *connect.Request[genposv1.PingRequest],
) (*connect.Response[genposv1.PingResponse], error) {
	return connect.NewResponse(&genposv1.PingResponse{
		Message: "pong",
	}), nil
}

var _ genposv1connect.GenposServiceHandler = (*Server)(nil)
