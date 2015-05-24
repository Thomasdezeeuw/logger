// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/Thomasdezeeuw/logger"
)

type jsonWriter struct {
	f   *os.File
	bw  *bufio.Writer
	enc *json.Encoder
}

func (jw *jsonWriter) Write(msg logger.Msg) error {
	return jw.enc.Encode(msg)
}

func (jw *jsonWriter) Close() error {
	jw.bw.Flush()
	return jw.f.Close()
}

var log *logger.Logger

func init() {
	f, err := os.OpenFile("./tmp.log.json", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	bw := bufio.NewWriter(f)
	mw := &jsonWriter{f, bw, json.NewEncoder(bw)}

	log, err = logger.New("AppJSON", mw)
	if err != nil {
		panic(err)
	}
}

func main() {
	// IMPORTANT! Otherwise no log will be written
	defer log.Close()

	defer func() {
		// Log an recoverd error (panic).
		if recv := recover(); recv != nil {
			log.Fatal(logger.Tags{"json.go", "main"}, recv)
		}
	}()

	// Log an error.
	err := doSomething("stuff")
	if err != nil {
		log.Error(logger.Tags{"json.go", "main"}, err)
	}

	// Log an informational message.
	address := "localhost:8080"
	log.Info(logger.Tags{"json.go", "main"}, "Listening on address %s", address)

	panic(errors.New("Oh no!"))
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"json.go", "doSomething"}, "doSomething(%q)", str)

	return errors.New("oops")
}

func unusedFunction() {
	// Log thumbstone, to see if the function is used in production.
	log.Thumbstone("unusedFunction in _examples/json.go")
}
