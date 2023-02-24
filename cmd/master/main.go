package main

import (
	"keepair/pkg/common"
	"keepair/pkg/master"
)

func main() {

	port := common.MustGetEnv("PORT")

	service := master.NewService()

	if err := service.Run(port); err != nil {
		panic(err)
	}

}
