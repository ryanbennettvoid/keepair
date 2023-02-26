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
	GetNodes() []Node
}

type Service struct {
	sync.RWMutex
	Nodes map[string]Node
}

func NewService() IService {
	return &Service{
		Nodes: make(map[string]Node),
	}
}

func (m *Service) RegisterNode(nd Node) {
	m.Lock()
	if _, ok := m.Nodes[nd.ID]; !ok {
		nextIndex := len(m.Nodes)
		nd.Index = nextIndex
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

func (m *Service) GetNodes() []Node {
	nodes := make([]Node, 0)
	for _, v := range m.Nodes {
		nodes = append(nodes, v)
	}
	return nodes
}
