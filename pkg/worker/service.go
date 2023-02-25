package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type IService interface {
	Run(ctx context.Context, port string) error
}

type Service struct {
	ID             string
	PrimaryNodeURL string
}

func NewService(primaryNodeURL string) IService {
	return &Service{
		ID:             uuid.NewString(),
		PrimaryNodeURL: primaryNodeURL,
	}
}

func (m *Service) Run(ctx context.Context, port string) error {
	// attempt to register self to primary node until
	// context is cancelled
	for err := m.registerSelf(ctx, port); err != nil; {
		if contextErr := ctx.Err(); contextErr != nil {
			return fmt.Errorf("failed to register self: %w\n", err)
		}
		time.Sleep(time.Second)
	}

	log.Default().Printf("running WORKER (%s) on port %s\n", m.ID, port)

	server := NewServer()
	return server.Run(ctx, port)
}

func (m *Service) registerSelf(ctx context.Context, port string) error {
	registerURL := fmt.Sprintf("%s/register", m.PrimaryNodeURL)
	body := map[string]any{
		"id":   m.ID,
		"port": port,
	}
	bodyStr, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewReader(bodyStr))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
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
