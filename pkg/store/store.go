package store

import (
	"fmt"
	"sync"

	"keepair/pkg/common"
)

type IStore interface {
	Set(key string, value []byte) error
	Delete(key string) error
	Get(key string) ([]byte, error)
	GetObjectCount() int
	StreamEntries() <-chan common.Entry
}

type MemStore struct {
	sync.RWMutex
	Data map[string][]byte
}

func NewMemStore() IStore {
	return &MemStore{
		Data: make(map[string][]byte),
	}
}

func (m *MemStore) Set(key string, value []byte) error {
	m.Lock()
	m.Data[key] = value
	m.Unlock()
	return nil
}

func (m *MemStore) Delete(key string) error {
	m.Lock()
	delete(m.Data, key)
	m.Unlock()
	return nil
}

func (m *MemStore) Get(key string) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	if value, ok := m.Data[key]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("no value found for key: %s", key)
}

func (m *MemStore) GetObjectCount() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.Data)
}

func (m *MemStore) StreamEntries() <-chan common.Entry {
	ch := make(chan common.Entry)
	go func() {
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
