package integration_tests

import "sync"

// testMu is locked at the start of each test
// to prevent stateful race conditions between
// test runs
var testMu sync.Mutex

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
