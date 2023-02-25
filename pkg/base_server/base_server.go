package base_server

import (
	"context"
	"net/http"
	"time"

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

	// all servers have health check
	r.GET("/health", func(c *gin.Context) {
		c.Data(200, "", []byte("ok"))
	})

	log.Default().Printf("running on port %s", port)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	errChan := make(chan error)

	go func() {
		// check if server should quit
		for {
			if err := ctx.Err(); err != nil {
				_ = httpServer.Close()
				log.Default().Println("closed server")
				errChan <- err
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	return <-errChan
}
