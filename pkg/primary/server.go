package primary

import (
	"context"

	"keepair/pkg/base_server"
	"keepair/pkg/node"
	"keepair/pkg/primary/endpoints"

	"github.com/gin-gonic/gin"
)

type Server struct{}

func NewServer() base_server.IServer {
	return &Server{}
}

func (s *Server) Run(ctx context.Context, port string) error {
	nodeService := node.NewService()
	nodeService.RunHealthChecksInBackground()

	r := gin.Default()
	r.POST("/register", endpoints.RegisterNodeHandler(nodeService))

	svr := base_server.NewBaseServer(r)
	return svr.Run(ctx, port)
}
