package main

import (
	"errors"

	"github.com/Thomasdezeeuw/logger"
)

var log *logger.Logger

func init() {
	var err error
	// Setup a new logger with a name, path to a file and the buffer size.
	log, err = logger.NewFile("App", "./tmp.log", 1024)
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

	log2.Info(logger.Tags{"file.go", "init"}, "Goes to the same file")
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
