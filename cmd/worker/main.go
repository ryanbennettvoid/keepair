package main

import (
	"context"

	"keepair/pkg/common"
	"keepair/pkg/worker"
)

func main() {

	port := common.MustGetEnv("PORT")
	masterNodeURL := common.MustGetEnv("MASTER_NODE_URL")

	service := worker.NewService(masterNodeURL)

	if err := service.Run(context.Background(), port); err != nil {
		panic(err)
	}

}
