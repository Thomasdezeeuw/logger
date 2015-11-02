// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

const (
	defaultStackSize        = 8192
	defaultEventChannelSize = 1024
)

// EventWriter is the backend of the logger package. It takes events and writes
// them to a backend, this can for example be a file, a database or the console.
type EventWriter interface {
	// todo: improve this doc.
	// Write is called every time a Log operation is called. This call should
	// block until the event is written, but advised is to buffer the events.
	// If an error is returned the event is expected to NOT have been written and
	// the event will be tried to written later on. If the EventWriter returns
	// more then 5 errors in a row we stop writing to it. Because of this it is
	// possible for the event to not be in order if it failed to be written the
	// first time.
	Write(Event) error

	// HandleError is called every time Write returns an error.
	HandleError(error)

	// Close is called on the EventWriter once Close() is called.
	Close() error
}

var (
	eventChannel       chan Event
	eventChannelClosed chan struct{}
	eventWriters       []EventWriter
	started            bool
)

// todo: doc.
func Start(ews ...EventWriter) {
	if started {
		panic("logger: can only Start once")
	}
	started = true

	if len(ews) < 1 {
		panic("logger: need atleast a single EventWriter to write to")
	}

	eventChannel = make(chan Event, defaultEventChannelSize)
	eventChannelClosed = make(chan struct{}, 1) // Can't block.
	eventWriters = ews
	go eventsWriter()
}

var badEventWriterErr = errors.New("EventWriter is bad, more then 5 faulty writes, EventWriter will be dropped")

// Needs to be run in it's own goroutine, it blocks until eventChannel is
// closed. After eventChannel is closed it sends a signal to eventChannelClosed.
// todo: benchmark and improve performance. Perhaps by passing the event writers
// and event channels as reference to this function.
func eventsWriter() {
	// Create a copy of the eventWriters which we can modify. This way we can
	// drop writers from the slice if they return to many write errors.
	var ews = make([]EventWriter, len(eventWriters))
	if n := copy(ews, eventWriters); n != len(eventWriters) {
		panic("Couldn't copy all the EventWriters")
	}

	// Slice of event that tried to be written, but returned an error. This is a
	// per EventWriter slice.
	var badWrites = make([][]Event, len(eventWriters))

	for event := range eventChannel {
		for i, eventWriter := range ews {
			if l := len(badWrites[i]); l != 0 {
				// After 5 bad writes we drop the EventWriter from the slice of
				// EventWriters, aswell as badWrites.
				if l == 5 {
					eventWriter.HandleError(badEventWriterErr)
					ews = append(ews[:i], ews[i+1:]...)
					badWrites = append(badWrites[:i], badWrites[i+1:]...)
					continue
				}

				// Try to rewite a previously failed write.
				if err := eventWriter.Write(badWrites[i][0]); err != nil {
					eventWriter.HandleError(err)
				} else {
					badWrites[i] = badWrites[i][1:]
				}
			}

			err := eventWriter.Write(event)
			if err != nil {
				eventWriter.HandleError(err)
				badWrites[i] = append(badWrites[i], event)
			}
		}
	}

	eventChannelClosed <- struct{}{}
}

// Close stops all the Log Operation from being usable (and they will panic if
// used after close is called). It also closes all EventWriters and returns the
// errors. The EventWriters are closed in the order they are added and the first
// non-nil error is returned.
func Close() error {
	close(eventChannel)
	<-eventChannelClosed

	var err error
	for _, eventWriter := range eventWriters {
		er := eventWriter.Close()
		if er != nil && err == nil {
			err = er
		}
	}
	return err
}

// Subbed for testing.
var now = func() time.Time {
	return time.Now()
}

// Debug logs the lowest level of information, only usefull when debugging
// the application. Only shows when Logger.ShowDebug is set to true, which
// defaults to false.
func Debug(tags Tags, msg string) {
	eventChannel <- Event{DebugEvent, now(), tags, msg, nil}
}

func Debugf(tags Tags, format string, v ...interface{}) {
	Debug(tags, fmt.Sprintf(format, v...))
}

// Info logs an informational message.
func Info(tags Tags, msg string) {
	eventChannel <- Event{InfoEvent, now(), tags, msg, nil}
}

func Infof(tags Tags, format string, v ...interface{}) {
	Info(tags, fmt.Sprintf(format, v...))
}

// Warn logs a warning message.
func Warn(tags Tags, msg string) {
	eventChannel <- Event{WarnEvent, now(), tags, msg, nil}
}

func Warnf(tags Tags, format string, v ...interface{}) {
	Warn(tags, fmt.Sprintf(format, v...))
}

// Error logs a recoverable error.
func Error(tags Tags, err error) {
	eventChannel <- Event{ErrorEvent, now(), tags, err.Error(), nil}
}

func Errorf(tags Tags, format string, v ...interface{}) {
	Error(tags, fmt.Errorf(format, v...))
}

// Fatal logs a recovered error which could have killed the application. Fatal
// adds a stack trace as Msg.Data to the Msg.
func Fatal(tags Tags, recv interface{}) {
	stackTrace := make([]byte, defaultStackSize)
	n := runtime.Stack(stackTrace, false)
	stackTrace = stackTrace[:n]

	msg := interfaceToString(recv)
	eventChannel <- Event{FatalEvent, now(), tags, msg, stackTrace}
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
func Thumbstone(tags Tags, functionName string) {
	var msg string
	if pc, file, line, ok := runtime.Caller(2); ok {
		fn := runtime.FuncForPC(pc)
		msg = fmt.Sprintf("Function %s called by %s, from file %s on line %d",
			functionName, fn.Name(), file, line)
	} else {
		msg = "Function " + functionName + " called from unkown location"
	}

	eventChannel <- Event{ThumbEvent, now(), tags, msg, nil}
}

// Log logs an event.
//
// Note: the timestamp is always set by Log.
func Log(event Event) {
	event.Timestamp = now()
	eventChannel <- event
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
