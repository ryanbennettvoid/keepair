package worker

import (
	"context"

	"keepair/pkg/base_server"
	"keepair/pkg/worker/endpoints"
	"keepair/pkg/worker/store"

	"github.com/gin-gonic/gin"
)

type Server struct {
	WorkerID string
	Store    store.IStore
}

func NewServer(workerID string, store store.IStore) base_server.IServer {
	return &Server{
		WorkerID: workerID,
		Store:    store,
	}
}

func (s *Server) Run(ctx context.Context, port string) error {
	r := gin.Default()

	r.POST("/keys/:key", endpoints.SetKeyHandler(s.Store))
	r.DELETE("/keys/:key", endpoints.DeleteKeyHandler(s.WorkerID, s.Store))
	r.GET("/keys/:key", endpoints.GetKeyHandler(s.Store))
	r.GET("/stats", endpoints.GetStatsHandler(s.Store))
	r.GET("/stream-entries", endpoints.StreamEntriesHandler(s.Store))
	r.POST("/queue-operations", endpoints.QueueOperationsHandler(s.Store))
	r.POST("/apply-operations", endpoints.ApplyOperationsHandler(s.Store))

	svr := base_server.NewBaseServer(r)
	return svr.Run(ctx, port)
}
