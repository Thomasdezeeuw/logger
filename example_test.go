// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"fmt"
	"time"
)

func ExampleTags_String() {
	tags := Tags{"tag1", "tag2"}
	fmt.Print(tags.String())
	// Output:
	// tag1, tag2
}

func ExampleMsg_String() {
	t, _ := time.Parse("2006-01-02 15:04:05", "2015-05-24 17:39:50")
	msg := Msg{Error, "My message", Tags{"tag1", "tag2"}, t}
	fmt.Print(msg.String())
	// Output:
	// 2015-05-24 17:39:50 [Error] tag1, tag2: My message
}

func ExampleLogger_Fatal() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

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
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	err = errors.New("Some error")
	log.Error(Tags{"file.go", "main"}, err)
	// Logs some like:
	// 2015-03-01 17:20:52 [Error] file.go, main: Some error
}

func ExampleLogger_Info() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs:
	// 2015-03-01 17:20:52 [info] file.go, main: My info message
}

func ExampleLogger_Debug() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	log.Debug(Tags{"file.go", "main"}, "my %s message", "debug")
	// Logs:
	// 2015-03-01 17:20:52 [debug] file.go, main: My debug message
}

func ExampleGet() {
	// First create a logger, for example in the main init function.
	log1, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	// Then get the logger somewhere else.
	log2, err := Get("App")
	if err != nil {
		panic(err)
	}

	log1.Info(Tags{"main"}, "Both these messages")
	log2.Info(Tags{"main"}, "are writting to the same logger")
}
