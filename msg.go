// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import "time"

// Msg is a message created by a log operation. The timezone of timestamp is
// alway is current timezone, advanced is to log time in the UTC timezone, by
// calling Msg.Timestamp.UTC().
type Msg struct {
	Level     LogLevel
	Msg       string
	Tags      Tags
	Timestamp time.Time
}

// String creates a string message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
//
// Note: time is to the UTC timezone.
func (msg *Msg) String() string {
	m := msg.Timestamp.UTC().Format("2006-01-02 15:04:05")
	m += " [" + msg.Level.String() + "] "
	m += msg.Tags.String() + ": "
	m += msg.Msg
	return m
}

// Bytes does the same as Tags.String, but returns a byte slice.
func (msg *Msg) Bytes() []byte {
	return []byte(msg.String())
}
