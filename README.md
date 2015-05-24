# Logger

[![GoDoc](https://godoc.org/github.com/Thomasdezeeuw/logger?status.svg)](https://godoc.org/github.com/Thomasdezeeuw/logger)
[![Build Status](https://travis-ci.org/Thomasdezeeuw/logger.png?branch=master)](https://travis-ci.org/Thomasdezeeuw/logger)

Logger is a [Go](https://golang.org/) package for logging, build for speed.

## Installation

Run the following line to install.

```bash
$ go get github.com/Thomasdezeeuw/logger
```

## Usage

You can create a logger once and retrieve it everywhere in each pacakge to log
items.

```go
package main

import "github.com/Thomasdezeeuw/logger"

const logName = "App"

func init() {
	// Setup a new logger with a name and log to the console.
	log, err := logger.NewConsole(logName)
	if err != nil {
		panic(err)
	}

	log.Info(logger.Tags{"init", "logger"}, "created a logger here")
}

func main() {
	// Get a logger by its name.
	log, err := logger.Get(logName)
	if err != nil {
		panic(err)
	}

	// IMPORTANT! Otherwise the file will never be written!
	defer log.Close()

	user := "Thomas"
	userId := "1"
	tags := logger.Tags{"README.md", "main", "user:1" + userId}
	log.Info(tags, "Hi %s!", user)
	log.Info(logger.Tags{"main"}, "This get logged to the same logger!")
}
```

## License

Licensed under the MIT license, copyright (C) Thomas de Zeeuw.
