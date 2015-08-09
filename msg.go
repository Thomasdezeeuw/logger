// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import "time"

// TimeFormat is used in Msg.String() to format the timestamp.
const TimeFormat = "2006-01-02 15:04:05"

// Msg is a message created by a log operation. The timezone of timestamp is
// alway is current timezone, recommend is to log time in the UTC timezone, by
// calling Msg.Timestamp.UTC(), Msg.String does this by default.
type Msg struct {
	Level     LogLevel
	Msg       string
	Tags      Tags
	Timestamp time.Time
	Data      interface{}
}

// String creates a string message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message, data
//
// Note: if is data is nil it doesn't get added to the message, so the format
// wil be:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
//
// Note: time is set to the UTC timezone.
func (msg *Msg) String() string {
	m := msg.Timestamp.UTC().Format(TimeFormat)
	m += " [" + msg.Level.String() + "] "
	m += msg.Tags.String() + ": "
	m += msg.Msg
	if msg.Data != nil {
		m += ", " + interfaceToString(msg.Data)
	}
	return m
}

// Bytes does the same as Tags.String, but returns a byte slice.
func (msg *Msg) Bytes() []byte {
	return []byte(msg.String())
}
