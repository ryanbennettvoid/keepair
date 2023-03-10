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
	"keepair/pkg/worker/store"

	"github.com/google/uuid"
)

type IService interface {
	GetID() string
	Run(ctx context.Context, port string) error
}

type Service struct {
	ID             string
	PrimaryNodeURL string
	Store          store.IStore
}

func NewService(primaryNodeURL string) IService {
	ID := uuid.NewString()
	return &Service{
		ID:             ID,
		PrimaryNodeURL: primaryNodeURL,
		Store:          store.NewMemStore(ID),
	}
}

func (m *Service) GetID() string {
	return m.ID
}

func (m *Service) Run(ctx context.Context, port string) error {

	// go func() {
	// 	for {
	// 		log.Get().Printf("STORAGE (%s): %d", m.ID, m.Store.GetObjectCount())
	// 		time.Sleep(time.Second * 2)
	// 	}
	// }()

	errChan := make(chan error)

	go func() {
		log.Get().Printf("running WORKER (%s) on port %s\n", m.ID, port)
		server := NewServer(m.ID, m.Store)
		errChan <- server.Run(ctx, port)
	}()

	// attempt to register self to primary node until
	// context is cancelled or success
	// TODO: don't repeat if rebalance failed
	success := false
	for !success {
		err := m.registerSelf(ctx, port)
		if err == nil {
			success = true
			break
		}
		log.Get().Printf("register self ERR: %s", err)
		if contextErr := ctx.Err(); contextErr != nil {
			return fmt.Errorf("context err (%w) while registering self: %s\n", contextErr, err.Error())
			break
		}
		log.Get().Printf("worker register self failed- trying again (%s)", m.ID)
		time.Sleep(time.Millisecond * 200)
	}
	log.Get().Printf("worker register self success: %s", m.ID)

	return <-errChan
}

func (m *Service) registerSelf(ctx context.Context, port string) error {
	registerURL := fmt.Sprintf("%s/nodes", m.PrimaryNodeURL)
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
