package application

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"keepair/pkg/primary"
	"keepair/pkg/worker"

	"github.com/stretchr/testify/assert"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func isTimeoutError(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}
	return false
}

func isServerClosedError(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}

// TestPrimaryNode checks to see that the
// primary node can run and be healthy
func TestPrimaryNode(t *testing.T) {

	errChan := make(chan error)
	primaryNodeContext, cancel := context.WithCancel(context.Background())

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(primaryNodeContext, "8000"); err != nil {
			if isTimeoutError(err) || isServerClosedError(err) {
				errChan <- nil
			} else {
				errChan <- err
			}
		}
	}()

	// wait a bit for primary node server to start
	time.Sleep(time.Millisecond * 500)

	// check if primary node is healthy
	res, err := http.Get("http://0.0.0.0:8000/health")
	panicErr(err)
	assert.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("ok"), body)

	cancel() // close server
	assert.NoError(t, <-errChan)
}

// TestWorkerNodeWithoutPrimaryNode should throw an
// error since the worker node cannot register to the
// primary node
func TestWorkerNodeWithoutPrimaryNode(t *testing.T) {

	errChan := make(chan error)
	workerNodeContext, cancel := context.WithCancel(context.Background())

	// run worker node in background
	go func() {
		service := worker.NewService("http://0.0.0.0:8000")
		if err := service.Run(workerNodeContext, "8001"); err != nil {
			if isTimeoutError(err) || isServerClosedError(err) {
				errChan <- nil
			} else {
				errChan <- err
			}
		}
	}()

	// wait a bit for worker node to connect
	time.Sleep(time.Millisecond * 500)

	cancel() // close server
	err := <-errChan
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register self")
}

// TestPrimaryNodeWithWorkerNode should successfully run the primary node
// and the worker node, with the worker node registered to the primary node
func TestPrimaryNodeWithWorkerNode(t *testing.T) {

	errChan := make(chan error)
	primaryNodeContext, cancel := context.WithTimeout(context.Background(), time.Second*3)

	// run primary node in background
	go func() {
		service := primary.NewService()
		if err := service.Run(primaryNodeContext, "8000"); err != nil {
			if isTimeoutError(err) || isServerClosedError(err) {
				errChan <- nil
			} else {
				errChan <- err
			}
		}
	}()

	// wait a bit for primary node server to start
	time.Sleep(time.Millisecond * 500)

	// run worker node in background
	go func() {
		service := worker.NewService("http://0.0.0.0:8000")
		if err := service.Run(primaryNodeContext, "8001"); err != nil {
			if isTimeoutError(err) || isServerClosedError(err) {
				errChan <- nil
			} else {
				errChan <- err
			}
		}
	}()

	// wait a bit for worker node to connect
	time.Sleep(time.Millisecond * 500)

	cancel() // close server
	assert.NoError(t, <-errChan)
}
