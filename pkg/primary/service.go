package primary

import (
	"context"

	"keepair/pkg/node"
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

	server := NewServer()
	return server.Run(ctx, port)
}
