package store

import (
	"fmt"
	"sync"

	"keepair/pkg/common"
	"keepair/pkg/log"
)

type IStore interface {
	Set(key string, value []byte) error
	Delete(key string) error
	Get(key string) ([]byte, error)
	GetObjectCount() int
	StreamEntries() <-chan common.Entry
	QueueOperations(operations []common.EntryOperation) error
	ApplyOperations() error
}

type MemStore struct {
	WorkerID string

	dataMu sync.RWMutex
	Data   map[string][]byte

	opQueueMu       sync.RWMutex
	OperationsQueue []common.EntryOperation
}

func NewMemStore(workerID string) IStore {
	return &MemStore{
		WorkerID: workerID,
		Data:     make(map[string][]byte),
	}
}

func (m *MemStore) Set(key string, value []byte) error {
	m.dataMu.Lock()
	m.Data[key] = value
	m.dataMu.Unlock()
	return nil
}

func (m *MemStore) Delete(key string) error {
	m.dataMu.Lock()
	delete(m.Data, key)
	m.dataMu.Unlock()
	return nil
}

func (m *MemStore) Get(key string) ([]byte, error) {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()

	if value, ok := m.Data[key]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("no value found for key: %s", key)
}

func (m *MemStore) GetObjectCount() int {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()

	return len(m.Data)
}

func (m *MemStore) StreamEntries() <-chan common.Entry {
	ch := make(chan common.Entry)
	go func() {
		m.dataMu.RLock()
		defer m.dataMu.RUnlock()

		for k, v := range m.Data {
			ch <- common.Entry{
				Key:   k,
				Value: v,
			}
		}
		close(ch)
	}()
	return ch
}

func (m *MemStore) QueueOperations(operations []common.EntryOperation) error {
	m.opQueueMu.Lock()
	defer m.opQueueMu.Unlock()

	for _, op := range operations {
		log.Get().Printf("[%s] ADDED TO QUEUE: %s => %s", m.WorkerID, op.Action, op.Entry.Key)
		m.OperationsQueue = append(m.OperationsQueue, op)
	}

	return nil
}

func (m *MemStore) ApplyOperations() error {
	m.opQueueMu.Lock()
	m.dataMu.Lock()
	defer func() {
		m.dataMu.Unlock()
		m.opQueueMu.Unlock()
	}()

	log.Get().Printf("=============")
	log.Get().Printf("[%s] BEFORE APPLY: %d", m.WorkerID, len(m.Data))
	log.Get().Printf("=============")

	for _, op := range m.OperationsQueue {
		log.Get().Printf("[%s] entry: %s %s (%d)", m.WorkerID, op.Action, op.Entry.Key, len(op.Entry.Value))
		switch op.Action {
		case common.SetEntry:
			m.Data[op.Entry.Key] = op.Entry.Value
		case common.DeleteEntry:
			delete(m.Data, op.Entry.Key)
		default:
			panic(fmt.Errorf("invalid entry action: %s", op.Action))
		}
	}

	log.Get().Printf("=============")
	log.Get().Printf("[%s] AFTER APPLY: %d", m.WorkerID, len(m.Data))
	log.Get().Printf("=============")

	m.OperationsQueue = make([]common.EntryOperation, 0)

	return nil
}
