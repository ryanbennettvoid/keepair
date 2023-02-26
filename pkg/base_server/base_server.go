package base_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"keepair/pkg/log"

	"github.com/gin-gonic/gin"
)

type IServer interface {
	Run(ctx context.Context, port string) error
}

// BaseServer runs a server and accepts Context to stop the server
type BaseServer struct {
	rootRouter *gin.Engine
}

func NewBaseServer(rootRouter *gin.Engine) IServer {
	return &BaseServer{
		rootRouter: rootRouter,
	}
}

func (s *BaseServer) Run(ctx context.Context, port string) error {

	r := s.rootRouter

	// all servers have health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.Data(200, "", []byte("ok"))
	})

	log.Get().Printf("running on port %s", port)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	serverErrChan := make(chan error)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
			return
		}
		serverErrChan <- nil
	}()

	for {
		select {
		case <-ctx.Done():
			if closeErr := httpServer.Close(); closeErr != nil {
				panic(closeErr)
			}
			return fmt.Errorf("server closed: %w", ctx.Err())
		case err := <-serverErrChan:
			return err
		}
	}
}
