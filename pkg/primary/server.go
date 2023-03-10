package primary

import (
	"context"

	"keepair/pkg/base_server"
	"keepair/pkg/primary/endpoints"
	"keepair/pkg/primary/node"

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
	r.GET("/nodes", endpoints.GetNodesHandler(s.NodeService))
	r.POST("/nodes", endpoints.RegisterNodeHandler(s.NodeService))
	r.DELETE("/nodes/:nodeID", endpoints.UnregisterNodeHandler(s.NodeService))
	r.POST("/keys/:key", endpoints.SetKeyHandler(s.NodeService))
	r.GET("/keys/:key", endpoints.GetKeyHandler(s.NodeService))
	r.DELETE("/keys/:key", endpoints.DeleteKeyHandler(s.NodeService))

	svr := base_server.NewBaseServer(r)
	return svr.Run(ctx, port)
}
