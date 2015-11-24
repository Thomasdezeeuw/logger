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

func TestFindEventType(t *testing.T) {
	customEvent1 := NewEventType("custom-event-1")
	customEvent2 := NewEventType("custom-event-2")

	tests := []struct {
		input    string
		expected EventType
		found    bool
	}{
		{"Debug", DebugEvent, true},
		{"Info", InfoEvent, true},
		{"Warn", WarnEvent, true},
		{"Error", ErrorEvent, true},
		{"Fatal", FatalEvent, true},
		{"Thumb", ThumbEvent, true},
		{"Log", LogEvent, true},

		{"custom-event-1", customEvent1, true},
		{"custom-event-2", customEvent2, true},
		{"not-found", 0, false},
	}

	for _, test := range tests {
		got, ok := findEventType(test.input)

		if ok != test.found || (ok && got != test.expected) {
			t.Fatalf("Expected findEventType(%q) to return %v and %t, but got %v and %t",
				test.input, test.expected, test.found, got, ok)
		}
	}
}

type eventTypeTest struct {
	EventType EventType
	Text      string
	JSON      string
}

func getEventTypesTests() []eventTypeTest {
	return []eventTypeTest{
		{DebugEvent, "Debug", `"Debug"`},
		{ThumbEvent, "Thumb", `"Thumb"`},
		{InfoEvent, "Info", `"Info"`},
		{WarnEvent, "Warn", `"Warn"`},
		{ErrorEvent, "Error", `"Error"`},
		{FatalEvent, "Fatal", `"Fatal"`},
		{LogEvent, "Log", `"Log"`},
		{EventType(255), "EventType(255)", `"EventType(255)"`},
		{NewEventType("my-event-type"), "my-event-type", `"my-event-type"`},
		{NewEventType("my-\"event\"-type"), "my-\"event\"-type", `"my-\"event\"-type"`},
	}
}

func TestEventTypeStringAndBytes(t *testing.T) {
	defer resetEventTypes()

	for _, test := range getEventTypesTests() {
		got, gotBytes := test.EventType.String(), string(test.EventType.Bytes())
		if gotBytes != test.Text {
			t.Errorf("Expected EventType(%v).Bytes() to return %q, but got %q",
				test.EventType, test.Text, gotBytes)
		} else if got != test.Text {
			t.Errorf("Expected EventType(%v).String() to return %q, but got %q",
				test.EventType, test.Text, got)
		}
	}
}

func TestEventTypeMarshalling(t *testing.T) {
	defer resetEventTypes()

	for _, test := range getEventTypesTests() {
		if json, err := test.EventType.MarshalJSON(); err != nil {
			t.Errorf("Unexpected error marshaling %v into json: %s", test.EventType, err.Error())
		} else if got := string(json); got != test.JSON {
			t.Errorf("Expected EventType(%v).MarshalJSON() to return %q, but got %q",
				test.EventType, test.JSON, got)
		}
	}
}

func TestEventTypeUnmarshalling(t *testing.T) {
	defer resetEventTypes()

	for _, test := range getEventTypesTests() {
		var e = EventType(0)
		var gotEventType = &e

		var expectedError error = nil
		if _, ok := findEventType(test.EventType.String()); !ok {
			expectedError = ErrEventTypeUnknown
		}

		if err := gotEventType.UnmarshalText([]byte(test.Text)); err != expectedError {
			t.Fatalf("Expected EventType.UnmarshalText(%s) to return error %v, but got %v",
				test.Text, expectedError, err)
		} else if expectedError == nil && *gotEventType != test.EventType {
			t.Fatalf("Expected the event type to be %v, but got %v",
				test.EventType, gotEventType)
		}

		if err := gotEventType.UnmarshalJSON([]byte(test.JSON)); err != expectedError {
			t.Fatalf("Expected EventType.UnmarshalJSON(%s) to return error %v, but got %v",
				test.JSON, expectedError, err)
		} else if expectedError == nil && *gotEventType != test.EventType {
			t.Fatalf("Expected the event type to be %v, but got %v",
				test.EventType, gotEventType)
		}
	}
}

func TestNewEventTypeWithEmptyName(t *testing.T) {
	defer resetEventTypes()

	defer expectPanic(t, "logger: EventType name can't be empty")

	NewEventType("")
}

func TestNewEventTypeNotUnique(t *testing.T) {
	defer resetEventTypes()
	defer expectPanic(t, "logger: EventType must be unique")
	NewEventType("my-event-type")
	NewEventType("my-event-type")
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

func TestNewEventTypeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestNewEventTypeLimit in short mode")
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
