package tester

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

type TestPlan interface {
	Execute(config Config)
	SetupOnce()
	SetupEveryTest()
	Test() int64
	Stop()
}

// Config is the settings to run a load test.
type Config struct {
	URL         string
	Concurrency int
	QPS         int
}

// Test is an instance of a test.
type Test struct {
	TestPlan
	Config   Config
	Start    time.Time
	Results  Results
	InitOnce sync.Once
	StopCh   chan struct{}
}

func NewTest(config Config, tp TestPlan) *Test {
	t := Test{
		TestPlan: tp,
		Config:   config,
		Start:    time.Now().UTC(),
		Results: Results{
			Latencies: []int64{},
		},
	}
	t.InitOnce.Do(func() {
		t.StopCh = make(chan struct{}, t.Config.Concurrency)
	})
	return &t
}

// sleepRandom sleeps for a random amount of time between [100,1000) milliseconds.
func sleepRandom(concurrency int) {
	r := rand.Intn(((concurrency + 1) * 100) % 1000)
	fmt.Println("sleeping for (ms):", time.Duration(r)*time.Millisecond)
	time.Sleep(time.Duration(r) * time.Millisecond)
}

func (t *Test) Run() {
	var wg sync.WaitGroup
	wg.Add(t.Config.Concurrency)

	for workerID := 0; workerID < t.Config.Concurrency; workerID++ {
		// sleep for a random amount of time so that workers don't start at the same time.
		sleepRandom(workerID)
		go func(id int) {
			t.runWorker(id)
			wg.Done()
		}(workerID)
	}
	wg.Wait()
}

func (t *Test) runWorker(id int) {
	var ticker *time.Ticker
	if t.Config.QPS > 0 {
		x := time.Duration(1e6/(t.Config.QPS)) * time.Microsecond
		if x < 0 {
			// ticket panics if less than 0
		}
		ticker = time.NewTicker(x)
		defer ticker.Stop()
	} else {
		// run test once if QPS is less than 1
	}

	defer func() {
		err := t.Cleanup()
		if err != nil {
			fmt.Println("err cleaning up:", err.Error())
		}
	}()

	t.SetupOnce()
	for {
		select {
		case <-t.StopCh:
			return
		case <-ticker.C:
			t.SetupEveryTest()
			t.Results.Latencies = append(t.Results.Latencies, t.Test())
		}
	}
}

func (t *Test) Stop() {
	for i := 0; i < t.Config.Concurrency; i++ {
		t.StopCh <- struct{}{}
	}
	if err := t.Cleanup(); err != nil {
		fmt.Println("cleanup err:", err)
	}
	os.Exit(0)
}

func (t *Test) Cleanup() error {
	return nil
}

// Results are the results of a test.
type Results struct {
	Latencies []int64
}
