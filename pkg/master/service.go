package master

import (
	"fmt"
	"log"

	"keepair/pkg/master/endpoints"
	"keepair/pkg/node"

	"github.com/gin-gonic/gin"
)

type IService interface {
	Run(port string) error
}

type MasterService struct {
	NodeService node.IService
}

func NewService() IService {
	return &MasterService{
		NodeService: node.NewService(),
	}
}

func (m *MasterService) Run(port string) error {
	m.NodeService.RunHealthChecks()

	r := gin.Default()
	r.GET("/health", endpoints.HealthHandler())
	r.POST("/register", endpoints.RegisterNodeHandler(m.NodeService))
	log.Default().Printf("running on port %s\n", port)
	return r.Run(fmt.Sprintf(":%s", port))
}
