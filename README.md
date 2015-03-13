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

You can put debug statements everywhere and only show ones you're interested in.

```go
package main

import (
	"os"

	"github.com/Thomasdezeeuw/logger"
)

var log *logger.Logger

func init() {
	var err error
	// Setup a new logger with a name, path to a file and the buffer size.
	log, err = logger.New("Std", 1024, os.Stdout)
	if err != nil {
		panic(err)
	}
}

func main() {
	// IMPORTANT! Otherwise the file will never be written!
	defer log.Close()

	user := "Thomas"
	tags := logger.Tags{"README.md", "main"}
	log.Info(tags, "Hi %s!", user)
}
```

## License

Licensed under the MIT license, copyright (C) Thomas de Zeeuw.