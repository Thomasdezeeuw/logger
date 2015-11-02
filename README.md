# Logger

[![GoDoc](https://godoc.org/github.com/Thomasdezeeuw/logger?status.svg)](https://godoc.org/github.com/Thomasdezeeuw/logger)
[![Build Status](https://travis-ci.org/Thomasdezeeuw/logger.png?branch=master)](https://travis-ci.org/Thomasdezeeuw/logger)

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
	// IMPORTANT! Otherwise not all logs will be written!
	defer logger.Close()

	ew := logger.NewConsoleEventWriter()
	logger.Start(ew)

	user := "Thomas"
	userId := "1"
	tags := logger.Tags{"README.md", "main", "user:" + userId}
	logger.Infof(tags, "Hi %s!", user)
	logger.Warn(logger.Tags{"main"}, "We need make this application functional")

	// Output:
	//2015-11-02 21:38:12 [Info] README.md, main, user:1: Hi Thomas!
	//2015-11-02 21:38:12 [Warn] main: We need make this application functional
}
```

## License

Licensed under the MIT license, copyright (C) Thomas de Zeeuw.
