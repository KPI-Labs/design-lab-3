package main

import (
	"math"
	"math/rand"
	"testing"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s *MySuite) TestBalancer(c *C) {
	counts := 10000
	serversCount := make(map[string]int, len(serversPool.pool))

	for _, server := range serversPool.pool {
		serversCount[server] = 0
	}

	for i := 0; i < counts; i++ {
		randURL := randStringBytes(20)
		server := getServerByURL(randURL)
		serversCount[server]++
	}

	expectedRatio := 1.0 / float64(len(serversPool.pool))
	for _, count := range serversCount {
		diff := math.Abs(expectedRatio - float64(count)/float64(counts))
		c.Assert(diff > 0.05, Equals, true)
	}
}
