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

	"github.com/jessicagreben/wide-load/pkg/loader"
)

type pluginType int

const (
	http pluginType = 0
	test pluginType = 99
)

var supportedPlugins = map[string]pluginType{
	"http": http,
	"test": test,
}

var (
	concurrency = flag.Int("concurrency", 1, "")
	qps         = flag.Int("qps", 1, "")
	duration    = flag.Duration("duration", 10*time.Second, "")
)

var usage = `Usage: wide-load [options...] <plugin name>

Supported plugin names:
  - http

Options:
  -concurrency  Number of workers (goroutines) to run concurrently. Will never be less than 1. Default is 1.
  -qps          Rate limit in queries per second (QPS) per worker (goroutine). If qps <= 0, then the test runs once. Default is 1 qps per worker.
  -duration     Duration of test. When duration is reached, the test stops and exits. Duration <= 0 will run forever. Default is 1 min. Examples: -duration 10s -duration 3m -duration -1.

Example usage:
$ wide-load http
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
	pluginName := flag.Args()[0]
	if _, ok := supportedPlugins[pluginName]; !ok {
		log.Fatalln("module not supported:", pluginName)
	}

	base, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	pluginPath := path.Join(base, "plugins", pluginName, "suite.so")
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		log.Fatalln("plugin.Open err:", err)
	}
	ts, err := plug.Lookup("TestSuite")
	if err != nil {
		log.Fatalln("lookup", err)
	}
	testsuite, ok := ts.(loader.TestSuite)
	if !ok {
		log.Fatalln("testsuite needs to implement TestSuite interface")
	}

	if *concurrency < 1 {
		*concurrency = 1
	}
	cfg := loader.Config{
		Concurrency: *concurrency,
		QPS:         *qps,
	}
	loadTests := loader.NewTestFramework(cfg, testsuite)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		loadTests.Stop()
	}()
	testDuration := *duration
	if testDuration > 0 {
		go func() {
			time.Sleep(testDuration)
			log.Println("test duration passed, stopping...")
			loadTests.Stop()
		}()
	}
	log.Println("Executing load tests for plugin:", pluginName, "-duration", duration, "-qps", *qps, "-concurrency", *concurrency)
	loadTests.Exec()
}
