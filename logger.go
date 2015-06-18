// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

// Package logger provides multiple ways to log information of different level
// of importance. No default logger is created, but Get is provided to get any
// logger at any location. See the provided examples, both in the documentation
// and the _examples directory (for complete examples).
package logger

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// The log operation levels.
const (
	FatalLevel = "FATAL"
	ErrorLevel = "ERROR"
	InfoLevel  = "INFO "
	DebugLevel = "DEBUG"
	ThumbLevel = "THUMB"
)

const (
	defaultStackSize = 4096
	defaultLogsSize  = 1024
)

const (
	fileFlag       = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	filePermission = 0644
)

// MsgWriter takes a msg and writes it to the output.
type MsgWriter interface {
	Write(Msg) error
	Close() error
}

// Collection of all created loggers by name, used by the Get function.
var loggers = map[string]*Logger{}

// The Logger is an logging object which logs to an io.Writer or MsgWriter.
// Each logging operation makes a single call to the Writer's Write method, but
// not necessarily at the same time a Log operation is called. A Logger can be
// used simultaneously from multiple goroutines, it guarantees to serialize
// access to the Writer. Messages (for the io.Writer) will always be in the
// following format (where level is always 5 characters long):
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
//
// There are four different log levels (from higher to lower): Fatal, Error,
// Info and Debug, aswell as Thumbstone which is a special case. Thumbstone is
// used for testing if a function is called in production.
//
// Note: Log operations (Fatal, Error etc.) don't instally write to the
// io.Writer, before closing the program call Logger.Close to ensure that all
// log operations are written to the io.Writer or MsgWriter.
type Logger struct {
	Name      string
	ShowDebug bool
	Errors    []error
	mw        MsgWriter
	logs      chan Msg
	closed    chan struct{}
}

// Fatal logs a recovered error which could have killed the program.
func (l *Logger) Fatal(tags Tags, recv interface{}) {
	// Capture the stack trace and drop null bytes (they'll show up as spaces).
	buf := make([]byte, defaultStackSize)
	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")

	// Try to make some sense of the recoverd value.
	var item string
	switch v := recv.(type) {
	case string:
		item = v
	case error:
		item = v.Error()
	default:
		item = fmt.Sprintf("%v", recv)
	}

	item += "\n" + string(buf)
	l.logs <- Msg{FatalLevel, item, tags, time.Now()}
}

// Error logs a recoverable error.
func (l *Logger) Error(tags Tags, err error) {
	l.logs <- Msg{ErrorLevel, err.Error(), tags, time.Now()}
}

// Info logs an informational message.
func (l *Logger) Info(tags Tags, format string, v ...interface{}) {
	l.logs <- Msg{InfoLevel, fmt.Sprintf(format, v...), tags, time.Now()}
}

// Debug logs the lowest level of information, only usefull when debugging
// the application. Only shows when Logger.ShowDebug is set to true, which
// defaults to false.
func (l *Logger) Debug(tags Tags, format string, v ...interface{}) {
	if l.ShowDebug {
		l.logs <- Msg{DebugLevel, fmt.Sprintf(format, v...), tags, time.Now()}
	}
}

// Thumbstone indicates a function is still used in production. When developing
// software it's possible to introduce dead code with updates and new features.
// If a function is being suspected of being dead (not used) in production, add
// a call to Thumbstone and check the production logs to see if you're right.
func (l *Logger) Thumbstone(item string) {
	l.logs <- Msg{ThumbLevel, item, Tags{"thumbstone"}, time.Now()}
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
// required to call logger.Close() before closing down the program! Otherwise
// logs might be lost!
//
// After calling logger.Close(), log.Errors can be accessed to check for any
// writing errors from the log operations.
func New(name string, mw MsgWriter) (*Logger, error) {
	log, err := new(name, mw)
	if err != nil {
		return nil, err
	}

	go logWriter(log)
	return log, nil
}

type fileMsgWriter struct {
	w *bufio.Writer
	f *os.File
}

func (fw *fileMsgWriter) Write(msg Msg) error {
	_, err := fw.w.Write(msg.Bytes())
	return err
}

func (fw *fileMsgWriter) Close() error {
	flushErr := fw.w.Flush()
	err := fw.f.Close()
	if err == nil {
		err = flushErr
	}
	return err
}

// NewFile creates a new logger that writes to a file.
func NewFile(name, path string) (*Logger, error) {
	f, err := os.OpenFile(path, fileFlag, filePermission)
	if err != nil {
		return nil, err
	}

	mw := &fileMsgWriter{bufio.NewWriter(f), f}
	return New(name, mw)
}

type ioWriterMsgWriter struct {
	w io.Writer
}

func (iw *ioWriterMsgWriter) Write(msg Msg) error {
	_, err := iw.w.Write(msg.Bytes())
	return err
}

func (iw *ioWriterMsgWriter) Close() error {
	return nil
}

// Error ouput, usefull for testing.
var stderr io.Writer = os.Stderr

// NewConsole creates a new logger that writes to error output (os.Stderr).
func NewConsole(name string) (*Logger, error) {
	mw := &ioWriterMsgWriter{stderr}
	return New(name, mw)
}

// Get gets a logger by its name.
func Get(name string) (*Logger, error) {
	log, ok := loggers[name]
	if !ok {
		return nil, errors.New("logger: no logger found with name " + name)
	}
	return log, nil
}

// Combine combines multiple loggers into a single logger.
//
// Note: ShowDebug is enable by default and should be set on the individual
// loggers.
func Combine(name string, logs ...*Logger) (*Logger, error) {
	if len(logs) == 0 {
		return nil, errors.New("logger: Combine requires atleast one logger")
	}

	log, err := new(name, nil)
	if err != nil {
		return nil, err
	}
	log.ShowDebug = true

	go combinedLogWriter(log, logs)
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
		if err := log.mw.Write(msg); err != nil {
			log.Errors = append(log.Errors, err)
		}
	}

	log.closed <- struct{}{}
}

// Needs to be run in it's own goroutine, it blocks until log.logs is closed.
func combinedLogWriter(log *Logger, logs []*Logger) {
	j := len(logs)
	for msg := range log.logs {
		for i := 0; i < j; i++ {
			if msg.Level != DebugLevel ||
				(msg.Level == DebugLevel && logs[i].ShowDebug) {
				logs[i].logs <- msg
			}
		}
	}

	// Close all underlying loggers.
	errChan := make(chan error, len(logs))
	for _, log := range logs {
		go func(log *Logger) {
			errChan <- log.Close()
		}(log)
	}

	// Wait for all underlying loggers to respond.
	for i := len(logs); i > 0; i-- {
		err := <-errChan
		if err != nil {
			log.Errors = append(log.Errors, err)
		}
	}

	// Add all underlying errors to the top one.
	for i := 0; i < j; i++ {
		log.Errors = append(log.Errors, logs[i].Errors...)
	}

	log.closed <- struct{}{}
}
