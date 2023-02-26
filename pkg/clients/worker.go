package clients

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type IWorkerClient interface {
	SetKey(key string, value []byte) error
	GetKey(key string) ([]byte, error)
}

type WorkerClient struct {
	WorkerNodeURL string
}

func NewWorkerClient(workerNodeURL string) IWorkerClient {
	return WorkerClient{
		WorkerNodeURL: workerNodeURL,
	}
}

func (w WorkerClient) SetKey(key string, value []byte) error {
	url := fmt.Sprintf("%s/set/%s", w.WorkerNodeURL, key)
	res, err := http.Post(url, "", bytes.NewReader(value))
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("set key request failed: %s", body)
	}
	return nil
}

func (w WorkerClient) GetKey(key string) ([]byte, error) {
	url := fmt.Sprintf("%s/get/%s", w.WorkerNodeURL, key)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("set key request failed: %s", body)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
