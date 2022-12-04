package loader

import "log"

const maxResults = 1000000

// Results are the results of a test.
type Results struct {
	ResultsCh    chan *Result
	TotalTests   int
	SuccessTotal int
	FailureTotal int
	TotalLatency int
	AvgLatency   float64
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
	log.Println("total tests executed:", r.TotalTests)
	if r.TotalTests != 0 {
		r.AvgLatency = float64(r.TotalLatency) / float64(r.TotalTests)
	}
	log.Println("avg latency", r.AvgLatency)
	log.Println("total results", r.TotalTests)
}
