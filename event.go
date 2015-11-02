// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

// todo: benchmark Event, EventType, Tags.String/Bytes with reference and not.

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// TimeFormat is used in Event.String() and Event.Bytes() to format the
// timestamp.
const TimeFormat = "2006-01-02 15:04:05"

// Event is created by a log operation. The timezone of timestamp is always the
// current timezone, recommend is to log time in the UTC timezone, by calling
// Event.Timestamp.UTC(), Event.String() does this by default.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Tags      Tags
	Message   string
	Data      interface{}
}

// String formats an event in the  following format:
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
		str += ", " + interfaceToString(event.Data)
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
		str += fmt.Sprintf(`, "data": %q`, interfaceToString(event.Data))
	}
	str += "}"
	return []byte(str), nil
}

// EventType indicent which level of detail a log operation has.
type EventType uint16

// Log levels available by default.
const (
	DebugEvent EventType = iota
	InfoEvent
	WarnEvent
	ErrorEvent
	FatalEvent
	ThumbEvent
)

var (
	eventTypeNames   = "DebugInfoWarnErrorFatalThumb"
	eventTypeIndices = []int{0, 5, 9, 13, 18, 23, 28}
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

func (eventType EventType) MarshalJSON() ([]byte, error) {
	qoutedEventType := strconv.Quote(eventType.String())
	return []byte(qoutedEventType), nil
}

// NewEventType creates a new fully supported custom EventType to be used in
// logging. This function makes sure that EventType.String() and
// EventType.Bytes() always return the correct name.
//
// Note: THIS FUNCTION IS NOT THREAD SAFE, use it before starting to log.
//
// Note: The maximum number of custom log levels is 65528, if more are created
// this function will panic.
func NewEventType(name string) EventType {
	if len(eventTypeIndices) >= math.MaxUint16 {
		panic("logger: can't have more then 65535 EventTypes")
	}

	eventTypeNames += name
	eventTypeIndices = append(eventTypeIndices, len(eventTypeNames))
	return EventType(len(eventTypeIndices) - 2)
}
