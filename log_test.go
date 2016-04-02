// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"errors"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

// Time returned in calling now(), setup and test in init.
var t1 = time.Date(2015, 9, 1, 14, 22, 36, 0, time.UTC)

func init() {
	now = func() time.Time {
		return t1
	}
}

// EventWriter that collects the events and errors.
type eventWriter struct {
	events []Event
	errors []error
	closed bool
}

func (ew *eventWriter) Write(event Event) error {
	ew.events = append(ew.events, event)
	return nil
}

func (ew *eventWriter) HandleError(err error) {
	ew.errors = append(ew.errors, err)
}

func (ew *eventWriter) Close() error {
	ew.closed = true
	return nil
}

// A data type to be used in calling Log.
type user struct {
	ID   int
	Name string
}

func TestLog(t *testing.T) {
	defer reset()
	var ew eventWriter
	Start(&ew)

	tags := Tags{"my", "tags"}
	eventType := NewEventType("my-event-type")
	data := user{1, "Thomas"}
	event := Event{
		Type:    eventType,
		Tags:    tags,
		Message: "My event",
		Data:    data,
	}
	recv := getPanicRecoveredValue("Fatal message")

	Debug(tags, "Debug message")
	Debugf(tags, "Debug %s message", "formatted")
	Info(tags, "Info message")
	Infof(tags, "Info %s message", "formatted")
	Warn(tags, "Warn message")
	Warnf(tags, "Warn %s message", "formatted")
	Error(tags, errors.New("Error message"))
	Errorf(tags, "Error %s message", "formatted")
	Fatal(tags, recv)
	testThumstone(tags)
	Log(event)

	if err := Close(); err != nil {
		t.Fatal("Unexpected error closing: " + err.Error())
	}

	if len(ew.errors) != 0 {
		t.Fatalf("Unexpected error(s): %v", ew.errors)
	}

	_, file, _, _ := runtime.Caller(0)

	expected := []Event{
		{Type: DebugEvent, Message: "Debug message"},
		{Type: DebugEvent, Message: "Debug formatted message"},
		{Type: InfoEvent, Message: "Info message"},
		{Type: InfoEvent, Message: "Info formatted message"},
		{Type: WarnEvent, Message: "Warn message"},
		{Type: WarnEvent, Message: "Warn formatted message"},
		{Type: ErrorEvent, Message: "Error message"},
		{Type: ErrorEvent, Message: "Error formatted message"},
		{Type: FatalEvent, Message: "Fatal message"},
		{Type: ThumbEvent, Message: "Function testThumstone called by github.com" +
			"/Thomasdezeeuw/logger.TestLog, from file " + file + " on line 79"},
		event,
	}

	if len(ew.events) != len(expected) {
		t.Fatalf("Expected to have %d events, but got %d",
			len(expected), len(ew.events))
	}

	for i, event := range ew.events {
		expectedEvent := expected[i]
		expectedEvent.Timestamp = now()
		expectedEvent.Tags = tags

		if expectedEvent.Type == FatalEvent {
			// sortof test the stack trace, best we can do.
			stackTrace := event.Data.([]byte)
			if !bytes.HasPrefix(stackTrace, []byte("goroutine")) {
				t.Errorf("Expected a stack trace as data for a Fatal event, but got %s ",
					string(stackTrace))
			} else if bytes.Contains(stackTrace, []byte("logger.getStackTrace")) ||
				bytes.Contains(stackTrace, []byte("logger.Fatal")) {
				t.Errorf("Expected the stack trace to not contain the logger.Fatal and "+
					"logger.getStackTrace, but got %s ", string(stackTrace))
			}

			event.Data = nil
		}

		if !reflect.DeepEqual(expectedEvent, event) {
			diff := pretty.Compare(event, expectedEvent)
			t.Errorf("Unexpected difference in event #%d: %s", i, diff)
		}
	}
}

func getPanicRecoveredValue(msg string) (recv interface{}) {
	defer func() {
		recv = recover()
	}()
	panic(msg)
}

func testThumstone(tags Tags) {
	Thumbstone(tags, "testThumstone")
}

func TestStartTwice(t *testing.T) {
	defer reset()
	var ew eventWriter
	Start(&ew)
	if err := Close(); err != nil {
		t.Fatal("Unexpected error closing initial log: " + err.Error())
	}

	defer expectPanic(t, "logger: can only Start once")
	Start(&ew)
}

func TestStartNoEventWriter(t *testing.T) {
	defer reset()
	defer expectPanic(t, "logger: need atleast a single EventWriter to write to")
	Start()
}

func expectPanic(t *testing.T, expected string) {
	recv := recover()
	if recv == nil {
		t.Fatal(`Expected a panic, but didn't get one`)
	}

	got := recv.(string)
	if got != expected {
		t.Fatalf("Expected panic value to be %s, but got %s", expected, got)
	}
}

// EventWriter that always returns a write error with the event message in it.
type errorEventWriter struct {
	closeError error
	errors     []error
}

func (eew *errorEventWriter) Write(event Event) error {
	return errors.New("Write error: " + event.Message)
}

func (eew *errorEventWriter) HandleError(err error) {
	eew.errors = append(eew.errors, err)
}

func (eew *errorEventWriter) Close() error {
	return eew.closeError
}

func TestErrorEventWriter(t *testing.T) {
	closeError := errors.New("Close error")

	defer reset()
	eew := errorEventWriter{closeError: closeError}
	Start(&eew)

	tags := Tags{"my", "tags"}
	Info(tags, "Info message1")
	Info(tags, "Won't be written to the writer")

	if err := Close(); err != closeError {
		t.Fatalf("Expceted the closing error to be %v, but got %v",
			closeError, err)
	}

	if expected, got := maxNWriteErrors+1, len(eew.errors); got != expected {
		t.Fatalf("Expected %d errors, but only got %d", expected, got)
	}

	// Expected errors:
	// 0 - 4: Write error: Info message1.
	// 5:     EventWriter is bad.
	expected := errors.New("Write error: Info message1")
	for i, got := range eew.errors {
		if i == 5 {
			expected = ErrBadEventWriter
		}

		if got.Error() != expected.Error() {
			t.Errorf("Expected error #%d to be %q, but got %q",
				i, expected.Error(), got.Error())
		}
	}
}

func reset() {
	eventChannel = make(chan Event, defaultEventChannelSize)
	eventChannelClosed = make(chan struct{}, 1)
	eventWriters = []EventWriter{}
	started = false
}

func TestGetStackTrace(t *testing.T) {
	t.Parallel()

	// Fake the Fatal call.
	var stackTrace []byte
	func() {
		stackTrace = getStackTrace()
	}()

	if !bytes.HasPrefix(stackTrace, []byte("goroutine")) {
		t.Errorf("Expected the stack trace to start with goroutine, but got %s ",
			string(stackTrace))
	} else if bytes.Contains(stackTrace, []byte("logger.getStackTrace")) ||
		bytes.Contains(stackTrace, []byte("logger.TestGetStackTrace.func1")) {
		t.Errorf("Expected the stack trace to not contain the "+
			"logger.TestGetStackTrace.func1 and logger.getStackTrace, but got it: %s",
			string(stackTrace))
	}
}
