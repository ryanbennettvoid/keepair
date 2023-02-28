package node

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"keepair/pkg/common"
	"keepair/pkg/log"
	"keepair/pkg/partition"
	"keepair/pkg/primary/clients"
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

	log.BigPrintf("OLD NODES: %+v", m.Nodes)

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

	log.BigPrintf("NEW NODES: %+v", nodes)

	defer func() {
		m.Nodes = nodes
		m.Indexes = indexes
	}()

	if numNodes == 0 {
		return nil
	}

	log.BigPrintf("[%s] REBALANCE STARTED...", "primary")
	defer log.BigPrintf("[%s] REBALANCE DONE", "primary")

	// handle buffering of operations and set flush callback
	q := NewTransferOperationsQueueWithCallback(50, func(items []TransferOperation) error {
		for _, item := range items {
			if err := handleBulkTransferOps(item.Entry, item.SourceNode, item.TargetNode); err != nil {
				return err
			}
		}
		return nil
	})

	// for each node, regenerate partition keys and move data to correct node
	for _, currentNode := range nodes {

		var entryChan <-chan common.Entry
		var errChan <-chan error

		workerClient := clients.NewWorkerClient(currentNode.URL())
		if operation == DeleteNode {
			// if deleting, then stream entries from the node
			// that will soon be deleted
			workerClient = clients.NewWorkerClient(opNode.URL())
		}
		entryChan, errChan = workerClient.StreamEntries()

		loop := true

		for loop {
			select {
			case err := <-errChan:
				if err != nil {
					return err
				}
				loop = false
			case entry := <-entryChan:
				switch operation {
				case AddNode:
					sourceNode := nodes[indexes[currentNode.Index]]
					targetNodeIndex := partition.GenerateDeterministicPartitionKey(entry.Key, numNodes)
					targetNode := nodes[indexes[targetNodeIndex]]
					if sourceNode.ID != targetNode.ID {

						// if err := handleBulkTransferOps(entry, sourceNode, targetNode); err != nil {
						// 	return err
						// }

						if err := q.Push(NewTransferOperation(entry, sourceNode, targetNode)); err != nil {
							return err
						}
					}
				case DeleteNode:
					sourceNode := opNode
					targetNodeIndex := partition.GenerateDeterministicPartitionKey(entry.Key, numNodes)
					targetNode := nodes[indexes[targetNodeIndex]]

					// if err := handleBulkTransferOps(entry, sourceNode, targetNode); err != nil {
					// 	return err
					// }

					if err := q.Push(NewTransferOperation(entry, sourceNode, targetNode)); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := q.Flush(); err != nil {
		return err
	}

	// apply operations for all nodes
	for _, n := range nodes {
		workerClient := clients.NewWorkerClient(n.URL())
		if err := workerClient.ApplyOperations(); err != nil {
			return err
		}
	}

	return nil
}

func handleBulkTransferOps(entry common.Entry, source Node, target Node) error {
	sourceNodeClient := clients.NewWorkerClient(source.URL())
	targetNodeClient := clients.NewWorkerClient(target.URL())
	// set key on new node
	if err := targetNodeClient.QueueOperations([]common.EntryOperation{
		{
			Action: common.SetEntry,
			Entry:  entry,
		},
	}); err != nil {
		return fmt.Errorf("failed to queue operation: %w", err)
	}
	// delete key on old node
	if err := sourceNodeClient.QueueOperations([]common.EntryOperation{
		{
			Action: common.DeleteEntry,
			Entry: common.Entry{
				Key:   entry.Key,
				Value: nil,
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to queue operations: %w", err)
	}
	log.Get().Printf("transferred key (%s) from node %s => %s", entry.Key, source.ID, target.ID)
	return nil
}
