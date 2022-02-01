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

	// send down the channel for duration at rate of qps
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

		waitFor := time.Duration(1e6/qps) * time.Microsecond
		fmt.Println("wait for:", waitFor)
		ticker := time.NewTicker(waitFor)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				close(outbound)
				return
			case <-ticker.C:
				fmt.Println("tick")
				outbound <- url
			}
		}
	}()

	return outbound
}

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

func finalStage(in <-chan int) {
	for r := range in {
		fmt.Println("latency (ms):", r)
	}
}
