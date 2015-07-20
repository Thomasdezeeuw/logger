// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package main

import (
	"errors"

	"github.com/Thomasdezeeuw/logger"
)

func init() {
	// Setup a new logger with a name, path to a file and the buffer size.
	log, err := logger.NewFile("App", "./tmp.log")
	if err != nil {
		panic(err)
	}

	log.Info(logger.Tags{"file.go", "init"}, "Create a new file logger")
}

var log *logger.Logger

func main() {
	var err error
	// Elsewhere in the application we can retrieve the logger by name.
	log, err = logger.Get("App")
	if err != nil {
		panic(err)
	}

	// IMPORTANT! Otherwise the file will never be written!
	defer log.Close()

	log.Info(logger.Tags{"file.go", "main"}, "This goes to the same file")

	defer func() {
		// Log an recoverd error (panic).
		if recv := recover(); recv != nil {
			log.Fatal(logger.Tags{"file.go", "main"}, recv)
		}
	}()

	// Log an error.
	if err := doSomething("stuff"); err != nil {
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
	log.Thumbstone(logger.Tags{"file.go"}, "unusedFunction")
}
