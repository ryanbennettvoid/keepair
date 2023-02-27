package node

import (
	"fmt"
	"net/http"
	"time"

	"keepair/pkg/clients"
	"keepair/pkg/common"
	"keepair/pkg/log"
)

type Node struct {
	Index                int              `json:"index"`
	ID                   string           `json:"id"`
	Address              string           `json:"address"`
	LastHealthCheckTime  time.Time        `json:"lastHealthCheckTime"`
	LastHealthCheckError error            `json:"lastHealthCheckError"`
	Stats                common.NodeStats `json:"stats"`
}

func NewNode(ID, address, port string) Node {
	return Node{
		ID:                   ID,
		Address:              fmt.Sprintf("%s:%s", address, port),
		LastHealthCheckTime:  time.Time{},
		LastHealthCheckError: nil,
	}
}

func (node *Node) PerformHealthCheck() {
	node.LastHealthCheckTime = time.Now()
	address := node.Address
	ID := node.ID

	url := fmt.Sprintf("http://%s/health", address)

	res, err := http.Get(url)
	success := err == nil && res.StatusCode == 200
	defer func() {
		if success {
			log.Get().Printf("health check succeeded for node %s (%s)", ID, url)
		} else {
			log.Get().Printf("health check failed for node %s (%s)", ID, url)
		}
	}()
	if err != nil {
		node.LastHealthCheckError = err
		return
	}

	if res.StatusCode != 200 {
		node.LastHealthCheckError = fmt.Errorf("health check failed with status code: %d", res.StatusCode)
		return
	}

	node.LastHealthCheckError = nil
}

func (node *Node) LoadStats() error {
	workerNodeURL := fmt.Sprintf("http://%s", node.Address)
	client := clients.NewWorkerClient(workerNodeURL)
	stats, err := client.GetStats()
	if err != nil {
		return err
	}
	node.Stats = stats
	return nil
}
