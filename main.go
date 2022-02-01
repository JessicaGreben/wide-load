package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"plugin"
	"time"

	pkgT "github.com/jessicagreben/wide-load/pkg/loader"
)

const (
	http = 0
)

var supportedModules = map[string]int{
	"http": http,
}

var (
	url         = flag.String("url", "", "")
	concurrency = flag.Int("concurrency", 1, "")
	qps         = flag.Int("qps", 1, "")
	duration    = flag.Duration("duration", time.Minute, "")
)

var usage = `Usage: wide-load [options...] <testModuleName>

Supported plugin module names:
  - http

Options:
  -url          URL to execute load test against for http/https.
  -concurrency  Number of workers (goroutines) to run concurrently. Will never be less than 1. Default is 1.
  -qps          Rate limit in queries per second (QPS) per worker (goroutine). If qps <= 0, then the test runs once. Default is 1 qps per worker.
  -duration     Duration of test. When duration is reached, the test stops and exits. Duration <= 0 will run forever. Default is 1 min. Examples: -duration 10s -duration 3m -duration -1.

Example usage:
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
	moduleName := flag.Args()[0]
	if _, ok := supportedModules[moduleName]; !ok {
		log.Fatalln("module not supported:", moduleName)
	}

	base, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	modulePkg := path.Join(base, "pkg", moduleName, "plan.so")
	plug, err := plugin.Open(modulePkg)
	if err != nil {
		log.Fatalln("plugin.Open err:", err)
	}
	tp, err := plug.Lookup("TestPlan")
	if err != nil {
		log.Fatalln("lookup", err)
	}
	testplan, ok := tp.(pkgT.TestPlan)
	if !ok {
		log.Fatalln("testplan needs to implement TestPlan interface")
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		testplan.Stop()
	}()
	testDuration := *duration
	if testDuration > 0 {
		go func() {
			time.Sleep(testDuration)
			log.Println("test duration passed, stopping...")
			testplan.Stop()
		}()
	}

	if *concurrency < 1 {
		*concurrency = 1
	}
	log.Println("Executing load test:", "-module", moduleName, "-duration", duration, "-qps", *qps, "-concurrency", *concurrency)
	testplan.Execute(pkgT.Config{
		URL:         *url,
		Concurrency: *concurrency,
		QPS:         *qps,
	})
}
