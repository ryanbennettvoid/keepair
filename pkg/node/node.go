package node

import (
	"fmt"
	"net/http"
	"path"
	"time"
)

type Node struct {
	ID                   string
	Address              string
	LastHealthCheckTime  time.Time
	LastHealthCheckError error
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

	url := "http://" + path.Join(node.Address, "/health")
	res, err := http.Get(url)
	success := err == nil && res.StatusCode == 200
	defer func() {
		if success {
			fmt.Printf("health check succeeded for node %s (%s)\n", node.ID, url)
		} else {
			fmt.Printf("health check failed for node %s (%s)\n", node.ID, url)
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
