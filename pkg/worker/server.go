package worker

import (
	"context"

	"keepair/pkg/base_server"
	"keepair/pkg/store"
	"keepair/pkg/worker/endpoints"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Store store.IStore
}

func NewServer(store store.IStore) base_server.IServer {
	return &Server{
		Store: store,
	}
}

func (s *Server) Run(ctx context.Context, port string) error {
	r := gin.Default()

	r.POST("/keys/:key", endpoints.SetKeyHandler(s.Store))
	r.DELETE("/keys/:key", endpoints.DeleteKeyHandler(s.Store))
	r.GET("/keys/:key", endpoints.GetKeyHandler(s.Store))
	r.GET("/stats", endpoints.GetStatsHandler(s.Store))
	r.GET("/stream-entries", endpoints.StreamEntriesHandler(s.Store))

	svr := base_server.NewBaseServer(r)
	return svr.Run(ctx, port)
}
