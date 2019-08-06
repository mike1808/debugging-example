package main

import (
	"github.com/mike1808/debugging-example/pkg/balancer"
	"math"
	"math/rand"
	"time"
)

func workFn(in int) float64 {
	return math.Sin(float64(in))
}

func requester(work chan<- Request) {
	c := make(chan float64)
	for {
		// Kill some time (fake load).
		time.Sleep(time.Duration(rand.Int63n(int64(time.Millisecond))))
		work <- Request{workFn, 12, c} // send request
		<-c                                       // wait for answer
	}
}

func main() {
	work := make(chan Request)
	for i := 0; i < nRequester; i++ {
		go requester(work)
	}
	InitBalancer().balance(work)
}


