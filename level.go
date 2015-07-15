package logger

import (
	"fmt"
	"math"
)

// LogLevel indicent which level of detail a log operation has.
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

// String return the name of the log level. Examples:
//
//	Debug.String() // "Debug"
//	Info.String()  // "Info"
//	Fatal.String() // "Fatal"
//
// Custom levels are also supported, if created with NewLogLevel.
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
