// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Thomasdezeeuw/logger/internal/util"
)

// TimeFormat is used in Event.String() and Event.Bytes() to format the
// timestamp.
const TimeFormat = "2006-01-02 15:04:05"

// Event is created by a log operation. The timezone of the timestamp is always
// the current timezone, recommend is to log time in the UTC timezone, by
// calling Event.Timestamp.UTC(), Event.String and Event.Bytes does this by
// default.
//
// Advised is to store an EventType using it's string format (using
// EventType.String or .Bytes), not it's numeral format. Because the numeral
// value of an EventType might change, this happens when a new builtin EventType
// gets added, or if the order of calls to NewEventType changes.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Tags      Tags
	Message   string
	Data      interface{}
}

// String formats an event in the following format:
//	YYYY-MM-DD HH:MM:SS [TYPE] tag1, tag2: message, data
//
// Note: the timestamp is set to the UTC timezone.
//
// Note: if is data is nil it doesn't get added to the message, so the format
// wil be:
//	YYYY-MM-DD HH:MM:SS [TYPE] tag1, tag2: message
func (event Event) String() string {
	str := event.Timestamp.UTC().Format(TimeFormat)
	str += " [" + event.Type.String() + "] "
	str += event.Tags.String() + ": "
	str += event.Message
	if event.Data != nil {
		str += ", " + util.InterfaceToString(event.Data)
	}
	return str
}

// Bytes does the same as Event.String(), but returns a byte slice.
func (event Event) Bytes() []byte {
	return []byte(event.String())
}

// MarshalJSON coverts the event to a JSON formatted byte slice. It uses
// time.RFC3339Nano to format the timestamp.
func (event Event) MarshalJSON() ([]byte, error) {
	tagsJSON, err := event.Tags.MarshalJSON()
	if err != nil {
		return []byte{}, err
	}

	str := fmt.Sprintf(`{"type": %q, "timestamp": %q, "tags": %s, "message": %q`,
		event.Type.String(), event.Timestamp.UTC().Format(time.RFC3339Nano),
		string(tagsJSON), event.Message)
	if event.Data != nil {
		str += fmt.Sprintf(`, "data": %q`, util.InterfaceToString(event.Data))
	}
	str += "}"
	return []byte(str), nil
}

// EventType indicates what type a log operation has.
type EventType uint16

// EventTypes available by default.
const (
	DebugEvent EventType = iota
	InfoEvent
	WarnEvent
	ErrorEvent
	FatalEvent
	ThumbEvent
	LogEvent // Used in relaying logs from the default log package.
)

// Names and indices for use in EventType.String and Event.Bytes, can be
// modified by NewEventType
var (
	eventTypeNames   = "DebugInfoWarnErrorFatalThumbLog"
	eventTypeIndices = []int{0, 5, 9, 13, 18, 23, 28, 31}
)

// String returns the name of the event type. Custom event types are also
// supported, if created with NewEventType.
func (eventType EventType) String() string {
	if !isDefinedEventType(eventType) {
		return fmt.Sprintf("EventType(%d)", eventType)
	}

	startIndex := eventTypeIndices[eventType]
	endIndex := eventTypeIndices[eventType+1]
	return eventTypeNames[startIndex:endIndex]
}

func isDefinedEventType(eventType EventType) bool {
	return int(eventType) <= len(eventTypeIndices)-1
}

// Bytes does the same as EventType.String(), but returns a byte slice.
func (eventType EventType) Bytes() []byte {
	return []byte(eventType.String())
}

// MarshalJSON returns a qouted string event type.
func (eventType EventType) MarshalJSON() ([]byte, error) {
	qoutedEventType := strconv.Quote(eventType.String())
	return []byte(qoutedEventType), nil
}

// ErrEventTypeUnknown gets returned by EventType.UnmarshalJSON and
// EventType.UnmarshalText if the EventType is not known.
var ErrEventTypeUnknown = errors.New("unkown EventType")

// UnmarshalJSON converts a qouted string EventType (e.g. "Error") to an actual
// typed EventTyped.
//
// Note: custom EventTypes are supported, but must created using NewEventType.
func (eventType *EventType) UnmarshalJSON(rawType []byte) error {
	if len(rawType) <= 2 {
		return ErrEventTypeUnknown
	}

	// Drop the qoutes.
	rawText, err := strconv.Unquote(string(rawType))
	if err != nil {
		return ErrEventTypeUnknown
	}

	return eventType.UnmarshalText([]byte(rawText))
}

// UnmarshalText converts a string EventType (e.g. Error) to an actual typed
// EventTyped.
//
// Note: custom EventTypes are supported, but must created using NewEventType.
func (eventType *EventType) UnmarshalText(rawType []byte) error {
	if len(rawType) == 0 {
		return ErrEventTypeUnknown
	}

	t, ok := findEventType(string(rawType))
	if !ok {
		return ErrEventTypeUnknown
	}

	*eventType = t
	return nil
}

// NewEventType creates a new fully supported custom EventType to be used in
// logging. This function makes sure that all EventType functions (e.g
// EventType.String) work correctly. The name can't be empty and must be unique.
//
// Advised is to store an EventType using it's string format (using
// EventType.String or .Bytes), not it's numeral format. Because the numeral
// value of an EventType might change, this happens when a new builtin EventType
// gets added, or if the order of calls to NewEventType changes.
//
// Note: THIS FUNCTION IS NOT SAFE FOR CONCURRENT USE, use it before starting to
// log.
//
// Note: The maximum number of custom log levels is 65528, if more are created
// this function will panic.
func NewEventType(name string) EventType {
	if len(eventTypeIndices) >= math.MaxUint16 {
		panic("logger: can't have more then 65535 EventTypes")
	} else if len(name) == 0 {
		panic("logger: EventType name can't be empty")
	}

	if _, ok := findEventType(name); ok {
		panic("logger: EventType must be unique")
	}

	eventTypeNames += name
	eventTypeIndices = append(eventTypeIndices, len(eventTypeNames))
	return EventType(len(eventTypeIndices) - 2)
}

func findEventType(name string) (EventType, bool) {
	for i, l := 0, len(eventTypeIndices)-1; i < l; i++ {
		start := eventTypeIndices[i]
		end := eventTypeIndices[i+1]
		possibly := eventTypeNames[start:end]

		if possibly == name {
			return EventType(i), true
		}
	}

	return 0, false
}
