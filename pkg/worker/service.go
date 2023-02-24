package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"keepair/pkg/common"
	"keepair/pkg/worker/endpoints"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IService interface {
	Run(port string) error
}

type Service struct {
	ID            string
	MasterNodeURL string
}

func NewService(masterNodeURL string) IService {
	return &Service{
		ID:            uuid.NewString(),
		MasterNodeURL: masterNodeURL,
	}
}

func (m *Service) registerSelf(port string) error {
	registerURL := fmt.Sprintf("%s/register", common.MustGetEnv("MASTER_NODE_URL"))
	body := map[string]any{
		"id":   m.ID,
		"port": port,
	}
	bodyStr, err := json.Marshal(body)
	if err != nil {
		return err
	}
	res, err := http.Post(registerURL, "application/json", bytes.NewReader(bodyStr))
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		defer res.Body.Close()
		resBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to register: %s", string(resBody))
	}
	return nil
}

func (m *Service) Run(port string) error {
	if err := m.registerSelf(port); err != nil {
		return err
	}
	r := gin.Default()
	r.GET("/health", endpoints.HealthHandler())
	log.Default().Printf("running WORKER (%s) on port %s\n", m.ID, port)
	return r.Run(fmt.Sprintf(":%s", port))
}
