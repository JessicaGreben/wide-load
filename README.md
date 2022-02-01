# wide-load
A distributed load tester. Supports adding new load tests via plugins. Written in golang.

## Build

Build each plugin separately for the load tester:
```
$ go build -buildmode=plugin -o pkg/http/plan.so pkg/http/plan.go
$ go build
```

## Run load test

To see supported plugins:
```
$ ./wide-load --help
```

```
$ ./wide-load <plugin name>
```

Example:
```
$ ./wide-load uplink

$ ./wide-load gatewayMT
```

## How to create a new load test plugin

In order to create a new load test, a package must be created that contains code that implment the `TestScenario` interface.

That will allow the testplan to use the test framework to execute the module's test scenario.
