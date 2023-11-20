# logng

[![Go Reference](https://pkg.go.dev/badge/github.com/goinsane/logng.svg)](https://pkg.go.dev/github.com/goinsane/logng)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=goinsane_logng&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=goinsane_logng)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=goinsane_logng&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=goinsane_logng)

**logng** is a Go (Golang) package that provides structured and leveled logging.

## Features

- Leveled logging: FATAL, ERROR, WARNING, INFO, DEBUG
- Verbose support
- Text and JSON output
- Customizable output
- Function and file logging
- Stack trace
- Field support
- Performance optimized

## Installation

You can install **logng** using the `go get` command:

```sh
go get github.com/goinsane/logng/v2
```

## Examples

```go
package main

import (
	"github.com/goinsane/logng/v2"
)

func main() {
	// log by severity and verbosity.
	// default Logger's severity is SeverityInfo.
	// default Logger's verbose is 0.
	logng.Debug("this is debug log. it won't be shown.")
	logng.Info("this is info log.")
	logng.Warning("this is warning log.")
	logng.V(1).Error("this is error log, verbosity 1. it won't be shown.")

	// set stack trace severity
	logng.SetStackTraceSeverity(logng.SeverityWarning)

	// log with fields
	logng.WithFieldKeyVals("user_name", "john", "ip", "1.2.3.4").Info("connected.")

	// logs with stack trace, because the log severity is error
	logng.Error("an error")
}

```

To run any example, please use the command like the following:

```sh
cd examples/example1/
go run *.go
```

## Tests

To run all tests, please use the following command:

```sh
go test -v
```

To run all examples, please use the following command:

```sh
go test -v -run=^Example
```

To run all benchmarks, please use the following command:

```sh
go test -v -run=^Benchmark -bench=.
```

## Contributing

Contributions are welcome! If you find a bug or want to add a feature, please open an issue or create a pull request.

## License

This project is licensed under the [BSD License](LICENSE).
