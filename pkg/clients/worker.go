package clients

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"keepair/pkg/common"
	"keepair/pkg/streamer"
	"keepair/pkg/values"
)

type IWorkerClient interface {
	SetKey(key string, value []byte) error
	DeleteKey(key string) error
	GetKey(key string) ([]byte, error)
	GetStats() (common.NodeStats, error)
	StreamEntries() (<-chan common.Entry, <-chan error)
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
	url := fmt.Sprintf("%s/keys/%s", w.WorkerNodeURL, key)
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

func (w WorkerClient) DeleteKey(key string) error {
	url := fmt.Sprintf("%s/keys/%s", w.WorkerNodeURL, key)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete key request failed: %s", body)
	}
	return nil
}

func (w WorkerClient) GetKey(key string) ([]byte, error) {
	url := fmt.Sprintf("%s/keys/%s", w.WorkerNodeURL, key)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("get key request failed: %s", body)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (w WorkerClient) GetStats() (common.NodeStats, error) {
	url := fmt.Sprintf("%s/stats", w.WorkerNodeURL)
	res, err := http.Get(url)
	if err != nil {
		return common.NodeStats{}, err
	}
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return common.NodeStats{}, fmt.Errorf("get stats request failed: %s", body)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return common.NodeStats{}, err
	}
	var stats struct {
		Stats common.NodeStats `json:"stats"`
	}
	if unmarshalErr := json.Unmarshal(body, &stats); unmarshalErr != nil {
		return common.NodeStats{}, unmarshalErr
	}
	return stats.Stats, nil
}

func (w WorkerClient) StreamEntries() (<-chan common.Entry, <-chan error) {
	entryChan := make(chan common.Entry)
	errChan := make(chan error)

	go func() {
		url := fmt.Sprintf("%s/stream-entries", w.WorkerNodeURL)
		res, err := http.Get(url)
		if err != nil {
			errChan <- fmt.Errorf("decode message err: %w", err)
			return
		}
		defer res.Body.Close()
		buf := make([]byte, values.StreamBufferSize)
		scanner := bufio.NewScanner(res.Body)
		scanner.Buffer(buf, values.StreamBufferSize)
		for scanner.Scan() {
			entry, err := streamer.DecodeMessage(scanner.Text())
			if err != nil {
				errChan <- fmt.Errorf("decode message err: %w", err)
				return
			}
			entryChan <- entry
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("scanner err: %w", err)
			return
		}
		errChan <- nil
	}()

	return entryChan, errChan
}
