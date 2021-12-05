package tester

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// TestFramework is an instance of a test.
type TestFramework struct {
	test     TestScenario
	Config   Config
	Start    time.Time
	Results  Results
	InitOnce sync.Once
	StopCh   chan struct{}
}

func NewTestFramework(config Config, tp TestScenario) *TestFramework {
	t := TestFramework{
		test:    tp,
		Config:  config,
		Start:   time.Now().UTC(),
		Results: Results{},
	}
	t.InitOnce.Do(func() {
		t.StopCh = make(chan struct{}, t.Config.Concurrency)
		t.Results.ResultsCh = make(chan *Result, maxResults)
	})
	return &t
}

func (t *TestFramework) Run() {
	go func() {
		t.Results.Process()
	}()
	// close the results channel once test is done
	defer close(t.Results.ResultsCh)

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

func (t *TestFramework) runWorker(id int) {
	var ticker *time.Ticker
	if t.Config.QPS > 0 {
		x := time.Duration(1e6/(t.Config.QPS)) * time.Microsecond
		if x < 0 {
			log.Fatalln("duration less than 0", x)
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

	t.test.SetupOnce()
	for {
		select {
		case <-t.StopCh:
			return
		case <-ticker.C:
			t.test.Setup()
			latency, err := t.test.Test()
			t.Results.ResultsCh <- &Result{
				LatencyMs: latency,
				Failure:   err != nil,
			}
		}
	}
}

func (t *TestFramework) Stop() {
	for i := 0; i < 10; i++ {
		x := struct{}{}
		t.StopCh <- x
	}
	if err := t.Cleanup(); err != nil {
		fmt.Println("cleanup err:", err)
	}
	os.Exit(0)
}

func (t *TestFramework) Cleanup() error {
	t.test.Cleanup()
	return nil
}

// sleepRandom sleeps for a random amount of time between [100,1000) milliseconds.
func sleepRandom(concurrency int) {
	r := rand.Intn(((concurrency + 1) * 100) % 1000)
	time.Sleep(time.Duration(r) * time.Millisecond)
}
