package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"plugin"
	"time"

	pkgT "github.com/jessicagreben/wide-load/pkg/tester"
)

const (
	http      = 0
	uplink    = 1
	gatewayMT = 2
)

var supportedModules = map[string]int{
	"http":      http,
	"uplink":    uplink,
	"gatewayMT": gatewayMT,
}

var (
	url         = flag.String("url", "", "")
	concurrency = flag.Int("concurrency", 1, "")
	qps         = flag.Int("qps", 1, "")
	duration    = flag.Duration("duration", time.Minute, "")
)

var usage = `Usage: wide-load [options...] <testModuleName>

Module Name options:
  - http
  - uplink
  - gatewayMT

Options:
  -url          URL to execute load test against for http/https.
  -concurrency  Number of workers (goroutines) to run concurrently. Will never be less than 1. Default is 1.
  -qps          Rate limit in queries per second (QPS) per worker (goroutine). If qps <= 0, then the test only runs once. Default is 1 qps per worker.
  -duration     Duration of test. When duration is reached, the test stops and exits. Duration <= 0 will run forever. Default is 1 min.
		        Examples: -duration 10s -duration 3m -duration -1.

Example:
$ wide-load http -url=https://testme.com

`

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Usage = func() {
		fmt.Print(usage)
	}
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Print(usage)
		os.Exit(0)
	}
	testToExecute, ok := supportedModules[flag.Args()[0]]
	if !ok {
		fmt.Println("test module type not supported:", flag.Args()[0])
		os.Exit(1)
	}

	var mod string
	switch testToExecute {
	case http:
		mod = "./pkg/default/plan.so"
	case gatewayMT:
		mod = "./pkg/gatewaymt/plan.so"
	case uplink:
		mod = "./pkg/uplink/plan.so"
	default:
		fmt.Println("test type not supported:", testToExecute)
		os.Exit(1)
	}

	plug, err := plugin.Open(mod)
	if err != nil {
		fmt.Println("plugin.Open err:", err)
		os.Exit(1)
	}
	testPlan, err := plug.Lookup("TestPlan")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tp, ok := testPlan.(pkgT.TestPlan)
	if !ok {
		fmt.Println("unexpected type from module")
		os.Exit(1)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		tp.Stop()
	}()
	testDuration := *duration
	if testDuration > 0 {
		go func() {
			time.Sleep(testDuration)
			fmt.Println("test duration passed, stopping...")
			tp.Stop()
		}()
	}

	if *concurrency < 1 {
		*concurrency = 1
	}
	tp.Execute(pkgT.Config{
		URL:         *url,
		Concurrency: *concurrency,
		QPS:         *qps,
	})
}
