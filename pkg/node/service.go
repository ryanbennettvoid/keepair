package node

import (
	"sync"
	"sync/atomic"
	"time"
)

type CancelFunc func()

type IService interface {
	RegisterNode(nd Node)
	UnregisterNode(ID string)
	RunHealthChecksInBackground() CancelFunc
}

type Service struct {
	sync.RWMutex
	Nodes map[string]Node
}

func NewService() *Service {
	return &Service{
		Nodes: make(map[string]Node),
	}
}

func (m *Service) RegisterNode(nd Node) {
	m.Lock()
	if _, ok := m.Nodes[nd.ID]; !ok {
		m.Nodes[nd.ID] = nd
	}
	m.Unlock()
}

func (m *Service) UnregisterNode(ID string) {
	m.Lock()
	delete(m.Nodes, ID)
	m.Unlock()
}

func (m *Service) RunHealthChecksInBackground() CancelFunc {

	quit := atomic.Bool{}

	go func() {
		for !quit.Load() {
			// TODO: check if lock affects read performance
			m.Lock()
			for _, node := range m.Nodes {
				node.PerformHealthCheck()
			}
			m.Unlock()
			time.Sleep(time.Second * 5)
		}
	}()

	return func() {
		quit.Store(true)
	}
}
