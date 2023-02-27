package integration_tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"keepair/pkg/primary"
	"keepair/pkg/worker"

	"github.com/stretchr/testify/assert"
)

// TestSetGetKV checks that a value can be set and get
// with a running cluster
func TestSetGetKV(t *testing.T) {

	testMu.Lock()
	defer testMu.Unlock()

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
		res, err := http.Post("http://0.0.0.0:8000/keys/myKey", "", bytes.NewReader(postBody))
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte("ok"), body)
		if res.StatusCode != 200 {
			panic(fmt.Errorf("set request failed: %s", body))
		}
	}

	// get the key
	{
		expectedBody := []byte("this is a value for a key")
		res, err := http.Get("http://0.0.0.0:8000/keys/myKey")
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, body)
		if res.StatusCode != 200 {
			panic(fmt.Errorf("get request failed: %s", body))
		}
	}

	// delete the key
	{
		req, err := http.NewRequest(http.MethodDelete, "http://0.0.0.0:8000/keys/myKey", nil)
		panicErr(err)
		res, err := http.DefaultClient.Do(req)
		panicErr(err)
		assert.Equal(t, 200, res.StatusCode)
		assert.NoError(t, err)
		if res.StatusCode != 200 {
			panic(fmt.Errorf("delete request failed: %s", "myKey"))
		}
	}

	cancel() // close servers
	assert.ErrorContains(t, <-errChan, "context canceled")
}
