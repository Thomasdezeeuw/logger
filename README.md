# Logger

[![GoDoc](https://godoc.org/github.com/Thomasdezeeuw/logger?status.svg)](https://godoc.org/github.com/Thomasdezeeuw/logger)
[![Build Status](https://img.shields.io/travis/Thomasdezeeuw/logger.svg)](https://travis-ci.org/Thomasdezeeuw/logger)
[![Coverage Status](https://coveralls.io/repos/Thomasdezeeuw/logger/badge.svg?branch=master&service=github)](https://coveralls.io/github/Thomasdezeeuw/logger?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Thomasdezeeuw/logger)](https://goreportcard.com/report/github.com/Thomasdezeeuw/logger)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/Thomasdezeeuw/logger/blob/master/LICENSE)

Package logger provides asynchronous logging for [Go](https://golang.org/). It
is build for customisation and speed. It uses a custom EventWriter so any custom
backend can be used to store the logs. Logger provides multiple ways to log
information with different levels of importance.

## Installation

Run the following line to install.

```bash
$ go get github.com/Thomasdezeeuw/logger
```

## Usage

The code says it all.

```go
package main

import "github.com/Thomasdezeeuw/logger"

func main() {
	eventWriter := logger.NewConsoleEventWriter(logger.DebugEvent)
	logger.Start(eventWriter)

	// IMPORTANT! Otherwise not all logs will be written!
	defer logger.Close()

	user := "Thomas"
	userId := "1"
	tags := logger.Tags{"README.md", "main", "user:" + userId}
	logger.Infof(tags, "%s says Hi!", user)
	logger.Warn(tags, "We need make this application functional")

	// Output:
	//2015-11-02 21:38:12 [Info] README.md, main, user:1: Thomas says Hi!
	//2015-11-02 21:38:12 [Warn] main: We need make this application functional
}
```

## License

Licensed under the MIT license, copyright (C) Thomas de Zeeuw.
