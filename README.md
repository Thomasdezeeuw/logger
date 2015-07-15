# Logger

[![GoDoc](https://godoc.org/github.com/Thomasdezeeuw/logger?status.svg)](https://godoc.org/github.com/Thomasdezeeuw/logger)
[![Build Status](https://travis-ci.org/Thomasdezeeuw/logger.png?branch=master)](https://travis-ci.org/Thomasdezeeuw/logger)

Logger is a asynchronous logging package for [Go](https://golang.org/). It is
build for customisation and speed. It uses a custom log writer so any custom
backend can be used to store the logs.

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

	log.Info(logger.Tags{"init", "logger"}, "created a logger in init function")
}

func main() {
	// Get a logger by its name.
	log, err := logger.Get(logName)
	if err != nil {
		panic(err)
	}

	// IMPORTANT! Otherwise not all logs will be written!
	defer log.Close()

	user := "Thomas"
	userId := "1"
	tags := logger.Tags{"README.md", "main", "user:" + userId}
	log.Info(tags, "Hi %s!", user)
	log.Warn(logger.Tags{"main"}, "We need make this application functional")
}
```

## License

Licensed under the MIT license, copyright (C) Thomas de Zeeuw.
