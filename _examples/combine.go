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
	consoleLog, err := logger.NewConsole(logName + "Console")
	if err != nil {
		panic(err)
	}

	// Setup a new logger with a name, path to a file and the buffer size.
	fileLog, err := logger.NewFile(logName+"File", "./tmp.log")
	if err != nil {
		panic(err)
	}

	// On the console we don't want all the debug messages, but in the file log
	// we do want them.
	consoleLog.SetMinLogLevel(logger.Info)
	fileLog.SetMinLogLevel(logger.Debug)

	// Combine the to logger into a single one, so we log to a file aswell as
	// the console.
	log, err = logger.Combine("App", consoleLog, fileLog)
	if err != nil {
		panic(err)
	}

	// This will only showup in the file log.
	log.Debug(logger.Tags{"console.go", "init"}, "Setup console and file logger")
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
