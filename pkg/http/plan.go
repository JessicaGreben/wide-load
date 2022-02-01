package main

import (
	"log"

	pkgT "github.com/jessicagreben/wide-load/pkg/loader"
)

type testPlan struct {
	test *pkgT.TestFramework
}

func (t testPlan) Execute(config pkgT.Config) {
	t.test = pkgT.NewTestFramework(config, &testScenario{test: t.test})
	t.test.Run()
	t.test.Results.Report()
}

func (t *testPlan) Stop() {
	// t.test.Stop()
}

type testScenario struct {
	test *pkgT.TestFramework
}

func (t *testScenario) SetupOnce() {
	log.Println("http setup once executing")
}

func (t *testScenario) Setup() {
	log.Println("http setup executing")
}

func (t *testScenario) Test() (int64, error) {
	log.Println("http cleanup executing")
	return 0, nil
}

func (t *testScenario) Cleanup() {
	log.Println("http cleanup executing")
}

var (
	TestPlan     testPlan
	TestScenario testScenario
)
