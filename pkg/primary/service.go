package primary

import (
	"context"

	"keepair/pkg/primary/node"
)

type IService interface {
	Run(ctx context.Context, port string) error
}

type Service struct{}

func NewService() IService {
	return &Service{}
}

func (m *Service) Run(ctx context.Context, port string) error {
	nodeService := node.NewService()
	cancelHealthCheck := nodeService.RunHealthChecksInBackground()
	defer cancelHealthCheck()

	server := NewServer(nodeService)
	return server.Run(ctx, port)
}
