// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/Thomasdezeeuw/logger/internal/util"
)

const (
	defaultStackSize        = 8192
	defaultEventChannelSize = 1024
	maxNWriteErrors         = 5
)

// EventWriter is the backend of the logger package. It takes events and writes
// them to a storage, this can for example be a file, a database or the
// terminal.
type EventWriter interface {
	// Write is called every time a Log operation is called. This call should
	// block until the event is written. Advised is to buffer the events. All
	// calls to Write are synchronous.
	//
	// If an error is returned the event is expected to NOT have been written. The
	// event will written again after the error handler is called. If the
	// EventWriter returns 5 errors in a row the EventWriter is consided to be bad
	// and will be removed from the list of event writers. HandlerError will be
	// called with ErrBadEventWriter if this happens.
	Write(Event) error

	// HandleError is called every time Write returns an error.
	HandleError(error)

	// Close is called on the EventWriter once Close() (on the package) is called.
	Close() error
}

var (
	eventChannel       = make(chan Event, defaultEventChannelSize)
	eventChannelClosed = make(chan struct{}, 1) // Can't block.
	eventWriters       []EventWriter
	started            bool
)

// Start starts the logger package and enables writing to the given
// EventWriters.
func Start(ews ...EventWriter) {
	if started {
		panic("logger: can only Start once")
	} else if len(ews) < 1 {
		panic("logger: need atleast a single EventWriter to write to")
	}

	started = true
	eventWriters = ews

	go writeEvents()
}

// ErrBadEventWriter gets passed to the error handler of an EventWriter after it
// returned too many write errors in a row. After the error handler of the
// EventWriter is called with this error the writer is considered faulty and
// will no longer recive any Events.
var ErrBadEventWriter = fmt.Errorf("EventWriter is bad, %d faulty writes, EventWriter will be dropped", maxNWriteErrors)

// Needs to be run in it's own goroutine, it blocks until eventChannel is
// closed. After eventChannel is closed it sends a signal to eventChannelClosed.
func writeEvents() {
	var wg sync.WaitGroup
	wg.Add(len(eventWriters))

	// Create event sub channels for each EventWriter and start each EventWriter.
	var eventSubChannels = make([]chan Event, len(eventWriters))
	for i, ew := range eventWriters {
		eventSubChannels[i] = make(chan Event, defaultEventChannelSize)
		go startEventWriter(ew, eventSubChannels[i], &wg)
	}

	// Fan out the events to all the sub channels.
	for event := range eventChannel {
		for _, eventSubChannel := range eventSubChannels {
			eventSubChannel <- event
		}
	}

	// Close each sub channel.
	for _, eventSubChannel := range eventSubChannels {
		close(eventSubChannel)
	}

	wg.Wait()
	eventChannelClosed <- struct{}{}
}

// StartEventWriter blocks until the events channel is closed.
func startEventWriter(ew EventWriter, events <-chan Event, wg *sync.WaitGroup) {
	var badEventWriter = false

	for event := range events {
		if badEventWriter {
			// Simply drain the channel.
			// todo: improve this, don't send to the channel anymore if the writer is
			// bad.
			continue
		}

		if err := writeEvent(ew, event); err != nil {
			// At this point the EventWriter is bad and we won't write to it anymore.
			ew.HandleError(err)
			badEventWriter = true
		}
	}

	wg.Done()
}

// WriteEvent tries to write the event to the given EventWriter, it tries it up
// to maxNWriteErrors times. If EventWriter.Write returns an error it gets
// passed to the error handler of the EventWriter. This function either returns
// ErrBadEventWriter or nil as an error.
func writeEvent(ew EventWriter, event Event) error {
	for n := 1; n <= maxNWriteErrors; n++ {
		err := ew.Write(event)
		if err == nil {
			return nil
		}

		// Handle the error and try again.
		ew.HandleError(err)
	}

	return ErrBadEventWriter
}

// Close stops all the Log Operations from being usable, and they will panic if
// used after Close is called. It also closes all EventWriters and returns the
// first returned error. The EventWriters are closed in the order they are
// passed to Start.
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
var now = time.Now

// Debug logs a debug message.
func Debug(tags Tags, msg string) {
	eventChannel <- Event{DebugEvent, now(), tags, msg, nil}
}

// Debugf is a formatted function of Debug.
func Debugf(tags Tags, format string, v ...interface{}) {
	Debug(tags, fmt.Sprintf(format, v...))
}

// Info logs an informational message.
func Info(tags Tags, msg string) {
	eventChannel <- Event{InfoEvent, now(), tags, msg, nil}
}

// Infof is a formatted function of Info.
func Infof(tags Tags, format string, v ...interface{}) {
	Info(tags, fmt.Sprintf(format, v...))
}

// Warn logs a warning message.
func Warn(tags Tags, msg string) {
	eventChannel <- Event{WarnEvent, now(), tags, msg, nil}
}

// Warnf is a formatted function of Warn.
func Warnf(tags Tags, format string, v ...interface{}) {
	Warn(tags, fmt.Sprintf(format, v...))
}

// Error logs an error message.
func Error(tags Tags, err error) {
	eventChannel <- Event{ErrorEvent, now(), tags, err.Error(), nil}
}

// Errorf is a formatted function of Error.
func Errorf(tags Tags, format string, v ...interface{}) {
	Error(tags, fmt.Errorf(format, v...))
}

// Fatal logs a recovered error which could have killed the application. Fatal
// adds a stack trace (type []byte) as Event.Data.
func Fatal(tags Tags, recv interface{}) {
	stackTrace := make([]byte, defaultStackSize)
	n := runtime.Stack(stackTrace, false)
	stackTrace = stackTrace[:n]

	msg := util.InterfaceToString(recv)
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

// Log logs a custom created event.
//
// Note: the timestamp doesn't need to be set, because it will be set by Log.
func Log(event Event) {
	event.Timestamp = now()
	eventChannel <- event
}
