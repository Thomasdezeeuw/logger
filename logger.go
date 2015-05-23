// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

// TODO: DRY New, NewMsgWriter & Combine functions.

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
	"strings"
	"time"
)

const (
	defaultStackSize  = 4096
	defaultTagsSize   = 50
	defaultMsgSize    = 100
	defaultLogsSize   = 1024
	defaultErrorsSize = 10
)

const (
	fileFlag       = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	filePermission = 0644
)

// MsgWriter takes a msg and writes it to the output.
type MsgWriter interface {
	WriteMsg(Msg) error
}

// FileWriter is an struct used in closing the underlying file if using a
// buffered writer.
type fileWriter struct {
	*bufio.Writer
	f *os.File
}

// Close calls Close on the underlying os.File.
func (w *fileWriter) Close() error {
	return w.f.Close()
}

// Tags are keywords usefull in searching logs. Examples of these are:
//	"file.go", "myFn" // indicating the location of the log operation.
//	"user:$id" // indicating a user is logged in (usefull in user specific bugs)
type Tags []string

// String creates a comma separated list from the tags in string.
func (tags *Tags) String() string {
	return string(tags.Bytes())
}

// Bytes creates a comma separated list from the tags in bytes.
func (tags *Tags) Bytes() []byte {
	buf := make([]byte, 0, defaultTagsSize)

	// Add each tag in the form of "tag, "
	for _, tag := range *tags {
		buf = append(buf, tag...)
		buf = append(buf, ',')
		buf = append(buf, ' ')
	}

	// Drop the last ", "
	if len(buf) > 2 {
		buf = buf[:len(buf)-2]
	}

	return buf
}

// Msg is a message created by a log operation.
type Msg struct {
	Level, Msg string
	Tags       Tags
	Timestamp  time.Time
}

// String creates a string message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
func (msg *Msg) String() string {
	return string(msg.Bytes())
}

// Bytes formats a message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
func (msg *Msg) Bytes() []byte {
	buf := make([]byte, 0, defaultMsgSize)

	// Write the date and time.
	// Format: "YYYY-MM-DD HH:MM:SS ".
	year, month, day := msg.Timestamp.Date()
	hour, min, sec := msg.Timestamp.Clock()
	itoa(&buf, year, 4)
	buf = append(buf, '-')
	itoa(&buf, int(month), 2)
	buf = append(buf, '-')
	itoa(&buf, day, 2)
	buf = append(buf, ' ')
	itoa(&buf, hour, 2)
	buf = append(buf, ':')
	itoa(&buf, min, 2)
	buf = append(buf, ':')
	itoa(&buf, sec, 2)
	buf = append(buf, ' ')

	// Write the level.
	// Format: "[LEVEL] " (level is always 5 characters long).
	buf = append(buf, '[')
	buf = append(buf, msg.Level...)
	buf = append(buf, ']')
	buf = append(buf, ' ')

	// Write the tags.
	// Format: "tag1, tag2: ".
	buf = append(buf, msg.Tags.Bytes()...)
	buf = append(buf, ':')
	buf = append(buf, ' ')

	// The actual message.
	buf = append(buf, strings.TrimSpace(msg.Msg)...)
	buf = append(buf, '\n')

	return buf
}

// Cheap integer to fixed-width decimal ASCII. Modified version from the Golang
// logger package.
func itoa(buf *[]byte, i int, wid int) {
	var b [4]byte // only used for year, month and day so 4 is enough.
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
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

	// The writers where the items are written to. Either w or wMsg is used, the
	// one not used should be nil.
	w    io.Writer
	wMsg MsgWriter

	// The log messages channel, it's used for actually writing the messages.
	logs chan Msg

	// Indicating the writer closed, having a possible flush or close error.
	// todo: how to expose these errors?!
	errors chan []error
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
	l.logs <- Msg{"FATAL", item, tags, time.Now()}
}

// Error logs a recoverable error.
func (l *Logger) Error(tags Tags, err error) {
	l.logs <- Msg{"ERROR", err.Error(), tags, time.Now()}
}

// Info logs an informational message.
func (l *Logger) Info(tags Tags, format string, v ...interface{}) {
	l.logs <- Msg{"INFO ", fmt.Sprintf(format, v...), tags, time.Now()}
}

// Debug logs the lowest level of information, only usefull when debugging
// the application. Only shows when Logger.ShowDebug is set to true, which
// defaults to false.
func (l *Logger) Debug(tags Tags, format string, v ...interface{}) {
	if l.ShowDebug {
		l.logs <- Msg{"DEBUG", fmt.Sprintf(format, v...), tags, time.Now()}
	}
}

// Thumbstone indicates a function is still used in production. When developing
// software it's possible to introduce dead code with updates and new features.
// If a function is being suspected of being dead (not used) in production, add
// a call to Thumbstone and check the production logs to see if you're right.
func (l *Logger) Thumbstone(item string) {
	l.logs <- Msg{"THUMB", item, Tags{"thumbstone"}, time.Now()}
}

