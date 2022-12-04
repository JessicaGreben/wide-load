package loader

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

type testFramework struct {
	cfg Config
	// TODO: add the ability to execute a list of TestSuites
	suite    TestSuite
	start    time.Time
	results  []Results
	initOnce sync.Once
	stopChs  []chan struct{}
}

func NewTestFramework(config Config, suite TestSuite) *testFramework {
	n := suite.AddTests()
	t := testFramework{
		cfg:     config,
		suite:   suite,
		start:   time.Now().UTC(),
		results: make([]Results, n),
		stopChs: make([]chan struct{}, n),
	}
	t.initOnce.Do(func() {
		for i := range t.results {
			t.stopChs[i] = make(chan struct{}, t.cfg.Concurrency)
			t.results[i].ResultsCh = make(chan *Result, maxResults)
		}
	})
	return &t
}

func (t *testFramework) Exec() {
	for testID, test := range t.suite.Tests() {
		go func() {
			t.results[testID].Process()
		}()
		// close the results channel once test is done
		defer close(t.results[testID].ResultsCh)

		var wg sync.WaitGroup
		wg.Add(t.cfg.Concurrency)

		for workerID := 0; workerID < t.cfg.Concurrency; workerID++ {
			// sleep for a random amount of time so that workers don't start at the same time.
			sleepRandom(workerID)
			go func(id, testid int) {
				t.runWorker(id, testid, test)
				wg.Done()
			}(workerID, testID)
		}
		wg.Wait()
		// Process results
		t.results[testID].Report()
	}
}

func (t *testFramework) runWorker(workerId int, testID int, test Testcase) {
	var ticker *time.Ticker
	if t.cfg.QPS > 0 {
		x := time.Duration(1e6/(t.cfg.QPS)) * time.Microsecond
		if x < 0 {
			log.Fatalln("duration less than 0", x)
		}
		ticker = time.NewTicker(x)
		defer ticker.Stop()
	} else {
		// run test once if QPS is less than 1
	}

	test.SetupOnce()
	for {
		select {
		case <-t.stopChs[testID]:
			return
		case <-ticker.C:
			test.Setup()
			start := time.Now()
			err := test.Test()
			latency := time.Since(start)
			t.results[testID].ResultsCh <- &Result{
				LatencyMs: latency.Milliseconds(),
				Failure:   err != nil,
			}
		}
	}
}

// cleanup all the testcases
// close channels
func (t *testFramework) cleanup() error {
	var wg sync.WaitGroup
	for _, test := range t.suite.Tests() {
		wg.Add(1)
		go func(testcase Testcase) {
			testcase.Cleanup()
			wg.Done()
		}(test)
	}
	wg.Wait()
	return nil
}

func (t *testFramework) Stop() {
	for testID := range t.suite.Tests() {
		for i := 0; i < 10; i++ {
			x := struct{}{}
			t.stopChs[testID] <- x
		}
	}
	if err := t.cleanup(); err != nil {
		fmt.Println("framework cleanup err:", err)
	}
	os.Exit(0)
}

// sleepRandom sleeps for a random amount of time between [100,1000) milliseconds.
func sleepRandom(concurrency int) {
	r := rand.Intn(((concurrency + 1) * 100) % 1000)
	time.Sleep(time.Duration(r) * time.Millisecond)
}
