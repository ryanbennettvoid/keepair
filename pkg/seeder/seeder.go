package seeder

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"

	"keepair/pkg/common"
)

type Seeder struct {
	MasterNodeURL  string
	MaxConcurrency int
	ObjectSize     int
}

func NewSeeder(masterNodeURL string, maxConcurrency, objectSize int) Seeder {
	return Seeder{
		MasterNodeURL:  masterNodeURL,
		MaxConcurrency: maxConcurrency,
		ObjectSize:     objectSize,
	}
}

func (s Seeder) SeedKVs(count int) (map[string][]byte, error) {

	items := make(map[string][]byte)

	for i := 0; i < count; i++ {
		numChars := (rand.Int() % 20) + 10
		key := common.GenerateRandomString(numChars)
		items[key] = []byte(common.GenerateRandomString(s.ObjectSize))
	}

	limiter := make(chan struct{}, s.MaxConcurrency)
	wg := sync.WaitGroup{}
	for k, v := range items {
		limiter <- struct{}{}
		wg.Add(1)

		go func(key string, value []byte) {
			defer func() {
				wg.Done()
				<-limiter
			}()

			url := fmt.Sprintf("%s/keys/%s", s.MasterNodeURL, key)
			res, err := http.Post(url, "", bytes.NewReader(value))
			if err != nil {
				panic(err)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}
			if res.StatusCode != 200 {
				panic(body)
			}
		}(k, v)

	}
	wg.Wait()

	return items, nil
}
