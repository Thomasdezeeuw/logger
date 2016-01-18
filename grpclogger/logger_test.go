// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package grpclogger

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/Thomasdezeeuw/logger"
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
	now := time.Now()

	grpclog.SetLogger(CreateLogger(tags, closeFn))
	grpclog.Print("Error message")
	grpclog.Printf("Error %s message", "formatted")
	grpclog.Println("Error message")
	grpclog.Fatal("Fatal message")
	grpclog.Fatalf("Fatal %s message", "formatted")
	grpclog.Fatalln("Fatal message")

	if err := logger.Close(); err != nil {
		t.Fatal("Unexpected error closing: " + err.Error())
	}

	expected := []logger.Event{
		{Type: logger.ErrorEvent, Timestamp: now, Tags: tags, Message: "Error message"},
		{Type: logger.ErrorEvent, Timestamp: now, Tags: tags, Message: "Error formatted message"},
		{Type: logger.ErrorEvent, Timestamp: now, Tags: tags, Message: "Error message"},
		{Type: logger.FatalEvent, Timestamp: now, Tags: tags, Message: "Fatal message"},
		{Type: logger.FatalEvent, Timestamp: now, Tags: tags, Message: "Fatal formatted message"},
		{Type: logger.FatalEvent, Timestamp: now, Tags: tags, Message: "Fatal message"},
	}

	if expectedN, got := len(expected), len(ew.events); expectedN != got {
		t.Fatalf("Expected %d events, but got only got %d", expectedN, got)
	}

	const margin = 100 * time.Millisecond
	for i, event := range ew.events {
		expectedEvent := expected[i]

		// Can't mock time in the logger package, so we have a truncate it.
		if !event.Timestamp.Truncate(margin).Equal(expectedEvent.Timestamp.Truncate(margin)) {
			t.Errorf("Expected event #%d to be %v, but got %v", i, expectedEvent, event)
			continue
		}
		event.Timestamp = expectedEvent.Timestamp

		if expectedEvent.Type == logger.FatalEvent {
			// sortof test the stacktrace, best we can do.
			stacktrace := event.Data.([]byte)
			if !bytes.HasPrefix(stacktrace, []byte("goroutine")) {
				t.Errorf("Expected a stacktrace as data for a Fatal event, but got %s ",
					string(stacktrace))
			}
			event.Data = nil
		}

		if expected, got := expectedEvent, event; !reflect.DeepEqual(expected, got) {
			t.Errorf("Expected event #%d to be %v, but got %v", i, expected, got)
		}
	}

	if *closedCalled != 6 {
		t.Fatalf("Expected the exit and close function to be called three times, but got %d",
			*closedCalled/2)
	}
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
		t.Fatalf("Expceted exit to be called with 1, but got %d", exitCode)
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
