package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"keepair/pkg/node"
	"keepair/pkg/primary"
	"keepair/pkg/worker"

	"github.com/stretchr/testify/assert"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

// TestPrimaryNode checks to see that the
// primary node can run and be healthy
func TestPrimaryNode(t *testing.T) {

	errChan := make(chan error)
	allContext, cancel := context.WithCancel(context.Background())

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(allContext, "8000"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for primary node server to start
	time.Sleep(time.Millisecond * 500)

	// check if primary node is healthy
	{
		res, err := http.Get("http://0.0.0.0:8000/health")
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte("ok"), body)
	}

	cancel() // close servers
	assert.ErrorContains(t, <-errChan, "context canceled")
}

// TestWorkerNodeWithoutPrimaryNode should throw an
// error since the worker node cannot register to the
// primary node
func TestWorkerNodeWithoutPrimaryNode(t *testing.T) {

	errChan := make(chan error)

	// run worker node in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		service := worker.NewService("http://0.0.0.0:8000")
		if err := service.Run(ctx, "8001"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for worker node to connect
	time.Sleep(time.Second * 1)

	err := <-errChan
	assert.Error(t, err)
	assert.ErrorContains(t, err, "context deadline exceeded")
	assert.ErrorContains(t, err, "while registering self")
}

// TestPrimaryNodeWithWorkerNode should successfully run the primary node
// and the worker node, with the worker node registered to the primary node
func TestPrimaryNodeWithWorkerNode(t *testing.T) {

	errChan := make(chan error, 2)
	allContext, cancel := context.WithCancel(context.Background())

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(allContext, "8000"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for primary node server to start
	time.Sleep(time.Millisecond * 500)

	// run worker node in background
	go func() {
		service := worker.NewService("http://0.0.0.0:8000")
		if err := service.Run(allContext, "8001"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for worker node to connect
	time.Sleep(time.Millisecond * 500)

	cancel() // close servers
	assert.ErrorContains(t, <-errChan, "context canceled")
}

// TestPrimaryNodeWithWorkerNodeRaceCondition should run the worker node first,
// then the primary node afterwards, with the worker node being registered
func TestPrimaryNodeWithWorkerNodeRaceCondition(t *testing.T) {

	errChan := make(chan error, 2)
	allContext, cancel := context.WithCancel(context.Background())

	// run worker node in background
	go func() {
		service := worker.NewService("http://0.0.0.0:8000")
		if err := service.Run(allContext, "8001"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for worker node to start registering
	time.Sleep(time.Millisecond * 500)

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(allContext, "8000"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for worker node to connect
	time.Sleep(time.Second * 1)

	// check if primary node is healthy
	{
		res, err := http.Get("http://0.0.0.0:8000/health")
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte("ok"), body)
		if res.StatusCode != 200 {
			panic(fmt.Errorf("primary node not healthy: %s", body))
		}
	}

	// check that the worker node is registered to the primary node
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
	}

	cancel() // close servers
	assert.ErrorContains(t, <-errChan, "context canceled")
}

// TestSetGetKV checks that a value can be set and get
// with a running cluster
func TestSetGetKV(t *testing.T) {

	errChan := make(chan error)
	allContext, cancel := context.WithCancel(context.Background())

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(allContext, "8000"); err != nil {
			errChan <- err
		}
	}()

	// run worker node in background
	go func() {
		service := worker.NewService("http://0.0.0.0:8000")
		if err := service.Run(allContext, "8001"); err != nil {
			errChan <- err
		}
	}()

	// wait a bit for primary node and worker node to init
	time.Sleep(time.Millisecond * 500)

	// set a key
	{
		postBody := []byte("this is a value for a key")
		res, err := http.Post("http://0.0.0.0:8000/set/myKey", "", bytes.NewReader(postBody))
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte("ok"), body)
		if res.StatusCode != 200 {
			panic(body)
		}
	}

	// get the key
	{
		expectedBody := []byte("this is a value for a key")
		res, err := http.Get("http://0.0.0.0:8000/get/myKey")
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, body)
		if res.StatusCode != 200 {
			panic(fmt.Errorf("get request failed: %s", body))
		}
	}

	cancel() // close servers
	assert.ErrorContains(t, <-errChan, "context canceled")
}
