package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var ErrorNotStarted = errors.New("healthcheck not started")
var ErrorPending = errors.New("healthcheck pending")

type CancelFunc = func()

type IService interface {
	Start() CancelFunc
	IsHealthy() bool
	Error() error
	WaitUntilHealthy(ctx context.Context) error
}

type Service struct {
	sync.RWMutex
	targetURL string
	lastError error
}

func NewService(targetURL string) IService {
	return &Service{
		targetURL: targetURL,
		lastError: ErrorNotStarted,
	}
}

func (s *Service) Start() CancelFunc {
	s.Lock()
	s.lastError = ErrorPending
	s.Unlock()

	quit := atomic.Bool{}

	go func() {
		for {
			if quit.Load() {
				break
			}
			res, err := http.Get(s.targetURL)
			if err != nil {
				s.Lock()
				s.lastError = err
				s.Unlock()
				return
			}
			if res.StatusCode >= 400 {
				s.Lock()
				s.lastError = fmt.Errorf("bad status code: %s", res.StatusCode)
				s.Unlock()
				return
			}
			s.Lock()
			s.lastError = nil
			s.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return func() {
		quit.Store(true)
	}
}

func (s *Service) IsHealthy() bool {
	s.RLock()
	defer s.RUnlock()
	return s.lastError == nil
}

func (s *Service) Error() error {
	return s.lastError
}

func (s *Service) WaitUntilHealthy(ctx context.Context) error {
	if errors.Is(s.lastError, ErrorNotStarted) {
		return s.lastError
	}
	attempt := 1
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if s.IsHealthy() {
			return nil
		}
		log.Default().Printf("healcheck attempt (%d) failed with error: %s\n", attempt, s.Error())
		attempt++
		time.Sleep(time.Second)
	}
}
