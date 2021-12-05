package tester

type TestPlan interface {
	Execute(config Config)
	Stop()
}

type TestScenario interface {
	SetupOnce()
	Setup()
	Test() (int64, error)
	Cleanup()
}

// Config is the settings to run a load test.
type Config struct {
	URL         string
	Concurrency int
	QPS         int
}
