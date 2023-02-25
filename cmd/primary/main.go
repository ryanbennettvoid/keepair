package main

import (
	"context"

	"keepair/pkg/common"
	"keepair/pkg/primary"
)

func main() {

	port := common.MustGetEnv("PORT")

	service := primary.NewService()

	if err := service.Run(context.Background(), port); err != nil {
		panic(err)
	}

}
