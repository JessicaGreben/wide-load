package main

import (
	"fmt"
	"net/http"
	"time"

	pkgT "github.com/jessicagreben/wide-load/pkg/tester"
)

type testPlan struct {
	test *pkgT.Test
}

func (t *testPlan) Execute(config pkgT.Config) {
	t.test = pkgT.NewTest(config, t)
	t.test.Run()
}

func (t *testPlan) SetupOnce() {
	fmt.Println("http pre test testplan executing")
}

func (t *testPlan) SetupEveryTest() {
	fmt.Println("http pre test testplan executing")
}

func (t *testPlan) Test() int64 {
	start := time.Now().UTC()
	_, err := http.Get(t.test.Config.URL)
	if err != nil {
		fmt.Println(err)
	}
	return time.Since(start).Milliseconds()
}

func (t *testPlan) Stop() {
	t.test.Stop()
}

var TestPlan testPlan
