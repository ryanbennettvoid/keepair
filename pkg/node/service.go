package node

import (
	"fmt"
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
	GetNodeByIndex(idx int) (Node, error)
	GetNumNodes() int
}

type Service struct {
	sync.RWMutex
	Indexes map[int]string
	Nodes   map[string]Node
}

func NewService() IService {
	return &Service{
		Indexes: make(map[int]string),
		Nodes:   make(map[string]Node),
	}
}

func (m *Service) RegisterNode(nd Node) {
	m.Lock()
	if _, ok := m.Nodes[nd.ID]; !ok {
		nextIndex := len(m.Nodes)
		nd.Index = nextIndex
		m.Nodes[nd.ID] = nd
		m.Indexes[nextIndex] = nd.ID
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
			for i, n := range m.Nodes {
				n.PerformHealthCheck()
				m.Nodes[i] = n
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

func (m *Service) GetNodeByIndex(idx int) (Node, error) {
	m.RLock()
	defer m.RUnlock()
	nodeID, ok := m.Indexes[idx]
	if !ok {
		return Node{}, fmt.Errorf("failed to find node index: %d", idx)
	}
	n, ok := m.Nodes[nodeID]
	if !ok {
		return Node{}, fmt.Errorf("failed to find node: %s", nodeID)
	}
	return n, nil
}

func (m *Service) GetNumNodes() int {
	return len(m.Indexes)
}
