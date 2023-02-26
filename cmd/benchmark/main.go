package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"keepair/pkg/common"
	"keepair/pkg/primary"
	"keepair/pkg/worker"

	"github.com/gin-gonic/gin"
)

func main() {

	gin.SetMode(gin.ReleaseMode)

	primaryPort := "9000"
	masterNodeURL := fmt.Sprintf("http://0.0.0.0:%s", primaryPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service := primary.NewService()
		if err := service.Run(ctx, primaryPort); err != nil {
			panic(err)
		}
	}()

	numWorkers := 2
	for i := 0; i < numWorkers; i++ {
		workerPort := fmt.Sprintf("%d", 9001+i)
		go func() {
			service := worker.NewService(masterNodeURL)
			if err := service.Run(ctx, workerPort); err != nil {
				panic(err)
			}
		}()
	}

	// wait for workers to register
	time.Sleep(time.Millisecond * 500)

	// set keys
	numObjects := 2_000
	maxConcurrency := 100

	type KV struct {
		Key   string
		Value []byte
	}

	items := make([]KV, 0)
	for i := 0; i < numObjects; i++ {
		numChars := (rand.Int() % 20) + 1
		items = append(items, KV{
			Key:   common.GenerateRandomString(numChars),
			Value: []byte(common.GenerateRandomString(50_000)),
		})
	}

	log.Default().Printf("generated %d items", len(items))

	started := time.Now()

	limiter := make(chan struct{}, maxConcurrency)
	wg := sync.WaitGroup{}
	for i := 0; i < len(items); i++ {
		limiter <- struct{}{}
		wg.Add(1)
		go func(kv KV) {
			defer func() {
				wg.Done()
				<-limiter
			}()

			url := fmt.Sprintf("%s/set/%s", masterNodeURL, kv.Key)
			res, err := http.Post(url, "", bytes.NewReader(kv.Value))
			if err != nil {
				panic(err)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}
			if res.StatusCode != 200 {
				panic(body)
			}
		}(items[i])
	}
	wg.Wait()
	duration := time.Now().Sub(started).Milliseconds()
	log.Default().Printf("done in %dms", duration)

}
