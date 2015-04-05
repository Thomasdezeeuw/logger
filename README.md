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

import (
	"fmt"
	"os"

	"github.com/Thomasdezeeuw/logger"
)

type msgWriter struct{}

func (sql *msgWriter) WriteMsg(msg logger.Msg) (int, error) {
	return fmt.Printf("%v [%s] %s: %s", msg.Timestamp, msg.Level,
		msg.Tags.String(), msg.Msg)
}

var log *logger.Logger
var log2 *logger.Logger

func init() {
	var err error
	// Setup a new logger with a name, a buffer size and an io.Writer.
	// The name is used to logger.Get(name) the same logger later in other files
	// or packages. The buffer size is used in the channel for log operations.
	log, err = logger.New("Std", 1024, os.Stdout)
	if err != nil {
		panic(err)
	}

	// Or we can create a logger with a special [MsgWriter]("https://godoc.org/github.com/Thomasdezeeuw/logger#MsgWriter").
	w := new(msgWriter)
	log2, err = logger.NewMsgWriter("Special", 1024, w)
	if err != nil {
		panic(err)
	}
}

func main() {
	// IMPORTANT! Otherwise the file will never be written!
	defer log.Close()

	user := "Thomas"
	tags := logger.Tags{"README.md", "main", "user"}
	log.Info(tags, "Hi %s!", user)
	log2.Info(tags, "Hi %s!", user)
}
```

## License

Licensed under the MIT license, copyright (C) Thomas de Zeeuw.