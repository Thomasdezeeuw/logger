// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	msg := Msg{Error, "My message", Tags{"tag1", "tag2"}, t, nil}
	fmt.Print(msg.String())
	// Output:
	// 2015-05-24 17:39:50 [Error] tag1, tag2: My message
}

// Keep in sync with the comment in ExampleMsg_String_data.
type User struct {
	Id   int
	Name string
}

func (u *User) String() string {
	return fmt.Sprintf("User: %s, id: %d", u.Name, u.Id)
}

func ExampleMsg_String_data() {
	// type User struct {
	// 	Id   int
	// 	Name string
	// }
	//
	// func (u *User) String() string {
	// 	return fmt.Sprintf("User: %s, id: %d", u.Name, u.Id)
	// }
	data := User{1, "Thomas"}
	t, _ := time.Parse("2006-01-02 15:04:05", "2015-05-24 17:39:50")
	msg := Msg{Error, "My message", Tags{"tag1", "tag2"}, t, &data}
	fmt.Print(msg.String())
	// Output:
	// 2015-05-24 17:39:50 [Error] tag1, tag2: My message, User: Thomas, id: 1
}

func ExampleNewLogLevel() {
	myLogLevel := NewLogLevel("myLogLevel")
	myLogLevel2 := NewLogLevel("myLogLevel2")
	fmt.Println(myLogLevel.String())
	fmt.Println(myLogLevel2.String())
	// Output:
	// myLogLevel
	// myLogLevel2
}

func ExampleLogLevel_String() {
	fmt.Println(Debug.String())
	fmt.Println(Error.String())
	fmt.Println(Info.String())
	fmt.Println(Fatal.String())
	// Output:
	// Debug
	// Error
	// Info
	// Fatal
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
	// 2015-03-01 17:20:52 [Fatal] file.go, main: Oh no!, goroutine 1 [running]:
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
	// Logs:
	// 2015-03-01 17:20:52 [Error] file.go, main: Some error
}

func ExampleLogger_Warn() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	log.Warn(Tags{"file.go", "main"}, "my %s message", "warning")
	// Logs:
	// 2015-03-01 17:20:52 [Warn] file.go, main: My warning message
}

func ExampleLogger_Info() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
}

func ExampleLogger_Debug() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	log.Debug(Tags{"file.go", "main"}, "my %s message", "debug")
	// Logs:
	// 2015-03-01 17:20:52 [Debug] file.go, main: My debug message
}

func ExampleLogger_Thumbstone() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	var myMaybeUnusedFunction = func() bool {
		tags := Tags{"example_test.go"}
		log.Thumbstone(tags, "myMaybeUnusedFunction")

		return false
	}

	myMaybeUnusedFunction()

	// Logs:
	// 2015-03-01 17:20:52 [Thumb] example_test.go: Function myMaybeUnusedFunction
	// called by logger.ExampleLogger_Thumbstone, from file
	// /github.com/Thomasdezeeuw/logger/example_text.go on line 160
}

func ExampleLogger_Message() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	myLogLevel := NewLogLevel("myLogLevel")
	msg := Msg{
		Level: myLogLevel,
		Msg:   "Hi there",
		Tags:  Tags{"example_test.go"},
		// The timestamp gets set by log.Message
	}

	log.Message(msg)
	// Logs:
	// 2015-03-01 17:20:52 [myLogLevel] example_test.go: Hi there
}

func ExampleLogger_SetMinLogLevel() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}
	log.SetMinLogLevel(Info)

	// This debug message will never show.
	log.Debug(Tags{"file.go", "main"}, "my %s message", "debug")

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
}

func ExampleLogger_Close() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	// Make sure we close the logger. This call will block until all the logs are
	// written. After calling Logger.Close and calls to a log operation
	// (Logger.Fatal, .Error etc.) will panic.
	defer log.Close()

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs to stderr:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
}

func ExampleNewConsole() {
	log, err := NewConsole("App")
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs to stderr:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
}

func ExampleNewFile() {
	filepath := filepath.Join(os.TempDir(), "Logger.log")
	log, err := NewFile("App", filepath)
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs to file:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
}

func ExampleCombine() {
	var Production bool

	filepath := filepath.Join(os.TempDir(), "Logger.log")
	fileLog, err := NewFile("FileLog", filepath)
	if err != nil {
		panic(err)
	}
	logs := []*Logger{fileLog}

	// When not in production, we could the output on our console, so we'll
	// create a console logger.
	if !Production {
		consoleLog, err := NewConsole("ConsoleLog")
		if err != nil {
			panic(err)
		}
		logs = append(logs, consoleLog)
	}

	// Now we combine the loggers, in we production only log to the file, but in
	// development we also logs to the console.
	log, err := Combine("App", logs...)
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs to stderr and/or the file:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
}

func ExampleNewWriter() {
	var buf bytes.Buffer
	log, err := NewWriter("App", &buf)
	if err != nil {
		panic(err)
	}

	log.Info(Tags{"file.go", "main"}, "my %s message", "info")
	// Logs to the buffer:
	// 2015-03-01 17:20:52 [Info] file.go, main: My info message
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
