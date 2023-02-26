package store

import (
	"fmt"
)

type IStore interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
}

type MemStore struct {
	Data map[string][]byte
}

func NewMemStore() IStore {
	return &MemStore{
		Data: make(map[string][]byte),
	}
}

func (m *MemStore) Set(key string, value []byte) error {
	m.Data[key] = value
	return nil
}

func (m *MemStore) Get(key string) ([]byte, error) {
	if value, ok := m.Data[key]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("no value found for key: %s", key)
}
