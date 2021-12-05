package tester

import "log"

const maxResults = 1000000

// Results are the results of a test.
type Results struct {
	ResultsCh    chan *Result
	TotalTests   int
	SuccessTotal int
	FailureTotal int
	TotalLatency int
	AvgLatency   int
	Fastest      float64
	Slowest      float64
	Average      float64
}

type Result struct {
	Failure   bool
	LatencyMs int64
}

func (r *Results) Process() {
	for res := range r.ResultsCh {
		r.TotalTests++
		if res.Failure {
			r.FailureTotal++
		} else {
			r.SuccessTotal++
		}
		r.TotalLatency += int(res.LatencyMs)
	}
}

func (r *Results) Report() {
	log.Println("total tests 2:", r.TotalTests)
	if r.TotalTests != 0 {
		r.AvgLatency = r.TotalLatency / r.TotalTests
	}
	log.Println("avg latency", r.AvgLatency)
	log.Println("total results", r.TotalTests)
}
