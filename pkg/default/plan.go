package main

import (
	"fmt"

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
	fmt.Println("gw pre test testplan executing")
}

func (t *testPlan) SetupEveryTest() {
	fmt.Println("gw pre test testplan executing")
}

func (t *testPlan) Test() {
	fmt.Println("gw testplan executing")
}

func (t *testPlan) Stop() {
	t.test.Stop()
}

var TestPlan testPlan
