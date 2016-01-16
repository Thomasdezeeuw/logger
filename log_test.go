// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// Time returned in calling now(), setup and test in init.
var t1 = time.Date(2015, 9, 1, 14, 22, 36, 0, time.UTC)

func init() {
	margin := time.Millisecond
	t, t2 := now(), time.Now()
	if !t.Truncate(margin).Equal(t2.Truncate(margin)) {
		panic(fmt.Sprintf("now() doesn't return time.Now()! Expected %s, got %s",
			t2, t))
	}

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
		Type:      eventType,
		Timestamp: now(),
		Tags:      tags,
		Message:   "My event",
		Data:      data,
	}

	Debug(tags, "Debug message")
	Debugf(tags, "Debug %s message", "formatted")
	Info(tags, "Info message")
	Infof(tags, "Info %s message", "formatted")
	Warn(tags, "Warn message")
	Warnf(tags, "Warn %s message", "formatted")
	Error(tags, errors.New("Error message"))
	Errorf(tags, "Error %s message", "formatted")

	defer func() {
		recv := recover()
		Fatal(tags, recv)
		testThumstone(tags)
		Log(event)

		if err := Close(); err != nil {
			t.Fatal("Unexpected error closing: " + err.Error())
		}

		if len(ew.errors) != 0 {
			t.Fatalf("Unexpected error(s): %v", ew.errors)
		}

		pc, file, _, _ := runtime.Caller(0)
		fn := runtime.FuncForPC(pc)

		expected := []Event{
			{Type: DebugEvent, Timestamp: now(), Tags: tags, Message: "Debug message"},
			{Type: DebugEvent, Timestamp: now(), Tags: tags, Message: "Debug formatted message"},
			{Type: InfoEvent, Timestamp: now(), Tags: tags, Message: "Info message"},
			{Type: InfoEvent, Timestamp: now(), Tags: tags, Message: "Info formatted message"},
			{Type: WarnEvent, Timestamp: now(), Tags: tags, Message: "Warn message"},
			{Type: WarnEvent, Timestamp: now(), Tags: tags, Message: "Warn formatted message"},
			{Type: ErrorEvent, Timestamp: now(), Tags: tags, Message: "Error message"},
			{Type: ErrorEvent, Timestamp: now(), Tags: tags, Message: "Error formatted message"},
			{Type: FatalEvent, Timestamp: now(), Tags: tags, Message: "Fatal message"},
			{Type: ThumbEvent, Timestamp: now(), Tags: tags, Message: "Function testThumstone called by " +
				fn.Name() + ", from file " + file + " on line 88"},
			event,
		}

		if len(ew.events) != len(expected) {
			t.Fatalf("Expected to have %d events, but got %d",
				len(expected), len(ew.events))
		}

		for i, event := range ew.events {
			expectedEvent := expected[i]

			if expectedEvent.Type == FatalEvent {
				// sortof test the stack trace, best we can do.
				stackTrace := event.Data.([]byte)
				if !bytes.HasPrefix(stackTrace, []byte("goroutine")) {
					t.Errorf("Expected a stack trace as data for a Fatal event, but got %s ",
						string(stackTrace))
				} else if bytes.Index(stackTrace, []byte("logger.getStackTrace")) != -1 ||
					bytes.Index(stackTrace, []byte("logger.Fatal")) != -1 {
					t.Errorf("Expected the stack trace to not contain the logger.Fatal and logger.getStackTrace, but got %s ",
						string(stackTrace))
				}

				event.Data = nil
			}

			if expected, got := expectedEvent, event; !reflect.DeepEqual(expected, got) {
				t.Errorf("Expected event #%d to be %v, but got %v", i, expected, got)
			}
		}
	}()
	panic("Fatal message")
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
	eew := errorEventWriter{
		closeError: closeError,
	}
	Start(&eew)

	tags := Tags{"my", "tags"}
	Info(tags, "Info message1")
	Info(tags, "Info message2")
	Info(tags, "Info message3")
	Info(tags, "Info message4")
	Info(tags, "Info message5")
	Info(tags, "Won't be written to the writer")

	if err := Close(); err == nil {
		t.Fatal("Expected a closing error, but didn't get one")
	} else if err != closeError {
		t.Fatalf("Expceted the closing error to be %q, but got %q",
			closeError.Error(), err.Error())
	}

	// 6 = 5 bad write errors + 1 bad EventWriter error.
	if expected, got := 6, len(eew.errors); got != expected {
		t.Fatalf("Expected %d errors, but only got %d", expected, got)
	}

	// Expected errors:
	// 0 - 4: write event 1.
	// 5:     EventWriter is bad.

	for i, got := range eew.errors {
		var expected error
		if i == 5 {
			expected = ErrBadEventWriter
		} else {
			d := 1
			expected = fmt.Errorf("Write error: Info message%d", d)
		}

		if got.Error() != expected.Error() {
			t.Errorf("Expected error #%d to be %q, but got %q",
				i, expected.Error(), got.Error())
		}
	}
}

func testThumstone(tags Tags) {
	Thumbstone(tags, "testThumstone")
}

func reset() {
	eventChannel = make(chan Event, defaultEventChannelSize)
	eventChannelClosed = make(chan struct{}, 1)
	eventWriters = []EventWriter{}
	started = false
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
	} else if bytes.Index(stackTrace, []byte("logger.getStackTrace")) != -1 ||
		bytes.Index(stackTrace, []byte("logger.TestGetStackTrace.func1")) != -1 {
		t.Errorf("Expected the stack trace to not contain the logger.TestGetStackTrace.func1 and logger.getStackTrace, but got %s ",
			string(stackTrace))
	}
}
