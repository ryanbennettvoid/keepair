package node

import (
	"sync"
	"time"
)

type IService interface {
	RegisterNode(nd Node)
	UnregisterNode(ID string)
	RunHealthChecks()
}

type Service struct {
	nodesMu sync.RWMutex
	Nodes   map[string]Node
}

func NewService() *Service {
	return &Service{
		nodesMu: sync.RWMutex{},
		Nodes:   make(map[string]Node),
	}
}

func (m *Service) RegisterNode(nd Node) {
	m.nodesMu.Lock()
	if _, ok := m.Nodes[nd.ID]; !ok {
		m.Nodes[nd.ID] = nd
	}
	m.nodesMu.Unlock()
}

func (m *Service) UnregisterNode(ID string) {
	m.nodesMu.Lock()
	delete(m.Nodes, ID)
	m.nodesMu.Unlock()
}

func (m *Service) RunHealthChecks() {
	go func() {
		for {
			m.nodesMu.RLock()
			for _, nd := range m.Nodes {
				nd.PerformHealthCheck()
			}
			m.nodesMu.RUnlock()
			time.Sleep(time.Second * 5)
		}
	}()
}
