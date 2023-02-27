package main

import (
	"context"
	"fmt"
	"time"

	"keepair/pkg/log"
	"keepair/pkg/primary"
	"keepair/pkg/seeder"
	"keepair/pkg/worker"
)

func main() {

	// gin.SetMode(gin.ReleaseMode)

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

	// start first worker node
	go func() {
		workerPort := "9001"
		service := worker.NewService(masterNodeURL)
		if err := service.Run(ctx, workerPort); err != nil {
			panic(err)
		}
	}()

	// wait for workers to register
	time.Sleep(time.Millisecond * 500)

	// set keys
	numObjects := 100
	maxConcurrency := 100
	objectSize := 50_000

	s := seeder.NewSeeder(masterNodeURL, maxConcurrency, objectSize)

	started := time.Now()
	if _, err := s.SeedKVs(numObjects); err != nil {
		panic(err)
	}
	duration := time.Now().Sub(started).Milliseconds()
	log.Get().Printf("done in %dms", duration)

	// start second worker node to trigger rebalance
	go func() {
		workerPort := "9002"
		service := worker.NewService(masterNodeURL)
		if err := service.Run(ctx, workerPort); err != nil {
			panic(err)
		}
	}()

	log.Get().Printf("added new worker")

	block := make(chan struct{})
	<-block
}
