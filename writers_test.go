// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestFileEventWriter(t *testing.T) {
	file := strconv.FormatInt(time.Now().UnixNano(), 10)
	path := filepath.Join(os.TempDir(), "logger_"+file+".log")

	ew, err := NewFileEventWriter(path)
	if err != nil {
		t.Fatal("Unexpected error creating new file event writer: " + err.Error())
	}
	defer os.Remove(path)

	event := Event{
		Type:      DebugEvent,
		Timestamp: now(),
		Tags:      Tags{"TestFileEventWriter"},
		Message:   "Log message",
	}

	if err := ew.Write(event); err != nil {
		t.Fatal("Unexpected error writing to FileEventWriter: " + err.Error())
	}

	ew.HandleError(errors.New("writing error"))

	if err := ew.Close(); err != nil {
		t.Fatal("Unexpected error closing: " + err.Error())
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal("Unexpected error reading file: " + err.Error())
	}

	expected := "2015-09-01 14:22:36 [Debug] TestFileEventWriter: Log message\n" +
		"2015-09-01 14:22:36 [Error] FileEventWriter: Error writing to file: writing error\n"

	if got := string(bytes); got != expected {
		t.Fatalf("Expected file to contain:\n%s\nBut got:\n%s", expected, got)
	}
}

func TestConsoleEventWriter(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	ew := NewConsoleEventWriter()

	cew := ew.(*consoleEventWriter)
	cew.w = &buf
	cew.errW = &errBuf

	event := Event{
		Type:      DebugEvent,
		Timestamp: now(),
		Tags:      Tags{"TestConsoleEventWriter"},
		Message:   "Log message",
	}

	if err := ew.Write(event); err != nil {
		t.Fatal("Unexpected error writing to ConsoleEventWriter: " + err.Error())
	}

	ew.HandleError(errors.New("writing error"))

	if err := ew.Close(); err != nil {
		t.Fatal("Unexpected error closing: " + err.Error())
	}

	bytes, err := ioutil.ReadAll(&buf)
	if err != nil {
		t.Fatal("Unexpected error reading output buffer: " + err.Error())
	}

	expected := "2015-09-01 14:22:36 [Debug] TestConsoleEventWriter: Log message\n"

	if got := string(bytes); got != expected {
		t.Fatalf("Expected buffer to contain:\n%s\nBut got:\n%s", expected, got)
	}

	errBytes, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		t.Fatal("Unexpected error reading error buffer: " + err.Error())
	}

	expectedErr := "2015-09-01 14:22:36 [Error] ConsoleEventWriter: Error writing to console: writing error\n"

	if got := string(errBytes); got != expectedErr {
		t.Fatalf("Expected buffer to contain:\n%s\nBut got:\n%s", expectedErr, got)
	}
}

func TestJSONEventWriter(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	errorHandler := func(err error) {
		errBuf.WriteString(err.Error())
	}
	ew := NewJSONEventWriter(&buf, errorHandler)

	event := Event{
		Type:      DebugEvent,
		Timestamp: now(),
		Tags:      Tags{"TestJSONEventWriter"},
		Message:   "Log message",
	}

	if err := ew.Write(event); err != nil {
		t.Fatal("Unexpected error writing to JSONEventWriter: " + err.Error())
	}

	ew.HandleError(errors.New("some error"))

	if err := ew.Close(); err != nil {
		t.Fatal("Unexpected error closing: " + err.Error())
	}

	bytes, err := ioutil.ReadAll(&buf)
	if err != nil {
		t.Fatal("Unexpected error reading output buffer: " + err.Error())
	}

	expected := `{"type":"Debug","timestamp":"2015-09-01T14:22:36Z","tags":` +
		`["TestJSONEventWriter"],"message":"Log message"}` + "\n"

	if got := string(bytes); got != expected {
		t.Fatalf("Expected buffer to contain:\n%s\nBut got:\n%s", expected, got)
	}
}
