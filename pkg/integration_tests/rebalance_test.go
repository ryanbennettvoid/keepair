package integration_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"
	"time"

	"keepair/pkg/node"
	"keepair/pkg/primary"
	"keepair/pkg/seeder"
	"keepair/pkg/worker"

	"github.com/stretchr/testify/assert"
)

// TestRebalanceOnAddWorkerNode checks that the worker node keys are rebalanced
// when a node is added
func TestRebalanceOnAddWorkerNode(t *testing.T) {

	testMu.Lock()
	defer testMu.Unlock()

	masterNodeURL := "http://0.0.0.0:8000"

	errChan := make(chan error)
	allContext, cancel := context.WithCancel(context.Background())

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(allContext, "8000"); err != nil {
			errChan <- err
		}
	}()

	// run worker0 node in background
	go func() {
		worker0 := worker.NewService(masterNodeURL)
		if err := worker0.Run(allContext, "8001"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for primary node and worker node to init
	time.Sleep(time.Millisecond * 500)

	// set keys
	numObjects := 100
	objectSize := 50
	var items map[string][]byte
	{
		maxConcurrency := 100
		s := seeder.NewSeeder(masterNodeURL, maxConcurrency, objectSize)
		result, err := s.SeedKVs(numObjects)
		if err != nil {
			panic(err)
		}
		items = result
	}

	// check all keys
	assert.Equal(t, numObjects, len(items))
	{
		for k, v := range items {
			url := "http://0.0.0.0:8000/keys/" + k
			res, err := http.Get(url)
			panicErr(err)
			assert.Equal(t, 200, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, v, body)
			if res.StatusCode != 200 {
				panic(fmt.Errorf("get request failed: %s", body))
			}
		}
	}

	// check object count of first node, should be equal to object count
	{
		res, err := http.Get("http://0.0.0.0:8000/nodes")
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		var nodes struct {
			Nodes []node.Node `json:"nodes"`
		}
		err = json.Unmarshal(body, &nodes)
		assert.NoError(t, err)
		assert.Len(t, nodes.Nodes, 1)
		assert.Equal(t, 0, nodes.Nodes[0].Index)
		assert.Equal(t, numObjects, nodes.Nodes[0].Stats.ObjectCount)
	}

	// run worker1 node in background
	go func() {
		worker1 := worker.NewService(masterNodeURL)
		if err := worker1.Run(allContext, "8002"); err != nil {
			errChan <- err
		}
	}()

	time.Sleep(time.Second * 1)

	// check both nodes, should be roughly equal in objects
	{
		res, err := http.Get("http://0.0.0.0:8000/nodes")
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		var nodes struct {
			Nodes []node.Node `json:"nodes"`
		}
		err = json.Unmarshal(body, &nodes)
		assert.NoError(t, err)
		assert.Len(t, nodes.Nodes, 2)

		count0 := nodes.Nodes[0].Stats.ObjectCount
		count1 := nodes.Nodes[1].Stats.ObjectCount

		assert.Equalf(t, count0+count1, numObjects, "counts should add up to total number of objects")
		assert.Truef(t, math.Abs(float64(count0-count1)) < float64(numObjects)*0.10, "delta between counts should be less than 10 percent of total")
	}

	cancel() // close servers
	assert.ErrorContains(t, <-errChan, "context canceled")
}
