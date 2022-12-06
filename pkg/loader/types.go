package loader

type TestSuite interface {
	Init() error
	AddTests() int
	Tests() []Testcase
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
