package logger

import (
	"errors"
	"log"
	"strings"
	"time"
)

const (
	logPrefix         = "logger:"
	logTimeLayout     = "2006/01/02 15:04:05.000000"
	logPrefixLength   = len(logPrefix)
	logMetadataLength = len(logPrefix) + len(logTimeLayout)
)

// ErrLogFormat indicates an incorrect format in a log line. This will be used
// as the message of an error Event with the original log line as Event.Data.
//
// BridgeLogPgk changes the log flags and prefix so it can parse the log line.
// If those flags and/or prefix are changed later it might cause this error to
// appear.
//
// EXPERIMENTAL, api might change, tied to BridgeLogPgk.
var ErrLogFormat = errors.New("logger: log format incorrect")

// BridgeLogPgk creates a bridge between the loggger package and the standard
// library's log package. Calls to log.Print* will be converted into an Event
// and will be written to to the event writes provided to the Start function.
// Events will have LogEvent as EventType, because the standard libary's log
// package doesn't have log levels.
//
// EXPERIMENTAL, api might change.
//
// Note: calls to log.Fatal* will not be written. It calls os.Exit right after
// it writes to the logger package. But os.Exit never calls any deffered
// functions and because because the logger package is asynchronous it will not
// write the last log.
func BridgeLogPgk(tags Tags) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix(logPrefix)

	w := logToEvent{tags, time.Now().Location()}
	log.SetOutput(&w)
}

// logToEvent takes bytes created by the standard library's log package and
// converts it to an Event and send it over the eventChannel.
type logToEvent struct {
	tags Tags
	loc  *time.Location
}

func (l *logToEvent) Write(b []byte) (int, error) {
	line := string(b)
	n := len(b)

	if !strings.HasPrefix(line, logPrefix) || len(line) < logMetadataLength {
		eventChannel <- createErrorLogEvent(ErrLogFormat, line, l.tags)
		return n, nil
	}

	timeStr := line[logPrefixLength:logMetadataLength]
	t, err := time.ParseInLocation(logTimeLayout, timeStr, l.loc)
	if err != nil {
		eventChannel <- createErrorLogEvent(err, line, l.tags)
		return n, nil
	}

	eventChannel <- Event{
		Type:      LogEvent,
		Timestamp: t,
		Tags:      l.tags,
		Message:   line[logMetadataLength+1 : n-1], // Drop metadata and newline.
	}

	return n, nil
}

func createErrorLogEvent(err error, line string, tags Tags) Event {
	return Event{
		Type:      ErrorEvent,
		Timestamp: now(),
		Tags:      tags,
		Message:   err.Error(),
		Data:      line,
	}
}
