# wide-load
A distributed load tester. Supports adding new load tests via plugins written in Go.

## Build

Build each plugin separately for the load tester then build the load tester.

For example, build the http plugin and build wide-load:
```
$ go build -o plugins/http/suite.so -buildmode=plugin plugins/http/suite.go
$ go build
```

## Run load test
To see supported plugins:
```
$ ./wide-load --help
```

Run the desired plugin
```
$ ./wide-load <plugin name>
```

Example:
```
$ ./wide-load http

```

## Testing

#### System test

Executing the system test will executes a single testcase that prints for its test.
```
$ go build -o plugins/test/suite.so -buildmode=plugin plugins/test/suite.go
$ go build
$ ./wide-load test
```

## How to create a new load test plugin

To load test a new software system a plugin must be added for that system. To do so, follow these steps: 
1. Create a new directory inside the `plugins` directory. Name the new directory the type of system that will be load tested.
2. Create a file in the new plugins directory called `suite.go`.
3. In the `suite.go` file write code that implements the `TestSuite` interface and the `Testcase` interface defined in `pkg/loader/types.go`. For an example see the existing plugins in `plugins/` directory.
4. In the `suite.go` file, export the implemented testsuite:
```
var (
	TestSuite testsuite
)
```
4. Add the plugin type and the supported plugin at the top of the `main.go` file.
5. Build the plugin code: `go build -buildmode=plugin -o plugins/<name>/plan.so plugins/<name>/`
6. Execute load test for the plugin: `./wide-load <name>`
