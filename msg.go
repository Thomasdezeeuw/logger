// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Msg is a message created by a log operation.
type Msg struct {
	Level     LogLevel
	Msg       string
	Tags      Tags
	Timestamp time.Time
}

// String creates a string message in the following format:
//	YYYY-MM-DD HH:MM:SS [LEVEL] tag1, tag2: message
func (msg *Msg) String() string {
	return string(msg.Bytes())
}

// Bytes does the same as Tags.String, but returns a byte slice.
func (msg *Msg) Bytes() []byte {
	var buf []byte

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

	// Write the log level.
	// Format: "[LEVEL] ".
	buf = append(buf, '[')
	buf = append(buf, msg.Level.Bytes()...)
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

type LogLevel uint8

// Log levels available by default.
const (
	Debug LogLevel = iota
	Thumb
	Info
	Warn
	Error
	Fatal
)

var (
	logLevelNames   = "DebugThumbInfoWarnErrorFatal"
	logLevelIndices = []int{0, 5, 10, 14, 18, 23, 28}
)

func (lvl LogLevel) String() string {
	if int(lvl) >= len(logLevelIndices)-1 {
		return fmt.Sprintf("LogLevel(%d)", lvl)
	}

	startIndex := logLevelIndices[lvl]
	endIndex := logLevelIndices[lvl+1]
	return logLevelNames[startIndex:endIndex]
}

func (lvl LogLevel) Bytes() []byte {
	return []byte(lvl.String())
}

// NewLogLevel creates a new fully supported custom log level for used in
//  logging, this function makes sure that LogLevel.String and LogLevel.Bytes
// return the correct given name.
//
// Note: THIS FUNCTION IS NOT THREAD SAFE, use it before starting to log.
//
// Note: The maximum number of custom log levels is 248, if more are created
// this function will panic.
func NewLogLevel(name string) LogLevel {
	if len(logLevelIndices) >= math.MaxUint8 {
		panic("ini: can't have more then 255 log levels")
	}

	logLevelNames += name
	logLevelIndices = append(logLevelIndices, len(logLevelNames))
	return LogLevel(len(logLevelIndices) - 2)
}
