package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"plugin"
	"time"

	pkgT "github.com/jessicagreben/wide-load/pkg/tester"
)

const (
	defaultTest = 0
	uplink      = 1
	gatewayMT   = 2
)

var supportedModules = map[string]int{
	"default":   defaultTest,
	"uplink":    uplink,
	"gatewayMT": gatewayMT,
}

var (
	concurrency = flag.Int("concurrency", 1, "")
	qps         = flag.Int("qps", 1, "")
	duration    = flag.Duration("duration", time.Minute, "")
)

var usage = `Usage: wide-load [options...] <testModuleName>

Module Name options:
  - default
  - uplink
  - gatewayMT

Options:
  -concurrency  Number of workers (goroutines) to run concurrently. Default is 1.
  -qps          Rate limit in queries per second (QPS) per worker. Default is 1 qps per worker.
  -duration     Duration of test. When duration is reached, the test stops and exits. Default is 1 min.
		        Examples: -duration 10s -duration 3m.

Example:
$ wide-load uplink

`

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
	case defaultTest:
		fmt.Println("default not implemented")
		os.Exit(0)
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
	dur := *duration
	if dur > 0 {
		go func() {
			time.Sleep(dur)
			fmt.Println("duration passed, stopping...")
			tp.Stop()
		}()
	}

	tp.Execute(pkgT.Config{
		Concurrency: *concurrency,
		QPS:         *qps,
	})
}
