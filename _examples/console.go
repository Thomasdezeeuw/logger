// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"errors"

	"github.com/Thomasdezeeuw/logger"
)

var log *logger.Logger

func init() {
	var err error
	// Setup a new logger with a name, path to a console and the buffer size.
	log, err = logger.NewConsole("App")
	if err != nil {
		panic(err)
	}

	// Show debug messages.
	log.ShowDebug = true

	// Elsewhere in the application we can retrieve the logger by name.
	log2, err := logger.Get("App")
	if err != nil {
		panic(err)
	}

	log2.Info(logger.Tags{"console.go", "init"}, "Goes to the same console")
}

func main() {
	// IMPORTANT! Otherwise the console will never be written!
	defer log.Close()

	defer func() {
		// Log an recoverd error (panic).
		if recv := recover(); recv != nil {
			log.Fatal(logger.Tags{"console.go", "main"}, recv)
		}
	}()

	// Log an error.
	err := doSomething("stuff")
	if err != nil {
		log.Error(logger.Tags{"console.go", "main"}, err)
	}

	// Log an informational message.
	address := "localhost:8080"
	log.Info(logger.Tags{"console.go", "main"}, "Listening on address %s", address)

	panic("Oh no!")
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"console.go", "doSomething"}, "doSomething(%q)", str)

	return errors.New("oops")
}

func unusedFunction() {
	// Log thumbstone, to see if the function is used in production.
	log.Thumbstone("unusedFunction in _examples/console.go")
}
