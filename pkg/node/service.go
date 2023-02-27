package node

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"keepair/pkg/clients"
	"keepair/pkg/log"
	"keepair/pkg/partition"
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
	defer m.Unlock()

	if _, ok := m.Nodes[nd.ID]; !ok {
		nextIndex := len(m.Nodes)
		nd.Index = nextIndex
		m.Nodes[nd.ID] = nd
		m.Indexes[nextIndex] = nd.ID
	}

	if err := m.rebalanceNodes(); err != nil {
		panic(err) // TODO: handle this better?
	}
}

func (m *Service) UnregisterNode(ID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Nodes, ID)
}

func (m *Service) RunHealthChecksInBackground() CancelFunc {

	quit := atomic.Bool{}

	go func() {
		for !quit.Load() {
			// TODO: check if lock affects read performance
			nodes := m.GetNodes()
			for _, n := range nodes {
				n.PerformHealthCheck()
				m.Lock()
				m.Nodes[n.ID] = n
				m.Unlock()
			}
			time.Sleep(time.Second * 5)
		}
	}()

	return func() {
		quit.Store(true)
	}
}

func (m *Service) GetNodes() []Node {
	m.RLock()
	defer m.RUnlock()
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
	m.RLock()
	defer m.RUnlock()
	return len(m.Indexes)
}

// rebalanceNodes redistributes data to be stored evenly across all nodes.
// Caller must handle locks.
func (m *Service) rebalanceNodes() error {

	if len(m.Nodes) <= 1 {
		return nil
	}

	log.Get().Printf("REBALANCE STARTED...")
	defer log.Get().Printf("REBALANCE DONE...")

	// for each node, regenerate partition keys and move data to correct node
	for _, n := range m.Nodes {
		log.Get().Printf("REBALANCE NODE %s", n.ID)

		workerNodeURL := fmt.Sprintf("http://%s", n.Address)
		workerClient := clients.NewWorkerClient(workerNodeURL)
		entryChan, errChan := workerClient.StreamEntries()

		for {
			select {
			case err := <-errChan:
				return err
			case entry := <-entryChan:
				log.Get().Printf("REBALANCE ENTRY: %s: %d", entry.Key, len(entry.Value))
				actualNodeIndex := n.Index
				expectedNodeIndex := partition.GenerateDeterministicPartitionKey(entry.Key, len(m.Nodes))
				if actualNodeIndex != expectedNodeIndex {
					// move data from actual node to expected node
					newNodeID := m.Indexes[expectedNodeIndex]
					newNode := m.Nodes[newNodeID]
					newNodeURL := fmt.Sprintf("http://%s", newNode.Address)
					newNodeClient := clients.NewWorkerClient(newNodeURL)
					// set key on new node
					if err := newNodeClient.SetKey(entry.Key, entry.Value); err != nil {
						return fmt.Errorf("failed to set key: %w", err)
					}
					// delete key on old node
					if err := workerClient.DeleteKey(entry.Key); err != nil {
						return fmt.Errorf("failed to delete key: %w", err)
					}
				}
			}
		}
	}

	return nil
}
