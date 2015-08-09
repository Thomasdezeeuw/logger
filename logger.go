// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

// Package logger is a asynchronous logging package, hence the name logger. It
// is build for customisation and speed. It uses a custom log writer so any
// custom backend can be used to store the logs. Logger provides multiple ways
// to log information of different levels of importance. No default logger is
// created, but Get is provided to get any logger from any location. See the
// provided examples, both in the documentation and the _examples directory
// (for complete examples).
package logger

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

const (
	defaultStackSize = 8192
	defaultLogsSize  = 1024
)

// MsgWriter takes a msg and writes it to the output.
type MsgWriter interface {
	Write(Msg) error
	Close() error
}

// Collection of all created loggers by name, used by the Get function.
var loggers = map[string]*Logger{}

// The Logger is an logging object which logs to a MsgWriter. Each logging
// operation makes a single call to the Writer's Write method, but not
// necessarily at the same time a Log operation is called. A Logger can be
// used simultaneously from multiple goroutines, it guarantees to serialize
// access to the MsgWriter.
//
// There are six different log levels (from higher to lower): Fatal, Error,
// Warn, Info, Thumb, Debug and Trace.
//
// Note: Log operations (Fatal, Error etc.) don't instally write to the
// MsgWriter, before closing the application call Logger.Close to ensure that
// all log operations are written to the MsgWriter.
type Logger struct {
	Name string

	// All errors, which can be read after Logger.Close is called, NOT THREAT
	// SAFE.
	Errors []error

	mw          MsgWriter
	minLogLevel LogLevel
	logs        chan Msg
	closed      chan struct{}
}

// Fatal logs a recovered error which could have killed the application. Fatal
// adds a stack trace as Msg.Data to the Msg.
func (l *Logger) Fatal(tags Tags, recv interface{}) {
	// Capture the stack trace.
	stackTrace := make([]byte, defaultStackSize)
	n := runtime.Stack(stackTrace, false)
	stackTrace = stackTrace[:n]

	msg := interfaceToString(recv)
	l.logs <- Msg{Fatal, msg, tags, time.Now(), stackTrace}
}

// Error logs a recoverable error.
func (l *Logger) Error(tags Tags, err error) {
	l.logs <- Msg{Error, err.Error(), tags, time.Now(), nil}
}

// Warn logs a warning message.
func (l *Logger) Warn(tags Tags, format string, v ...interface{}) {
	l.logs <- Msg{Warn, fmt.Sprintf(format, v...), tags, time.Now(), nil}
}

// Info logs an informational message.
func (l *Logger) Info(tags Tags, format string, v ...interface{}) {
	l.logs <- Msg{Info, fmt.Sprintf(format, v...), tags, time.Now(), nil}
}

// Debug logs the lowest level of information, only usefull when debugging
// the application. Only shows when Logger.ShowDebug is set to true, which
// defaults to false.
func (l *Logger) Debug(tags Tags, format string, v ...interface{}) {
	l.logs <- Msg{Debug, fmt.Sprintf(format, v...), tags, time.Now(), nil}
}

// Thumbstone indicates a function is still used in production. When developing
// software it's possible to introduce dead code with updates and new features.
// If a function is being suspected of being dead (not used) in production, add
// a call to Thumbstone and check the production logs to see if you're right.
//
// The caller of the (possibly) dead function will be put in the message, using
// the following format:
//	Function functionName called by callerFunctionName, from file /path/to/file on line lineNumber
// For example:
//	Function myFunction called by main.main, from file /main.go on line 20
func (l *Logger) Thumbstone(tags Tags, functionName string) {
	var msg string

	// Get caller information.
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		fn := runtime.FuncForPC(pc)
		msg = fmt.Sprintf("Function %s called by %s, from file %s on line %d",
			functionName, fn.Name(), file, line)
	} else {
		msg = "Function " + functionName + " called from unkown location"
	}

	l.logs <- Msg{Thumb, msg, tags, time.Now(), nil}
}

// Message logs the given message.
//
// Note: the timestamp is always set to  the time of calling the function.
func (l *Logger) Message(msg Msg) {
	msg.Timestamp = time.Now()
	l.logs <- msg
}

// SetMinLogLevel sets the minimum log level to log. See the order of the log
// level at the LogLevel constants documentation, any custom log levels created
// will be higher then Fatal.
//
// Note: NOT THREAT SAFE.
func (l *Logger) SetMinLogLevel(min LogLevel) {
	l.minLogLevel = min
}

// Close blocks until all logs are written to the writer. After all logs are
// written it will call Close() on the message writer.
//
// Note: if a log operation is called after Close is called it will panic.
func (l *Logger) Close() error {
	close(l.logs)
	<-l.closed
	if l.mw != nil {
		return l.mw.Close()
	}
	return nil
}

// New creates a new logger, which starts a go routine which writes to the
// message writer, this way the main thread won't be blocked. Name is the name
// of the logger, used in getting the logger, via Get, from any location within
// your code. The logger is thread same and won't block, unless the message
// channel buffer is full.
//
// Because the logging isn't done on the main thread it's possible that the
// program will close before all the log items are written to the writer. It is
// required to call Logger.Close() before closing down the program! Otherwise
// logs might be lost!
//
// After calling Logger.Close(), log.Errors can be accessed to check for any
// writing errors from the log operations. Any call to Logger.Error,Info etc
// will panic!
func New(name string, mw MsgWriter) (*Logger, error) {
	log, err := new(name, mw)
	if err != nil {
		return nil, err
	}

	go logWriter(log)
	return log, nil
}

// Get gets a logger by its name.
func Get(name string) (*Logger, error) {
	log, ok := loggers[name]
	if !ok {
		return nil, errors.New("logger: no logger found with name " + name)
	}
	return log, nil
}

func new(name string, mw MsgWriter) (*Logger, error) {
	if _, ok := loggers[name]; ok {
		return nil, errors.New("logger: name " + name + " already taken")
	}

	log := &Logger{
		Name:   name,
		mw:     mw,
		logs:   make(chan Msg, defaultLogsSize),
		closed: make(chan struct{}, 1), // Can't block.
	}
	loggers[name] = log

	return log, nil
}

// Needs to be run in it's own goroutine, it blocks until log.logs is closed.
func logWriter(log *Logger) {
	for msg := range log.logs {
		if msg.Level < log.minLogLevel {
			continue
		}

		if err := log.mw.Write(msg); err != nil {
			log.Errors = append(log.Errors, err)
		}
	}

	log.closed <- struct{}{}
}

func interfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case []byte:
		return string(v)
	case error:
		return v.Error()
	}
	return fmt.Sprintf("%v", value)
}
