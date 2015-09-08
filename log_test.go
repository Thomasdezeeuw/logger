package logger

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// Time returned in calling now(), setup and test in init.
var t1 = time.Date(2015, 9, 1, 14, 22, 36, 0, time.UTC)

func init() {
	if !testing.Short() {
		os.Stderr.WriteString("Not running tests in short mode, this might take a while\n")
	}

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

// A data type to be used in calling Log.
type user struct {
	ID   int
	Name string
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
			Event{Type: DebugEvent, Tags: tags, Message: "Debug message"},
			Event{Type: DebugEvent, Tags: tags, Message: "Debug formatted message"},
			Event{Type: InfoEvent, Tags: tags, Message: "Info message"},
			Event{Type: InfoEvent, Tags: tags, Message: "Info formatted message"},
			Event{Type: WarnEvent, Tags: tags, Message: "Warn message"},
			Event{Type: WarnEvent, Tags: tags, Message: "Warn formatted message"},
			Event{Type: ErrorEvent, Tags: tags, Message: "Error message"},
			Event{Type: ErrorEvent, Tags: tags, Message: "Error formatted message"},
			Event{Type: FatalEvent, Tags: tags, Message: "Fatal message"},
			Event{Type: ThumbEvent, Tags: tags, Message: "Function testThumstone called by " + fn.Name() + ", from file " + file +
				" on line 87"},
			event,
		}

		if len(ew.events) != len(expected) {
			t.Fatalf("Expected to have %d events, but got %d",
				len(expected), len(ew.events))
		}

		for i, event := range ew.events {
			expectedEvent := expected[i]

			if expected, got := expectedEvent.Type, event.Type; expected != got {
				t.Errorf("Expected event #%d to have type %q, but got %q",
					i, expected, got)
			}

			if expected, got := t1, event.Timestamp; !expected.Equal(got) {
				t.Errorf("Expected event #%d to have timestamp %q, but got %q",
					i, expected, got)
			}

			if expected, got := expectedEvent.Tags, event.Tags; !reflect.DeepEqual(expected, got) {
				t.Errorf("Expected event #%d to have tags %q, but got %q",
					i, expected, got)
			}

			if expected, got := expectedEvent.Message, event.Message; expected != got {
				t.Errorf("Expected event #%d to have message %q, but got %q",
					i, expected, got)
			}

			// todo: test if we get a stacktrace with calling Fatal.
			if expectedEvent.Type == FatalEvent {
				event.Data = nil
			}

			if expected, got := expectedEvent.Data, event.Data; !reflect.DeepEqual(expected, got) {
				t.Errorf("Expected event #%d to have data %v, but got %v",
					i, expected, got)
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

	defer func() {
		recv := recover()
		if recv == nil {
			t.Fatal("Expected a second call to Start to panic, but it didn't")
		}

		const expected = "logger: can only Start once"
		got := interfaceToString(recv)
		if expected != got {
			t.Fatalf("Expected to panic with %q, but paniced with %q", expected, got)
		}
	}()
	Start(&ew)
}

func TestStartNoEventWriter(t *testing.T) {
	defer reset()
	defer func() {
		recv := recover()
		if recv == nil {
			t.Fatal("Expected a call to Start without any EventWriters to panic, but it didn't")
		}

		const expected = "logger: need atleast a single EventWriter to write to"
		got := interfaceToString(recv)
		if expected != got {
			t.Fatalf("Expected to panic with %q, but paniced with %q", expected, got)
		}
	}()
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

	// 10 = 5 bad write errors + 4 rewrite error of the first event + 1 bad
	// EventWriter error.
	if expected, got := 10, len(eew.errors); got != expected {
		t.Fatalf("Expected %d errors, but only got %d", expected, got)
	}

	// Expected errors:
	// 0 - write event 1
	// 1 - rewrite event 1
	// 2 - write event 2
	// 3 - rewrite event 1
	// 4 - write event 3
	// 5 - rewrite event 1
	// 6 - write event 4
	// 7 - rewrite event 1
	// 8 - write event 5
	// 9 - EventWriter is bad, more then 5 faulty writes, EventWriter will be dropped

	for i, got := range eew.errors {
		var expected error
		if i == 9 {
			expected = badEventWriterErr
		} else {
			d := 1
			switch i {
			case 2:
				d = 2
			case 4:
				d = 3
			case 6:
				d = 4
			case 8:
				d = 5
			}
			expected = fmt.Errorf("Write error: Info message%d", d)
		}

		if got.Error() != expected.Error() {
			t.Errorf("Expected error #%d to be %q, but got %q",
				i, expected.Error(), got.Error())
		}
	}
}

// todo: test with flaky writter, return nil, error, nil, error etc.
// todo: test with both good and bad Writer
// todo: test Thumbstone called from main, create a new sub proces and calling
// it and log to the stdout. Then parse the event and check it.

func testThumstone(tags Tags) {
	Thumbstone(tags, "testThumstone")
}

func reset() {
	eventChannel = make(chan Event, 1)
	eventChannelClosed = make(chan struct{}, 1)
	eventWriters = []EventWriter{}
	started = false
}
