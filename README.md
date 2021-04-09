# wide-load
A distributed load tester. Supports adding new load tests via plugins. Written in golang.

## Build

Build each plugin separately for the load tester:
```
$ go build -buildmode=plugin -o pkg/uplink/plan.so pkg/uplink/plan.go
$ go build -buildmode=plugin -o pkg/gatewaymt/plan.so pkg/gatewaymt/plan.go
$ go build
```

## Run load test

```
$ ./wide-load <plugin name>
```

Example:
```
$ ./wide-load uplink

$ ./wide-load gatewayMT
```
