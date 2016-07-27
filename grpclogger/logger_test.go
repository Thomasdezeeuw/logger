// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package grpclogger

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Thomasdezeeuw/logger"
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc/grpclog"
)

// EventWriter that collects the events and errors.
type eventWriter struct {
	events []logger.Event
	errors []error
	closed bool
}

func (ew *eventWriter) Write(event logger.Event) error {
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

func TestGrpcLogger(t *testing.T) {
	closedCalled := setupExitCounter()
	defer resetExitFns()
	closeFn := func() {
		*closedCalled++
	}

	var ew eventWriter
	logger.Start(&ew)

	tags := logger.Tags{"TestGrpcLogger"}
	logTime := time.Now()

	grpclog.SetLogger(CreateLogger(tags, closeFn))
	expectedEvents := callGrpcLogger(tags)

	if err := logger.Close(); err != nil {
		t.Fatal("Unexpected error closing logger: " + err.Error())
	}

	if expectedN, got := len(expectedEvents), len(ew.events); expectedN != got {
		t.Fatalf("Expected %d events, but got got %d", expectedN, got)
	}

	for i, event := range ew.events {
		expected, got := expectedEvents[i], event
		expected.Timestamp = logTime

		if err := compareEvents(i, expected, got); err != nil {
			t.Error(err)
		}
	}

	if *closedCalled != 6 {
		t.Fatalf("Expected the exit and close function to be called three times, but got %d",
			*closedCalled/2)
	}
}

// Make calls to the grpclog package and returns the expected events.
func callGrpcLogger(tags logger.Tags) (expected []logger.Event) {
	grpclog.Print("Error message")
	grpclog.Printf("Error %s message", "formatted")
	grpclog.Println("Error message")
	grpclog.Fatal("Fatal message")
	grpclog.Fatalf("Fatal %s message", "formatted")
	grpclog.Fatalln("Fatal message")

	return []logger.Event{
		{Type: logger.ErrorEvent, Tags: tags, Message: "Error message"},
		{Type: logger.ErrorEvent, Tags: tags, Message: "Error formatted message"},
		{Type: logger.ErrorEvent, Tags: tags, Message: "Error message"},
		{Type: logger.FatalEvent, Tags: tags, Message: "Fatal message"},
		{Type: logger.FatalEvent, Tags: tags, Message: "Fatal formatted message"},
		{Type: logger.FatalEvent, Tags: tags, Message: "Fatal message"},
	}
}

func compareEvents(i int, expected, got logger.Event) error {
	const margin = time.Millisecond

	// Can't mock time in the grpclog package, so we'll make sure it falls within
	// the margin.
	if got.Timestamp.Sub(expected.Timestamp) > margin {
		diff := pretty.Compare(got.Timestamp.Format(time.RFC3339Nano),
			expected.Timestamp.Format(time.RFC3339Nano))
		return fmt.Errorf("Expected and actual event #%d timestamps don't match\n%s",
			i, diff)
	}

	// Now the timestamp is tested, make sure we don't fall over it later on.
	got.Timestamp = expected.Timestamp

	if expected.Type == logger.FatalEvent {
		// Sortof test the stack trace, best we can do.
		stackTrace := got.Data.([]byte)
		if !bytes.HasPrefix(stackTrace, []byte("goroutine")) {
			return fmt.Errorf("Expected a stack trace as data for a Fatal event, but got %s ",
				string(stackTrace))
		}
		got.Data = nil
	}

	if !reflect.DeepEqual(expected, got) {
		diff := pretty.Compare(got, expected)
		return fmt.Errorf("Expected and actual #%d event don't match\n%s", i, diff)
	}

	return nil
}

func TestExit(t *testing.T) {
	defer resetExitFns()

	var exitCode int
	var closedCalled bool
	osExit = func(n int) {
		exitCode = n
	}
	closeFn := func() {
		closedCalled = true
	}

	exit(closeFn)

	if !closedCalled {
		t.Fatal("Close function not called")
	} else if exitCode != 1 {
		t.Fatalf("Expected exit to be called with 1, but got %d", exitCode)
	}
}

var (
	oldExit   = exit
	oldOSExit = osExit
)

func setupExitCounter() (counter *int) {
	var cnt int
	exit = func(closeFn func()) {
		closeFn()
		cnt++
	}
	return &cnt
}

func resetExitFns() {
	exit = oldExit
	osExit = oldOSExit
}
