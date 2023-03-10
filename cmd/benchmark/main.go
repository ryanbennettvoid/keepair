package main

import (
	"context"
	"fmt"
	"time"

	"keepair/pkg/log"
	"keepair/pkg/primary"
	"keepair/pkg/seeder"
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
	objectSize := 50_000

	s := seeder.NewSeeder(masterNodeURL, maxConcurrency, objectSize)

	started := time.Now()
	if _, err := s.SeedKVs(numObjects); err != nil {
		panic(err)
	}
	duration := time.Now().Sub(started).Milliseconds()
	log.Get().Printf("done in %dms", duration)

}
