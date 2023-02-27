package node

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"keepair/pkg/clients"
	"keepair/pkg/common"
	"keepair/pkg/log"
	"keepair/pkg/partition"
)

type CancelFunc func()

type IService interface {
	RegisterNode(nd Node) error
	UnregisterNode(ID string) error
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

func (m *Service) RegisterNode(nd Node) error {
	m.Lock()
	defer m.Unlock()

	// consider registration to be a health check
	nd.LastHealthCheckTime = time.Now()

	if err := m.rebalanceNodes(AddNode, nd); err != nil {
		return fmt.Errorf("failed to rebalance nodes: %w", err)
	}

	return nil
}

func (m *Service) UnregisterNode(ID string) error {
	m.Lock()
	defer m.Unlock()

	nd, ok := m.Nodes[ID]
	if !ok {
		return fmt.Errorf("failed to find node: %s", ID)
	}

	if err := m.rebalanceNodes(DeleteNode, nd); err != nil {
		return fmt.Errorf("failed to rebalance nodes: %w", err)
	}

	return nil
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

type RebalanceOperation string

var AddNode = RebalanceOperation("add")
var DeleteNode = RebalanceOperation("delete")

// rebalanceNodes redistributes data to be stored evenly across all nodes.
// Caller must handle locks.
func (m *Service) rebalanceNodes(operation RebalanceOperation, opNode Node) error {

	// make copy of nodes map
	nodes := Map(m.Nodes)
	if operation == AddNode {
		nodes = nodes.Add(opNode)
	}
	if operation == DeleteNode {
		nodes = nodes.Delete(opNode)
		opNode.Index = -1 // for logging
	}
	indexes := nodes.CreateIndexes()
	numNodes := len(nodes)

	defer func() {
		m.Nodes = nodes
		m.Indexes = indexes
	}()

	if numNodes == 0 {
		return nil
	}

	log.Get().Printf("REBALANCE STARTED...")
	defer log.Get().Printf("REBALANCE DONE")

	// for each node, regenerate partition keys and move data to correct node
	for _, n := range nodes {
		log.Get().Printf("REBALANCE NODE %s", n.ID)
		workerNodeURL := fmt.Sprintf("http://%s", n.Address)
		workerClient := clients.NewWorkerClient(workerNodeURL)
		entryChan, errChan := workerClient.StreamEntries()

		for {
			select {
			case err := <-errChan:
				return err
			case entry := <-entryChan:
				switch operation {
				case AddNode:
					sourceNode := nodes[indexes[n.Index]]
					targetNodeIndex := partition.GenerateDeterministicPartitionKey(entry.Key, numNodes)
					targetNode := nodes[indexes[targetNodeIndex]]
					if sourceNode.ID != targetNode.ID {
						if err := transferEntry(entry, sourceNode, targetNode); err != nil {
							return err // this would be bad!
						}
					}
				case DeleteNode:
					sourceNode := opNode
					targetNodeIndex := partition.GenerateDeterministicPartitionKey(entry.Key, numNodes)
					targetNode := nodes[indexes[targetNodeIndex]]
					if err := transferEntry(entry, sourceNode, targetNode); err != nil {
						return err // this would be bad!
					}
				}
			}
		}
	}

	return nil
}

// transferEntry moves data from source node to target node
func transferEntry(entry common.Entry, source Node, target Node) error {
	sourceNodeClient := clients.NewWorkerClient(fmt.Sprintf("http://%s", source.Address))
	targetNodeClient := clients.NewWorkerClient(fmt.Sprintf("http://%s", target.Address))
	// set key on new node
	if err := targetNodeClient.SetKey(entry.Key, entry.Value); err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}
	// delete key on old node
	if err := sourceNodeClient.DeleteKey(entry.Key); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	log.Get().Printf("transferred key (%s) from node %d|%s to %d|%s", entry.Key, source.Index, source.Address, target.Index, target.Address)
	return nil
}
