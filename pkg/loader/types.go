package loader

type TestSuite interface {
	AddTests() int
	Tests() []Testcase
	Exec()
	Stop()
}
type Testcase interface {
	SetupOnce()
	Setup()
	Test() error
	Cleanup()
}

// Config contains the settings to run a load test.
type Config struct {
	Concurrency int
	QPS         int
}
