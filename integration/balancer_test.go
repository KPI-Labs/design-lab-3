package integration

import (
	"fmt"
	"math/rand"
        "math"
	"net/http"
	"sync"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

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
	mutex    sync.Mutex
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

func worker(wg *sync.WaitGroup, client http.Client, serverCounts *ServerCounts, i int, c *C) {
	defer wg.Done()
	randUrl := randStringBytes(10)
	resp, err := client.Get(fmt.Sprintf("%s/%s", baseAddress, randUrl))
	respServer := resp.Header.Get("lb-from")

	if err != nil {
		c.Errorf("response from [%s]", respServer)
	}

	serverCounts.mutex.Lock()
	serverCounts.counters[respServer] += 1
	serverCounts.mutex.Unlock()
}

func (s *MySuite) TestBalancer(c *C) {
	m := ServerCounts{counters: make(map[string]int)}
	var wg sync.WaitGroup
	counts := 1000
	sigma := 0.05
	for i := 0; i < counts; i++ {
		wg.Add(1)
		go worker(&wg, client, &m, i, c)
	}
	wg.Wait()
	expectedRatio := 1.0 / float64(len(serversPool))
	for _, count := range m.counters {
		diff := math.Abs(expectedRatio - float64(count)/float64(counts))
		c.Assert(diff < sigma, Equals, true)
	}
}

func (s *MySuite) BenchmarkBalancer(c *C) {
	var wg sync.WaitGroup
	for n := 0; n < c.N; n++ {
		wg.Add(1)
		go func(group *sync.WaitGroup) {
			defer wg.Done()
			randUrl := randStringBytes(10)
			_, err := client.Get(fmt.Sprintf("%s/%s", baseAddress, randUrl))
			if err != nil {
				c.Error(err)
			}
		}(&wg)
	}
	wg.Wait()
}
