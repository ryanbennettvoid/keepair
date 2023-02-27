package primary

import (
	"context"

	"keepair/pkg/base_server"
	"keepair/pkg/node"
	"keepair/pkg/primary/endpoints"

	"github.com/gin-gonic/gin"
)

type Server struct {
	NodeService node.IService
}

func NewServer(nodeService node.IService) base_server.IServer {
	return &Server{
		nodeService,
	}
}

func (s *Server) Run(ctx context.Context, port string) error {

	r := gin.Default()
	r.POST("/register", endpoints.RegisterNodeHandler(s.NodeService))
	r.GET("/nodes", endpoints.GetNodesHandler(s.NodeService))
	r.POST("/keys/:key", endpoints.SetKeyHandler(s.NodeService))
	r.GET("/keys/:key", endpoints.GetKeyHandler(s.NodeService))

	svr := base_server.NewBaseServer(r)
	return svr.Run(ctx, port)
}
