package worker

import (
	"context"

	"keepair/pkg/base_server"

	"github.com/gin-gonic/gin"
)

type Server struct{}

func NewServer() base_server.IServer {
	return &Server{}
}

func (s *Server) Run(ctx context.Context, port string) error {
	r := gin.Default()
	svr := base_server.NewBaseServer(r)
	return svr.Run(ctx, port)
}
