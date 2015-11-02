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

var log *logger.Logger

func init() {
	var err error
	mw := newJSONWriter()
	log, err = logger.New("AppJSON", &mw)
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
	log.Thumbstone(logger.Tags{"json.go"}, "unusedFunction")
}

type jsonLog struct {
	Level     string
	Message   string
	Tags      string
	Timestamp string
	Data      string
}

type jsonWriter struct {
	f   *os.File
	w   *bufio.Writer
	enc *json.Encoder
}

func (jw *jsonWriter) Write(msg logger.Msg) error {
	json := jsonLog{
		Level:     msg.Level.String(),
		Message:   msg.Msg,
		Tags:      msg.Tags.String(),
		Timestamp: msg.Timestamp.UTC().Format(logger.TimeFormat),
	}

	if msg.Level == logger.Fatal {
		stacktrace, ok := msg.Data.([]byte)
		if ok {
			json.Data = string(stacktrace)
		}
	}

	return jw.enc.Encode(json)
}

func (jw *jsonWriter) Close() error {
	flushErr := jw.w.Flush()
	err := jw.f.Close()
	if err == nil {
		err = flushErr
	}
	return err
}

func newJSONWriter() jsonWriter {
	f, err := os.OpenFile("./tmp.log.json", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)
	return jsonWriter{f, w, json.NewEncoder(w)}
}
