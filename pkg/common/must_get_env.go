package common

import (
	"fmt"
	"os"
)

func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Errorf("env is not set: %s", key))
	}
	return value
}
