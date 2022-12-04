package main

import (
	"log"

	"github.com/jessicagreben/wide-load/pkg/loader"
)

type testCase struct {
}

func newTestCase() *testCase {
	return &testCase{}
}

func (t *testCase) SetupOnce() {
	log.Println("http setup once")
}

func (t *testCase) Setup() {
	log.Println("http setup")
}

func (t *testCase) Test() error {
	log.Println("http test")
	return nil
}

func (t *testCase) Cleanup() {
	log.Println("http cleanup")
}

type testsuite struct {
	testcases []loader.Testcase
}

func (suite *testsuite) AddTests() int {
	suite.testcases = append(suite.testcases, newTestCase())
	return len(suite.testcases)
}
func (suite *testsuite) Tests() []loader.Testcase {
	return suite.testcases
}
func (suite *testsuite) Exec() {

}
func (suite *testsuite) Stop() {

}

var (
	TestSuite testsuite
)
