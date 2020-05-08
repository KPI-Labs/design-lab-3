package integration

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
	"math/rand"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

var serversPool = []string{
	"http://localhost:8080",
	"http://localhost:8081",
	"http://localhost:8082",
}

type ServerCounts struct {
	mutex sync.Mutex
	counters map[string]int
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func worker(wg *sync.WaitGroup, client http.Client, serverCounts *ServerCounts, i int, t *testing.T) {
	defer wg.Done()
	randUrl := randStringBytes(10)
	resp, err := client.Get(fmt.Sprintf("%s/%s", baseAddress, randUrl))
	respServer := resp.Header.Get("lb-from")

	if err != nil {
		t.Logf("response from [%s]", respServer)
	}

	serverCounts.mutex.Lock()
	serverCounts.counters[respServer] += 1
	serverCounts.mutex.Unlock()
}

func TestBalancer(t *testing.T) {
	m := ServerCounts{counters: make(map[string]int)}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go worker(&wg, client, &m, i, t)
	}
	wg.Wait()
}

func BenchmarkBalancer(b *testing.B) {
	var wg sync.WaitGroup
	for n := 0; n < b.N; n++ {
		wg.Add(1)
		go func(group sync.WaitGroup) {
			defer wg.Done()
			randUrl := randStringBytes(10)
			_, err := client.Get(fmt.Sprintf("%s/%s", baseAddress, randUrl))
			if err != nil {
				b.Error(err)
			}
		}(wg)
	}
	wg.Wait()
}

