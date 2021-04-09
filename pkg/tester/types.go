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
	PreTest()
	Test()
	Stop()
}

// Config is the settings to run a load test.
type Config struct {
	Concurrency int
	QPS         int
}

// Test is an instance of a test.
type Test struct {
	Config        Config
	Start         time.Time
	Results       Results
	InitOnce      sync.Once
	StopCh        chan struct{}
	modulePreTest func()
	moduleTest    func()
}

func NewTest(config Config, moduletest, pretest func()) *Test {
	t := Test{
		Config:        config,
		Start:         time.Now().UTC(),
		Results:       Results{},
		modulePreTest: pretest,
		moduleTest:    moduletest,
	}
	t.InitOnce.Do(func() {
		t.StopCh = make(chan struct{}, t.Config.Concurrency)
	})
	return &t
}

func sleepRandom(concurrency int) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(((concurrency + 1) * 100) % 1000)
	fmt.Println("sleeping for:", time.Duration(r)*time.Millisecond)
	time.Sleep(time.Duration(r) * time.Millisecond)
}

func (t *Test) Run() {
	var wg sync.WaitGroup
	wg.Add(t.Config.Concurrency)

	for workerID := 0; workerID < t.Config.Concurrency; workerID++ {
		// sleep for a random amount so all workers don't start at the exact same time.
		sleepRandom(t.Config.Concurrency)
		go func(id int) {
			t.runWorker(id)
			wg.Done()
		}(workerID)
	}
	wg.Wait()
}

func (t *Test) runWorker(id int) {
	var throttle <-chan time.Time
	if t.Config.QPS > 0 {
		x := time.Duration(1e6/(t.Config.QPS)) * time.Microsecond
		throttle = time.Tick(x)
	}
	defer func() {
		err := t.Cleanup()
		if err != nil {
			fmt.Println("err cleaning up:", err.Error())
		}
	}()

	for {
		// Check if application is stopped. Do not send into a closed channel.
		select {
		case <-t.StopCh:
			return
		default:
			if t.Config.QPS > 0 {
				<-throttle
			}
			t.modulePreTest() // runs any needed setup for test
			t.moduleTest()    // runs the actual test of the module
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
type Results struct{}
