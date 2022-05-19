package loader

import (
	"fmt"
	"net/http"
	"time"
)

const (
	minQPS = 1
	maxQPS = 1000
)

func load(url string, qps, concurrency int, duration time.Duration) {
	fmt.Println("starting load")
	out := firstStage(url, qps, duration)

	results := []<-chan int{}
	for i := 0; i < concurrency; i++ {
		result := secondStage(out)
		results = append(results, result)
	}

	for _, result := range results {
		finalStage(result)
	}
	fmt.Println("finished")
}

// firstStage is the first stage in the pipeline.
// It creates a single outbound channel. Consumers of the outbound channel should
// use the info to make load test requests.
func firstStage(url string, qps int, duration time.Duration) <-chan string {
	outbound := make(chan string)

	go func() {
		if qps <= 0 {
			qps = minQPS
		}
		if qps > maxQPS {
			qps = maxQPS
		}

		done := make(chan bool)
		go func() {
			time.Sleep(duration)
			done <- true
		}()

		ticker := time.NewTicker(time.Duration(1e6/qps) * time.Microsecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				close(outbound)
				return
			case <-ticker.C:
				outbound <- url
			}
		}
	}()

	return outbound
}

// secondStage is the second stage in the pipeline.
// This stage reads from a url from the input and makes an http request
// and returns the latency of that request.
func secondStage(in <-chan string) <-chan int {
	outbound := make(chan int)

	go func() {
		for url := range in {
			start := time.Now()
			if _, err := http.Get(url); err != nil {
				fmt.Println("http.Get err:", err)
			}
			outbound <- int(time.Since(start).Milliseconds())
		}
		close(outbound)
	}()
	return outbound
}

// finalStage is the final stage in the pipeline.
// This stage processes the results by printing them.
func finalStage(in <-chan int) {
	for r := range in {
		fmt.Println("latency (ms):", r)
	}
}
