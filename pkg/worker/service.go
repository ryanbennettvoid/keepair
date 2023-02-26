package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"keepair/pkg/log"

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
	registerErr := make(chan error)
	serverErr := make(chan error)

	go func() {
		// attempt to register self to primary node until
		// context is cancelled
		success := false
		for !success {
			err := m.registerSelf(ctx, port)
			if err == nil {
				success = true
				break
			}
			log.Get().Printf("register self ERR: %s", err)
			if contextErr := ctx.Err(); contextErr != nil {
				registerErr <- fmt.Errorf("context err (%w) while registering self: %s\n", contextErr, err.Error())
				return
			}
			log.Get().Printf("ATTEMPTING REGISTER: %s", m.PrimaryNodeURL)
			time.Sleep(time.Millisecond * 500)
		}
		log.Get().Println("REGISTER SUCCESS")
	}()

	go func() {
		log.Get().Printf("running WORKER (%s) on port %s\n", m.ID, port)
		server := NewServer()
		serverErr <- server.Run(ctx, port)
	}()

	// wait for either error to occur
	for {
		select {
		case err := <-registerErr:
			return err
		case err := <-serverErr:
			return err
		}
	}
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
	req.Header.Set("Cache-Control", "no-cache")
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
