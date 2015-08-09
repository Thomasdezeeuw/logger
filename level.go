package logger

import (
	"errors"
	"fmt"
	"math"
)

// LogLevel indicent which level of detail a log operation has.
type LogLevel uint8

// Log levels available by default.
const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
	Thumb
)

var (
	logLevelNames   = "DebugInfoWarnErrorFatalThumb"
	logLevelIndices = []int{0, 5, 9, 13, 18, 23, 28}
)

// String return the name of the log level. Custom levels are also supported,
// if created with NewLogLevel.
func (lvl LogLevel) String() string {
	if int(lvl) >= len(logLevelIndices)-1 {
		return fmt.Sprintf("LogLevel(%d)", lvl)
	}

	startIndex := logLevelIndices[lvl]
	endIndex := logLevelIndices[lvl+1]
	return logLevelNames[startIndex:endIndex]
}

// Bytes does the same as LogLevel.String, but returns a byte slice.
func (lvl LogLevel) Bytes() []byte {
	return []byte(lvl.String())
}

// UnmarshalJSON provides a way to covert a string Log level to a LogLevel type.
//
// Note: custom log levels must be created first (with NewLogLevel), before
// they can be unmarshalled.
func (lvl *LogLevel) UnmarshalJSON(b []byte) error {
	l := len(logLevelIndices) - 1
	name := string(b)
	if len(name) >= 2 {
		// Drop the qoutes aroung the loglevels name.
		name = name[1 : len(name)-1]
	}

	for i, start := range logLevelIndices {
		if i == l {
			break
		}

		if end := logLevelIndices[i+1]; logLevelNames[start:end] == name {
			*lvl = LogLevel(i)
			return nil
		}
	}

	return errors.New("LogLevel not found")
}

// NewLogLevel creates a new fully supported custom log level for used in
// logging. This function makes sure that LogLevel.String and LogLevel.Bytes
// return the correct name.
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
