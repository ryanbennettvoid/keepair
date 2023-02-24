package main

import (
	"context"
	"fmt"
	"time"

	"keepair/pkg/common"
	"keepair/pkg/healthcheck"
	"keepair/pkg/worker"
)

func main() {

	port := common.MustGetEnv("PORT")
	masterNodeURL := common.MustGetEnv("MASTER_NODE_URL")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
	defer cancel()

	masterHealthcheckURL := fmt.Sprintf("%s/health", masterNodeURL)
	healthchecker := healthcheck.NewService(masterHealthcheckURL)
	stopHealthcheck := healthchecker.Start()
	if err := healthchecker.WaitUntilHealthy(ctx); err != nil {
		panic(err)
	}
	stopHealthcheck()

	service := worker.NewService(masterNodeURL)

	if err := service.Run(port); err != nil {
		panic(err)
	}

}
