// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"strings"
	"time"
)

const defaultMsgSize = 100

// Msg is a message created by a log operation.
type Msg struct {
	Level, Msg string
	Tags       Tags
	Timestamp  time.Time
}

// String creates a string message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
func (msg *Msg) String() string {
	return string(msg.Bytes())
}

// Bytes formats a message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
func (msg *Msg) Bytes() []byte {
	buf := make([]byte, 0, defaultMsgSize)

	// Write the date and time.
	// Format: "YYYY-MM-DD HH:MM:SS ".
	year, month, day := msg.Timestamp.Date()
	hour, min, sec := msg.Timestamp.Clock()
	itoa(&buf, year, 4)
	buf = append(buf, '-')
	itoa(&buf, int(month), 2)
	buf = append(buf, '-')
	itoa(&buf, day, 2)
	buf = append(buf, ' ')
	itoa(&buf, hour, 2)
	buf = append(buf, ':')
	itoa(&buf, min, 2)
	buf = append(buf, ':')
	itoa(&buf, sec, 2)
	buf = append(buf, ' ')

	// Write the level.
	// Format: "[LEVEL] " (level is always 5 characters long).
	buf = append(buf, '[')
	buf = append(buf, msg.Level...)
	buf = append(buf, ']')
	buf = append(buf, ' ')

	// Write the tags.
	// Format: "tag1, tag2: ".
	buf = append(buf, msg.Tags.Bytes()...)
	buf = append(buf, ':')
	buf = append(buf, ' ')

	// The actual message.
	buf = append(buf, strings.TrimSpace(msg.Msg)...)
	buf = append(buf, '\n')

	return buf
}

// Cheap integer to fixed-width decimal ASCII. Modified version from the Golang
// logger package.
func itoa(buf *[]byte, i int, wid int) {
	var b [4]byte // only used for year, month and day so 4 is enough.
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