// Close blocks until all logs are written to the writer. It will call Flush()
// and Close() on the writer if supported. If either of the functions returns
// an error it will be returned by this function (Flush error first).
//
// Note: if a log operation is called after Close is called it will panic.
func (l *Logger) Close() error {
	type flusher interface {
		Flush() error
	}

	close(l.logs)
	errors := <-l.errors

	// Either try to flush the io writer or the message writer. Also try to close
	// the writer.
	// TODO: improve thise code.
	var err error
	if l.w != nil {
		if fw, ok := l.w.(flusher); ok {
			err = fw.Flush()
		}
		if cw, ok := l.w.(io.Closer); ok {
			if closeErr := cw.Close(); err == nil {
				err = closeErr
			}
		}
	} else {
		if fw, ok := l.wMsg.(flusher); ok {
			err = fw.Flush()
		}
		if cw, ok := l.wMsg.(io.Closer); ok {
			if closeErr := cw.Close(); err == nil {
				err = closeErr
			}
		}
	}

	if err == nil && len(errors) >= 1 {
		err = errors[0]
	}

	return err
}

// New creates a new logger, which starts a go routine which writes to the
// writer. This way the main thread won't be blocked. Name is the name of the
// logger used in getting (via Get) from any location within your code.
//
// Note: because the logging isn't done on the main thread it's possible that
// the program will close before all the log items are written to the writer.
// It is required to call logger.Close before closing down the program!
// Otherwise logs might be lost!
func New(name string, w io.Writer) (*Logger, error) {
	log, err := newLogger(name, w)
	if err != nil {
		return nil, err
	}

	go func(log *Logger) {
		var errors []error

		for msg := range log.logs {
			_, err := log.w.Write(msg.Bytes())
			if err != nil {
				errors = append(errors, err)
			}
		}

		log.errors <- errors
	}(log)

	return log, nil
}

// NewMsgWriter creates a new logger, similar to New but with a message writer.
// A message writer takes a Msg, which is usefull for writer which don't write
// to a text output.
func NewMsgWriter(name string, w MsgWriter) (*Logger, error) {
	// Create a regular logger with nil as an io.Writer
	log, err := newLogger(name, nil)
	if err != nil {
		return nil, err
	}

	// Then set our message writer.
	log.wMsg = w

	go func(log *Logger) {
		var errors []error

		for msg := range log.logs {
			err := log.wMsg.WriteMsg(msg)
			if err != nil {
				errors = append(errors, err)
			}
		}

		log.errors <- errors
	}(log)

	return log, nil
}

// NewFile creates a new logger that logs to a file, it uses bufio to buffer
// the writes.
func NewFile(name, path string) (*Logger, error) {
	file, err := os.OpenFile(path, fileFlag, filePermission)
	if err != nil {
		return nil, err
	}

	// Create a new writer that writes and flushes the bufio.Writer, but closed
	// the os.File.
	w := &fileWriter{bufio.NewWriter(file), file}
	return New(name, w)
}

// Get gets a logger by its name.
func Get(name string) (*Logger, error) {
	log, ok := loggers[name]
	if !ok {
		return nil, errors.New("logger: no logger found with name " + name)
	}

	return log, nil
}

// Combine combines multiple loggers.
//
// todo: provide usefull example, using Stdout in development (with Debug
// enabled) and only to a file in production.
func Combine(name string, logs ...*Logger) (*Logger, error) {
	if len(logs) == 0 {
		return nil, errors.New("logger: Combine requires atleast one logger")
	}

	log, err := newLogger(name, nil)
	if err != nil {
		return nil, err
	}

	go func(log *Logger, logs []*Logger) {
		var errors []error

		// Relay our messages to the other loggers.
		for msg := range log.logs {
			for _, l := range logs {
				l.logs <- msg
			}
		}

		errChan := make(chan error, 5)

		// Close all underlying loggers.
		for _, log := range logs {
			go func(log *Logger) {
				errChan <- log.Close()
			}(log)
		}

		// Wait for all underlying loggers to respond.
		for i := len(logs); i > 0; i-- {
			err := <-errChan
			if err != nil {
				errors = append(errors, err)
			}
		}

		log.errors <- errors
	}(log, logs)

	return log, nil
}

// newLogger creates a new logger and add it to the loggers map. It only
// returns an error if the name is already used.
func newLogger(name string, w io.Writer) (*Logger, error) {
	if _, ok := loggers[name]; ok {
		return nil, errors.New("logger: name " + name + " already taken")
	}

	log := &Logger{
		Name:   name,
		w:      w,
		logs:   make(chan Msg, defaultLogsSize),
		errors: make(chan []error, defaultErrorsSize),
	}

	loggers[name] = log
	return log, nil
}
