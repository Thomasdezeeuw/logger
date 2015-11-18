// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

type stringer struct{}

func (s *stringer) String() string {
	return "data"
}

func TestEvent(t *testing.T) {
	defer resetEventTypes()

	now := time.Now()
	tStr := now.UTC().Format(TimeFormat)
	tStrNano := now.UTC().Format(time.RFC3339Nano)

	var msgTests = []struct {
		event        Event
		expected     string
		expectedJSON string
	}{
		{Event{DebugEvent, now, Tags{"tag1", "tag2", "tag3"}, "Message6", 0},
			tStr + " [Debug] tag1, tag2, tag3: Message6, 0",
			`{"type": "Debug", "timestamp": "` + tStrNano + `", "tags": ["tag1", "tag2", "tag3"], ` +
				`"message": "Message6", "data": "0"}`},
		{Event{InfoEvent, now, Tags{"tag1", "tag2"}, "Message4", []byte("data")},
			tStr + " [Info] tag1, tag2: Message4, data",
			`{"type": "Info", "timestamp": "` + tStrNano + `", "tags": ["tag1", "tag2"], ` +
				`"message": "Message4", "data": "data"}`},
		{Event{WarnEvent, now, Tags{"tag1"}, "Message3", &stringer{}},
			tStr + " [Warn] tag1: Message3, data",
			`{"type": "Warn", "timestamp": "` + tStrNano + `", "tags": ["tag1"], ` +
				`"message": "Message3", "data": "data"}`},
		{Event{ErrorEvent, now, Tags{"tag1"}, "Message2", "data"},
			tStr + " [Error] tag1: Message2, data",
			`{"type": "Error", "timestamp": "` + tStrNano + `", "tags": ["tag1"], ` +
				`"message": "Message2", "data": "data"}`},
		{Event{FatalEvent, now, Tags{}, "Message1", nil},
			tStr + " [Fatal] : Message1",
			`{"type": "Fatal", "timestamp": "` + tStrNano + `", "tags": [], ` +
				`"message": "Message1"}`},
		{Event{ThumbEvent, now, Tags{"tag1", "tag2", "tag3"}, "Message5", errors.New("error data")},
			tStr + " [Thumb] tag1, tag2, tag3: Message5, error data",
			`{"type": "Thumb", "timestamp": "` + tStrNano + `", "tags": ["tag1", "tag2", "tag3"], ` +
				`"message": "Message5", "data": "error data"}`},
		{Event{NewEventType("My-event-type"), now, Tags{"tag1"}, "Message7", nil},
			tStr + " [My-event-type] tag1: Message7",
			`{"type": "My-event-type", "timestamp": "` + tStrNano + `", "tags": ["tag1"], ` +
				`"message": "Message7"}`},
		{Event{NewEventType(`my-"event"-type`), now, Tags{`tag"1"`}, "Message7", `"`},
			tStr + " [my-\"event\"-type] tag\"1\": Message7, \"",
			`{"type": "my-\"event\"-type", "timestamp": "` + tStrNano + `", "tags": ["tag\"1\""], ` +
				`"message": "Message7", "data": "\""}`},
	}

	for _, test := range msgTests {
		got, gotBytes := test.event.String(), string(test.event.Bytes())
		if gotBytes != test.expected {
			t.Errorf("Expected Event(%v).Bytes() to return %q, but got %q",
				test.event, test.expected, gotBytes)
		} else if got != test.expected {
			t.Errorf("Expected Event(%v).String() to return %q, but got %q",
				test.event, test.expected, got)
		}

		if json, err := test.event.MarshalJSON(); err != nil {
			t.Errorf("Unexpected error marshaling %v into json: %s", test.event, err.Error())
		} else if got := string(json); got != test.expectedJSON {
			t.Errorf("Expected Event(%v).MarshalJSON() to return %q, but got %q",
				test.event, test.expectedJSON, got)
		}
	}
}

func TestEventType(t *testing.T) {
	defer resetEventTypes()

	tests := []struct {
		eventType    EventType
		expected     string
		expectedJSON string
	}{
		{DebugEvent, "Debug", `"Debug"`},
		{ThumbEvent, "Thumb", `"Thumb"`},
		{InfoEvent, "Info", `"Info"`},
		{WarnEvent, "Warn", `"Warn"`},
		{ErrorEvent, "Error", `"Error"`},
		{FatalEvent, "Fatal", `"Fatal"`},
		{EventType(255), "EventType(255)", `"EventType(255)"`},
		{NewEventType("my-event-type"), "my-event-type", `"my-event-type"`},
		{NewEventType("my-\"event\"-type"), "my-\"event\"-type", `"my-\"event\"-type"`},
	}

	for _, test := range tests {
		got, gotBytes := test.eventType.String(), string(test.eventType.Bytes())
		if gotBytes != test.expected {
			t.Errorf("Expected EventType(%v).Bytes() to return %q, but got %q",
				test.eventType, test.expected, gotBytes)
		} else if got != test.expected {
			t.Errorf("Expected EventType(%v).String() to return %q, but got %q",
				test.eventType, test.expected, got)
		}

		if json, err := test.eventType.MarshalJSON(); err != nil {
			t.Errorf("Unexpected error marshaling %v into json: %s", test.eventType, err.Error())
		} else if got := string(json); got != test.expectedJSON {
			t.Errorf("Expected EventType(%v).MarshalJSON() to return %q, but got %q",
				test.eventType, test.expectedJSON, got)
		}
	}
}

var (
	// Minus builtin event types.
	maxCostumEventTypes = math.MaxUint16 - len(eventTypeIndices)

	oldEventTypeNames   = eventTypeNames
	oldEventTypeIndices = eventTypeIndices
)

func resetEventTypes() {
	eventTypeNames = oldEventTypeNames
	eventTypeIndices = oldEventTypeIndices
}

func TestNewLogLevelLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestNewLogLevelLimit in short mode")
	}
	defer resetEventTypes()

	for i := 1; i <= maxCostumEventTypes; i++ {
		expected := fmt.Sprintf("EventType-%d", i)
		eventType := NewEventType(expected)
		if got := eventType.String(); got != expected {
			t.Fatalf("Expected NewEventType(%q).String() to return %q, but got %q",
				expected, expected, got)
		}
	}

	defer func() {
		recv := recover()
		if recv == nil {
			t.Fatal("Expected a panic after creating 65528 log levels, but didn't get one")
		}

		got, ok := recv.(string)
		if !ok {
			t.Fatalf("Expected the recoverd panic to be a string, but it's %v", recv)
		}

		expected := "logger: can't have more then 65535 EventTypes"
		if got != expected {
			t.Fatalf("Expected the recoverd panic to be %s, but got %s", expected, got)
		}
	}()

	NewEventType(fmt.Sprintf("EventType-%d", maxCostumEventTypes+1))
}
