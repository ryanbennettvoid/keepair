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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

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

	t := &http.Transport{}
	t.MaxIdleConnsPerHost = 10
	http.DefaultClient.Transport = t

	limiter := make(chan struct{}, s.MaxConcurrency)
	wg := sync.WaitGroup{}
	for k, v := range items {
		limiter <- struct{}{}
		wg.Add(1)

		go func(key string, value []byte) {
			defer func() {
				<-limiter
				wg.Done()
			}()

			url := fmt.Sprintf("%s/keys/%s", s.MasterNodeURL, key)
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(value))
			checkErr(err)
			res, err := http.DefaultClient.Do(req)
			checkErr(err)
			body, err := io.ReadAll(res.Body)
			checkErr(err)
			if res.StatusCode != 200 {
				panic(string(body))
			}
		}(k, v)

	}
	wg.Wait()

	return items, nil
}
