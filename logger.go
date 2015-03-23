// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

// TODO: DRY New, NewMsgWriter & Combine functions.

// Package logger provides multiple ways to log information of different level
// of importance. No default logger is created, but Get is provided to get any
// logger at any location.
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

// MsgWriter takes a msg and writes it to the output.
type MsgWriter interface {
	WriteMsg(Msg) (int, error)
}

// Flusher interface to check if the writer can flush.
type flusher interface {
	Flush() error
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

// Tags are keywords usefull in searching logs. Examples of these are
// - "file.go", "myFn"; indicating the location of the log operation.
// - "user:$id"; indicating a user is logged in (usefull in user specific bugs)
type Tags []string

// String creates a comma separated list from the tags in string.
func (tags *Tags) String() string {
	return string(tags.Bytes())
}

// Bytes creates a comma separated list from the tags in bytes.
func (tags *Tags) Bytes() []byte {
	buf := make([]byte, 0, 50)

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
	// Time the log operation is called.
	Timestamp time.Time

	// Level of the log operation.
	Level string

	// The tags from the log operation.
	Tags Tags

	// The message that needs to be logged.
	Msg string
}

// String creates a string message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2...: message
func (msg *Msg) String() string {
	return string(msg.Bytes())
}

// Bytes formats a message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2...: message
func (msg *Msg) Bytes() []byte {
	buf := make([]byte, 0, 100)

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
	// Format: "[LEVEL] " (levle is always 5 characters long).
	buf = append(buf, '[')
	buf = append(buf, msg.Level...)
	buf = append(buf, ']')
	buf = append(buf, ' ')

	// Write the tags.
	// Format: "tag1, tag2, etc...: ".
	buf = append(buf, msg.Tags.Bytes()...)
	buf = append(buf, ':')
	buf = append(buf, ' ')

	// The actual message.
	buf = append(buf, strings.TrimSpace(msg.Msg)...)
	buf = append(buf, '\n')

	return buf
}

// Collection of all created logger by name, used by the Get function.
var loggers = map[string]*Logger{}

// The Logger is an logging object which logs to an io.Writer where each Write
// call is a single log item. Each logging operation makes a single call to the
// Writer's Write method, but not necessarily at the same time a Log operation
// (Fatal, Error etc.) is called. A Logger can be used simultaneously from
// multiple goroutines, it guarantees to serialize access to the Writer.
// Messages will always be in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2...: message
//
// There are four different log levels (from higher to lower): Fatal, Error,
// Info and Debug, aswell as Thumbstone which is a special case. Thumbstone is
// used for testing if a function is used in production.
//
// Note: Log operations (Fatal, Error etc.) don't instally write to the
// io.Writer, before closing the program call Logger.Close to ensure that all
// log items are written to the io.Writer.
type Logger struct {
	// Name of the logger, used in getting the logger.
	Name string

	// Wether or not to show debug log statements, should really only be set on
	// creating of the logger.
	ShowDebug bool

	// The writers where the items are written to. Either w or wMsg is used, the
	// one not used should be nil.
	w    io.Writer
	wMsg MsgWriter

	// The log messages channel, it's used for actually writing the messages.
	logs chan Msg

	// Indicating the writer closed, having a possible flush or close error.
	closed chan error
}

// Fatal logs a recovered error which could have killed the program.
func (l *Logger) Fatal(tags Tags, recv interface{}) {
	// Capture the stack trace and drop null bytes (they'll show up as spaces).
	buf := make([]byte, 8192)
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
	l.logs <- Msg{time.Now(), "FATAL", tags, item}
}

// Error logs a recoverable error.
func (l *Logger) Error(tags Tags, err error) {
	l.logs <- Msg{time.Now(), "ERROR", tags, err.Error()}
}

// Info logs an informational message.
func (l *Logger) Info(tags Tags, format string, v ...interface{}) {
	l.logs <- Msg{time.Now(), "INFO ", tags, fmt.Sprintf(format, v...)}
}

// Debug logs the lowest level of information, only usefull when debugging
// the application. Only shows when Logger.ShowDebug is set to true, which
// defaults to false.
func (l *Logger) Debug(tags Tags, format string, v ...interface{}) {
	if l.ShowDebug {
		l.logs <- Msg{time.Now(), "DEBUG", tags, fmt.Sprintf(format, v...)}
	}
}

// Thumbstone indicates a function is still used in production. When developing
// software it's possible to introduce dead code with updates and new features.
// If a function is being suspected of being dead (not used) in production, add
// a call to Thumbstone and check the production logs to see if you're right.
func (l *Logger) Thumbstone(item string) {
	l.logs <- Msg{time.Now(), "THUMB", Tags{"thumbstone"}, item}
}

// Close blocks until all logs are written to the writer. It will call Flush()
// and Close() on the writer if supported. If either of the functions returns
// an error it will be returned by this function (Flush error first).
//
// Note: it will block forever if logs are still being written, so it's
// required to first stop writing logs and then call Logger.Close.
func (l *Logger) Close() error {
	// First send to the go routine that we need to close, then wait for the go
	// routine to close.
	l.closed <- nil

	// Wait for an response with a possible error.
	err := <-l.closed

	// Either try to flush the io writer or the message writer. Also try to close
	// the writer.
	// TODO: improve thise code.
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

	return err
}

// New creates a new logger, which starts a go routine which writes to the
// writer. This way the main thread won't be blocked.
//
// Name is the name of the logger used in getting (via Get) from any location.
// The size is the buffer size in number of log item, generally 1000 should
// suffice. But once the buffer is full log operations will block when called.
//
// Note: because the logging isn't done on the main thread it's possible that
// the program will close before all the log items are written to the writer.
// It is required to call logger.Close before closing down the program!
// Otherwise logs might be lost!
func New(name string, size int, w io.Writer) (*Logger, error) {
	log, err := newLogger(name, size, w)
	if err != nil {
		return nil, err
	}

	go func(log *Logger) {
		var closed bool

		for {
			select {
			case <-log.closed: // Indicator that the logger is closing.
				closed = true
			case msg := <-log.logs: // Received a log message.
				// TODO: handle error!
				log.w.Write(msg.Bytes())
			case <-time.After(50 * time.Millisecond):
				// After a timeout we check if we need to close the logger.
				if closed {
					// We logged all the log items, and we can let the Close function
					// know we're done, without an error.
					log.closed <- nil
					return
				}
			}
		}
	}(log)

	return log, nil
}

// NewMsgWriter creates a new logger, similar to New but with a message writer.
// A message writer takes a Msg, which is usefull for writer which don't write
// to a text output.
func NewMsgWriter(name string, size int, w MsgWriter) (*Logger, error) {
	// Create a regular logger with nil as an io.Writer
	log, err := newLogger(name, size, nil)
	if err != nil {
		return nil, err
	}

	// Then set our message writer.
	log.wMsg = w

	go func(log *Logger) {
		var closed bool

		for {
			select {
			case <-log.closed: // Indicator that the logger is closing.
				closed = true
			case msg := <-log.logs: // Received a log message.
				// TODO: handle error!
				log.wMsg.WriteMsg(msg)
			case <-time.After(50 * time.Millisecond):
				// After a timeout we check if we need to close the logger.
				if closed {
					// We logged all the log items, and we can let the Close function
					// know we're done, without an error.
					log.closed <- nil
					return
				}
			}
		}
	}(log)

	return log, nil
}

// NewFile creates a new logger that logs to a file, it uses bufio to buffer
// the writes.
func NewFile(name, path string, size int) (*Logger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create a new writer that writes and flushes the bufio.Writer, but closed
	// the os.File.
	w := &fileWriter{bufio.NewWriter(file), file}
	return New(name, size, w)
}

// Get gets a logger by its name.
func Get(name string) (*Logger, error) {
	log, ok := loggers[name]
	if !ok {
		return nil, errors.New("logger: no logger found with name " + name)
	}

	return log, nil
}

// Combine combines multiple loggers. It's advised to use the same size for all
// loggers.
func Combine(name string, size int, logs ...*Logger) (*Logger, error) {
	if len(logs) == 0 {
		return nil, errors.New("logger: Combine requires atleast one logger")
	}

	log, err := newLogger(name, size, nil)
	if err != nil {
		return nil, err
	}

	go func(log *Logger, logs []*Logger) {
		var closed bool

		for {
			select {
			case <-log.closed: // Indicator that the logger is closing.
				closed = true
			case msg := <-log.logs: // Received an log message.
				for _, log := range logs {
					log.logs <- msg
				}
			case <-time.After(50 * time.Millisecond):
				// After a timeout we check if we need to close the logger.
				if closed {
					i := len(logs)
					errChan := make(chan error, i)

					// Close all underlying loggers.
					for _, log := range logs {
						go func(log *Logger) {
							errChan <- log.Close()
						}(log)
					}

					// Wait for all underlying loggers to respond.
					var err error
					for i > 0 {
						// Check for the error, if no error has happend yet we'll use this
						// error.
						if lErr := <-errChan; err == nil && lErr != nil {
							err = lErr
						}
						i--
					}

					// Return a possible error.
					log.closed <- err
					return
				}
			}
		}
	}(log, logs)

	return log, nil
}

// newLogger creates a new logger and add it to the loggers map. It only
// returns an error if the name is already used.
func newLogger(name string, size int, w io.Writer) (*Logger, error) {
	if _, ok := loggers[name]; ok {
		return nil, errors.New("logger: name " + name + " already taken")
	}

	log := &Logger{
		Name:   name,
		w:      w,
		logs:   make(chan Msg, size),
		closed: make(chan error), // needs to block!
	}

	loggers[name] = log
	return log, nil
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
