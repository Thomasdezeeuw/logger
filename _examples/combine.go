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
	// Setup a new logger with a name, path to a console and the buffer size.
	log1, err := logger.NewConsole("Console")
	if err != nil {
		panic(err)
	}

	// Setup a new logger with a name, path to a file and the buffer size.
	log2, err := logger.NewFile("File", "./tmp.log")
	if err != nil {
		panic(err)
	}

	// Combine the to logger into a single one, so we log to a file aswell as
	// the console.
	log, err = logger.Combine("App", log1, log2)
	if err != nil {
		panic(err)
	}

	// Show debug messages for both logger.
	// This overwrite the settings in the individual loggers.
	log.ShowDebug = true
}

func main() {
	// IMPORTANT! Otherwise the file will never be written!
	defer log.Close()

	defer func() {
		// Log an recoverd error (panic).
		if recv := recover(); recv != nil {
			log.Fatal(logger.Tags{"file.go", "main"}, recv)
		}
	}()

	// Log an error.
	err := doSomething("stuff")
	if err != nil {
		log.Error(logger.Tags{"file.go", "main"}, err)
	}

	// Log an informational message.
	address := "localhost:8080"
	log.Info(logger.Tags{"file.go", "main"}, "Listening on address %s", address)

	panic("Oh no!")
}

func doSomething(str string) error {
	// Log an debug message.
	log.Debug(logger.Tags{"file.go", "doSomething"}, "doSomething(%q)", str)

	return errors.New("oops")
}

func unusedFunction() {
	// Log thumbstone, to see if the function is used in production.
	log.Thumbstone("unusedFunction in _examples/file.go")
}