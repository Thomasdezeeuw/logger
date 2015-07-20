// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"errors"

	"github.com/Thomasdezeeuw/logger"
)

const logName = "App"

func init() {
	// Setup a new logger with a name, path to a console and the buffer size.
	log, err := logger.NewConsole(logName)
	if err != nil {
		panic(err)
	}

	log.Debug(logger.Tags{"console.go", "init"}, "Setup new console logger")
}

var log *logger.Logger

func main() {
	var err error
	// Elsewhere in the application we can retrieve the logger by name.
	log, err = logger.Get(logName)
	if err != nil {
		panic(err)
	}

	// IMPORTANT! Because the logger is asynchronous we need to make sure that
	// everything is written to the log.
	defer log.Close()

	defer func() {
		// Log an recoverd error (panic).
		if recv := recover(); recv != nil {
			log.Fatal(logger.Tags{"console.go", "main"}, recv)
		}
	}()

	// Log an error.
	err = doSomething("stuff")
	if err != nil {
		log.Error(logger.Tags{"console.go", "main"}, err)
	}

	// Log an informational message.
	address := "localhost:8080"
	log.Info(logger.Tags{"console.go", "main"}, "Listening on address %s", address)

	unusedFunction()
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"console.go", "doSomething"}, "doSomething(%q) called", str)

	return errors.New("oops")
}

func unusedFunction() {
	// Log thumbstone, to see if the function is used in production.
	log.Thumbstone(logger.Tags{"console.go"}, "unusedFunction")

	// It's unused for a reason.
	panic("Oh no!")
}
