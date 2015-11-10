// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

const (
	defaultFileFlag       = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	defaultFilePermission = 0600
)

type fileEventWriter struct {
	w       *bufio.Writer
	f       *os.File
	minType EventType
}

func (ew *fileEventWriter) Write(event Event) error {
	if event.Type < ew.minType {
		return nil
	}
	bytes := append(event.Bytes(), '\n')
	_, err := ew.w.Write(bytes)
	return err
}

func (ew *fileEventWriter) HandleError(err error) {
	msg := now().Format(TimeFormat) + " [Error] FileEventWriter: "
	msg += "Error writing to file: " + err.Error() + "\n"
	ew.w.WriteString(msg)
}

func (ew *fileEventWriter) Close() error {
	flushErr := ew.w.Flush()
	err := ew.f.Close()
	if err == nil {
		err = flushErr
	}
	return err
}

// NewFileEventWriter creates a EventWriter that writes to the given file.
// MinType is the minimal EventType an event must have to be logged. For example
// if minType is InfoEvent, then any events with an EventType of Debug will not
// be logged.
func NewFileEventWriter(path string, minType EventType) (EventWriter, error) {
	f, err := os.OpenFile(path, defaultFileFlag, defaultFilePermission)
	if err != nil {
		return nil, err
	}

	return &fileEventWriter{bufio.NewWriter(f), f, minType}, nil
}

type consoleEventWriter struct {
	w       io.Writer
	errW    io.Writer
	minType EventType
}

func (ew *consoleEventWriter) Write(event Event) error {
	if event.Type < ew.minType {
		return nil
	}
	bytes := append(event.Bytes(), '\n')
	_, err := ew.w.Write(bytes)
	return err
}

func (ew *consoleEventWriter) HandleError(err error) {
	msg := now().Format(TimeFormat) + " [Error] ConsoleEventWriter: "
	msg += "Error writing to console: " + err.Error() + "\n"
	ew.errW.Write([]byte(msg))
}

func (ew *consoleEventWriter) Close() error {
	return nil
}

// Stubbed for testing
var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// NewConsoleEventWriter creates a new EventWriter that writes to standard out
// and standard error. MinType is the minimal EventType an event must have to
// be logged. For example if minType is InfoEvent, then any events with an
// EventType of Debug will not be logged.
func NewConsoleEventWriter(minType EventType) EventWriter {
	return &consoleEventWriter{stdout, stderr, minType}
}

type jsonEventWriter struct {
	enc          *json.Encoder
	errorHandler func(error)
	minType      EventType
}

func (ew *jsonEventWriter) Write(event Event) error {
	if event.Type < ew.minType {
		return nil
	}
	return ew.enc.Encode(event)
}

func (ew *jsonEventWriter) HandleError(err error) {
	ew.errorHandler(err)
}

func (ew *jsonEventWriter) Close() error {
	return nil
}

// NewJSONEventWriter creates a new EventWriter that writes JSON to the given
// writer. MinType is the minimal EventType an event must have to be logged. For
// example if minType is InfoEvent, then any events with an EventType of Debug
// will not be logged.
func NewJSONEventWriter(w io.Writer, errorHandler func(error), minType EventType) EventWriter {
	return &jsonEventWriter{json.NewEncoder(w), errorHandler, minType}
}
