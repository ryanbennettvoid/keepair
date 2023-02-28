package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"keepair/pkg/log"
	"keepair/pkg/primary"
	"keepair/pkg/primary/node"
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

	log.BigPrintf("adding worker (0)...")

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
	numObjects := 40
	maxConcurrency := 10
	objectSize := 1024

	s := seeder.NewSeeder(masterNodeURL, maxConcurrency, objectSize)

	log.BigPrintf("ading seed data...")
	started := time.Now()
	if _, err := s.SeedKVs(numObjects); err != nil {
		panic(err)
	}
	duration := time.Now().Sub(started).Milliseconds()
	log.BigPrintf("done seeding in %dms", duration)

	log.BigPrintf("adding worker (1)...")

	// start second worker node to trigger rebalance
	go func() {
		workerPort := "9002"
		service := worker.NewService(masterNodeURL)
		if err := service.Run(ctx, workerPort); err != nil {
			panic(err)
		}
	}()

	_ = worker0ID

	time.Sleep(time.Second * 2)

	log.BigPrintf("removing worker (0)...")
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
		res, err := http.Get("http://0.0.0.0:9000/nodes")
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		var nodes struct {
			Nodes []node.Node `json:"nodes"`
		}
		err = json.Unmarshal(body, &nodes)
		if err != nil {
			panic(err)
		}
		log.BigPrintf("NODES: %+v", nodes)
	}

	block := make(chan struct{})
	<-block
}
