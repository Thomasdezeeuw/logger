// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"fmt"
	"os"
)

func ExampleTags() {
	tags := Tags{"tag1", "tag2"}
	fmt.Print(tags.String())
	// Prints:
	// tag1, tag2
}

func ExampleLogger_Fatal() {
	log, _ := New("App", 1024, os.Stdout)
	defer func() {
		if recv := recover(); recv != nil {
			log.Fatal(Tags{"file.go", "main"}, recv)
		}
	}()
	panic("Oh no!")
	// Logs:
	// 2015-03-01 17:20:52 [FATAL] file.go, main: Oh no!
	// goroutine 1 [running]:
	// github.com/Thomasdezeeuw/logger.(*Logger).Fatal(0xc08200a200,0xc08201fe00)
	// 	/go/src/github.com/Thomasdezeeuw/logger/logger.go:97 +0x8d
	// main.func·001()
	// 	/go/src/github.com/Thomasdezeeuw/logger/_examples/file.go:35 +0xc4
	// main.main()
	// 	/go/src/github.com/Thomasdezeeuw/logger/_examples/file.go:53 +0x2a9
}

func ExampleLogger_Error() {
	log, _ := New("App", 1024, os.Stdout)
	err := errors.New("Some error")
	log.Error(Tags{"file.go", "main"}, err)
	// Logs:
	// 2015-03-01 17:20:52 [ERROR] file.go, main: Some error
}

func ExampleLogger_Info() {
	log, _ := New("App", 1024, os.Stdout)
	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs:
	// 2015-03-01 17:20:52 [INFO ] file.go, main: My info message
}

func ExampleLogger_Debug() {
	log, _ := New("App", 1024, os.Stdout)
	log.Debug(Tags{"file.go", "main"}, "my %s message", "debug")
	// Logs:
	// 2015-03-01 17:20:52 [DEBUG] file.go, main: My debug message
}

func ExampleGet() {
	// First create a logger, for example in the main init function.
	_, err := NewFile("File", "./application.log", 1024)
	if err != nil {
		panic(err)
	}

	// Then get the logger somewhere else.
	log, err := Get("File")
	if err != nil {
		panic(err)
	}
	log.Info(Tags{"file.go", "main"}, "Written to application.log")
}

func ExampleCombine() {
	bufSize := 1024
	fileLog, err := NewFile("File", "./application.log", bufSize)
	if err != nil {
		panic(err)
	}

	var logs []*Logger
	if production := true; !production {
		// In none production env add a logger to stdout.
		stdLog, err := New("Stdout", bufSize, os.Stdout)
		if err != nil {
			panic(err)
		}

		logs = []*Logger{fileLog, stdLog}
	} else {
		logs = []*Logger{fileLog}
	}

	// Then combine them to log to the stdout and a file.
	log, err := Combine("App", bufSize, logs...)
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "Written to application.log and stdout")
}
