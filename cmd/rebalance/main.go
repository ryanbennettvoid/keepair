package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"keepair/pkg/clients"
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
	worker0ID := ""
	go func() {
		workerPort := "9001"
		service := worker.NewService(masterNodeURL)
		worker0ID = service.GetID()
		if err := service.Run(ctx, workerPort); err != nil {
			panic(err)
		}
	}()

	// wait for workers to register
	time.Sleep(time.Millisecond * 500)

	// set keys
	numObjects := 10_000
	maxConcurrency := 10
	objectSize := 1024

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

	log.Get().Printf("added new worker (2)")

	time.Sleep(time.Second * 2)

	// remove first node
	{
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://0.0.0.0:9000/nodes/%s", worker0ID), nil)
		if err != nil {
			panic(err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		if res.StatusCode != 200 {
			panic(fmt.Errorf("delete failed: %s", string(body)))
		}
	}

	// check remaining node object count (should equal total)
	{
		client := clients.NewWorkerClient("http://0.0.0.0:9002")
		stats, err := client.GetStats()
		if err != nil {
			panic(err)
		}
		log.Get().Printf("worker stats: %+v", stats.ObjectCount)
	}

	block := make(chan struct{})
	<-block
}
