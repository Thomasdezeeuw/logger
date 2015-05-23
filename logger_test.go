// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

// TODO: Test NewFile with error on creating/opening file.
package logger

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := newLogger(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating logger: " + err.Error())
	}

	if log.Name != name {
		t.Errorf("Expected newLogger(%q, %d, %v) to have name %q, but got %q",
			name, buf, name, log.Name)
	}

	storedLogger, ok := loggers[name]
	if !ok {
		t.Errorf("Expected newLogger(%q, %d, %v) to store the logger in the "+
			"loggers map, but it didn't", name, buf)
	}

	if log != storedLogger {
		t.Errorf("Expected newLogger(%q, %d, %v) to store the logger and return "+
			"the same logger, but it didn't", name, buf)
	}
}

func TestNewLoggerExisting(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	_, err := newLogger(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating logger: " + err.Error())
	}

	_, err = newLogger(name, &buf)
	if err == nil {
		t.Fatal("Expected an error creating a logger with the same name a " +
			"second time, but didn't get one")
	} else {
		errMsg := err.Error()
		expectedMsg := "logger: name " + name + " already taken"
		if errMsg != expectedMsg {
			t.Fatalf("Expected the error message to be %q, got %q, creating a "+
				"second logger with the same name", errMsg, expectedMsg)
		}
	}
}

func TestLogger(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := New(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log.ShowDebug = false

	tags := Tags{"Test"}
	msg := "Msg"
	err = errors.New("Msg")

	log.Thumbstone(msg)
	log.Debug(tags, msg) // Shouldn't show
	log.Info(tags, msg)
	log.Info(tags, "%s", msg)
	log.Error(tags, err)
	log.Fatal(tags, err)
	now := time.Now()

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing a logger: " + err.Error())
	}

	expectedSlice := []Msg{
		Msg{"THUMB", msg, Tags{"thumbstone"}, now},
		Msg{"INFO ", msg, tags, now},
		Msg{"INFO ", msg, tags, now},
		Msg{"ERROR", msg, tags, now},
		Msg{"FATAL", msg, tags, now},
	}

	scanner := bufio.NewScanner(&buf)
	i := 0
	for scanner.Scan() {
		if i >= len(expectedSlice) {
			break
		}
		got := scanner.Text()
		expected := expectedSlice[i].String()
		expected = expected[:len(expected)-1] // Drop the newline
		i++
		if got != expected {
			t.Fatalf("Expected logger to write %d. %q, but got %q", i, expected, got)
		}
	}
}

func TestLogger2(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := New(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log.ShowDebug = true

	tags := Tags{"Test"}
	msg := "Msg"
	err = errors.New("Msg")

	defer func() {
		recv := recover()
		log.Debug(tags, msg)
		log.Debug(tags, "%s", msg)
		log.Fatal(tags, recv)
		now := time.Now()

		err = log.Close()
		if err != nil {
			t.Fatal("Unexpected error, closing a logger: " + err.Error())
		}

		expectedSlice := []Msg{
			Msg{"DEBUG", msg, tags, now},
			Msg{"DEBUG", msg, tags, now},
			Msg{"FATAL", msg, tags, now},
		}

		scanner := bufio.NewScanner(&buf)
		i := 0
		for scanner.Scan() {
			if i >= len(expectedSlice) {
				break
			}
			got := scanner.Text()
			expected := expectedSlice[i].String()
			expected = expected[:len(expected)-1] // Drop the newline
			i++
			if got != expected {
				t.Fatalf("Expected logger to write %d. %q, but got %q",
					i, expected, got)
			}
		}
	}()

	panic(msg)
}

func TestLoggerFatal(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := New(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log.ShowDebug = false

	tags := Tags{"Test"}
	msg := "Msg"

	log.Fatal(tags, msg)
	now := time.Now()

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing a logger: " + err.Error())
	}

	expectedSlice := []Msg{
		Msg{"FATAL", msg, tags, now},
	}

	scanner := bufio.NewScanner(&buf)
	i := 0
	for scanner.Scan() {
		if i >= len(expectedSlice) {
			break
		}
		got := scanner.Text()
		expected := expectedSlice[i].String()
		expected = expected[:len(expected)-1] // Drop the newline
		i++
		if got != expected {
			t.Fatalf("Expected logger to write %d. %q, but got %q", i, expected, got)
		}
	}
}

func TestLoggerFatal2(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := New(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log.ShowDebug = false

	tags := Tags{"Test"}
	msg := "1"

	log.Fatal(tags, 1)
	now := time.Now()

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing a logger: " + err.Error())
	}

	expectedSlice := []Msg{
		Msg{"FATAL", msg, tags, now},
	}

	scanner := bufio.NewScanner(&buf)
	i := 0
	for scanner.Scan() {
		if i >= len(expectedSlice) {
			break
		}
		got := scanner.Text()
		expected := expectedSlice[i].String()
		expected = expected[:len(expected)-1] // Drop the newline
		i++
		if got != expected {
			t.Fatalf("Expected logger to write %d. %q, but got %q", i, expected, got)
		}
	}
}

func TestGet(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := newLogger(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	storedLogger, err := Get(name)
	if err != nil {
		t.Fatalf("Unexpected error, getting a logger: " + err.Error())
	}

	if log != storedLogger {
		t.Errorf("Expected Get(%q) to return the same logger as newLogger(), but "+
			"it didn't", name)
	}
}

func TestGetNoneExisting(t *testing.T) {
	defer reset()
	name := "Test"
	_, err := Get(name)
	if err == nil {
		t.Fatal("Expected an error when getting a unkown logger, but didn't get " +
			"one")
	} else {
		errMsg := err.Error()
		expectedMsg := "logger: no logger found with name " + name
		if errMsg != expectedMsg {
			t.Fatalf("Expected the error message to be %q, got %q, getting a "+
				"unkown logger", errMsg, expectedMsg)
		}
	}
}

type msgWriter struct {
	buf string
}

func (mw *msgWriter) Write(msg Msg) error {
	mw.buf += msg.String()
	return nil
}

func (mw *msgWriter) String() string {
	return mw.buf
}

func (mw *msgWriter) Close() error {
	return nil
}

func (mw *msgWriter) Flush() error {
	return nil
}

func TestNewMsgWriter(t *testing.T) {
	defer reset()
	buf := msgWriter{}
	name := "Test"
	log, err := NewMsgWriter(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	tags := Tags{"Test"}
	msg := "Msg"

	t1 := time.Now()
	log.Info(tags, msg)
	time.Sleep(100 * time.Millisecond)
	t2 := time.Now()
	log.Info(tags, msg)

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing a logger: " + err.Error())
	}

	m1, m2 := Msg{"INFO ", msg, tags, t1}, Msg{"INFO ", msg, tags, t2}
	expected := m1.String() + m2.String()
	got := buf.String()
	if got != expected {
		t.Fatalf("Expected logger to write %q, but got %q", expected, got)
	}
}

func TestNewMsgWriterExisting(t *testing.T) {
	defer reset()
	buf := msgWriter{}
	name := "Test"
	log, err := NewMsgWriter(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, when create a new logger: " + err.Error())
	}

	_, err = NewMsgWriter(name, &buf)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name a " +
			"second time, but didn't get one")
	} else {
		errMsg := err.Error()
		expectedMsg := "logger: name " + name + " already taken"
		if errMsg != expectedMsg {
			t.Fatalf("Expected the error message to be %q, got %q, when creating a "+
				"second logger with the same name", errMsg, expectedMsg)
		}
	}
	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, when closing logger: " + err.Error())
	}
}

func TestNew(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := New(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	tags := Tags{"Test"}
	msg := "Msg"

	t1 := time.Now()
	log.Info(tags, msg)
	time.Sleep(100 * time.Millisecond)
	t2 := time.Now()
	log.Info(tags, msg)

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing a logger: " + err.Error())
	}

	m1, m2 := Msg{"INFO ", msg, tags, t1}, Msg{"INFO ", msg, tags, t2}
	expected := m1.String() + m2.String()
	got := buf.String()
	if got != expected {
		t.Fatalf("Expected logger to write %q, but got %q", expected, got)
	}
}

func TestNewExisting(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	log, err := New(name, &buf)
	if err != nil {
		t.Fatal("Unexpected error, when create a new logger: " + err.Error())
	}

	_, err = New(name, &buf)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name a " +
			"second time, but didn't get one")
	} else {
		errMsg := err.Error()
		expectedMsg := "logger: name " + name + " already taken"
		if errMsg != expectedMsg {
			t.Fatalf("Expected the error message to be %q, got %q, when creating a "+
				"second logger with the same name", errMsg, expectedMsg)
		}
	}
	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, when closing logger: " + err.Error())
	}
}

func TestNewFile(t *testing.T) {
	path := "./tmp.log"
	defer reset()
	defer func() {
		if err := os.Remove(path); err != nil {
			t.Fatal("Unexpected error, remove tmp.log file: " + err.Error())
		}
	}()
	name := "Test"
	log, err := NewFile(name, path)
	if err != nil {
		t.Fatal("Unexpected error, creating a file logger: " + err.Error())
	}

	tags := Tags{"Test"}
	msg := "Msg"

	t1 := time.Now()
	log.Info(tags, msg)
	time.Sleep(100 * time.Millisecond)
	t2 := time.Now()
	log.Info(tags, msg)

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing file: " + err.Error())
	}

	gotBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal("Unexpected error, reading log file: " + err.Error())
	}
	got := string(gotBytes)

	m1, m2 := Msg{"INFO ", msg, tags, t1}, Msg{"INFO ", msg, tags, t2}
	expected := m1.String() + m2.String()
	if got != expected {
		t.Fatalf("Expected logger to write %q, but got %q", expected, got)
	}
}

func TestCombine(t *testing.T) {
	defer reset()
	var buf1, buf2 bytes.Buffer
	name := "Test"
	log1, err := New(name+"1", &buf1)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log2, err := New(name+"2", &buf2)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	log, err := Combine(name, log1, log2)
	if err != nil {
		t.Fatal("Unexpected error, combining two loggers: " + err.Error())
	}

	tags := Tags{"Test"}
	msg := "Msg"
	err = errors.New("Msg")

	log.Info(tags, msg)
	log.Info(tags, msg)

	err = log.Close()
	if err != nil {
		t.Fatal("Unexpected error, closing a logger: " + err.Error())
	}

	now := time.Now()
	expectedSlice := []Msg{
		Msg{"INFO ", msg, tags, now},
		Msg{"INFO ", msg, tags, now},
	}

	scanner1 := bufio.NewScanner(&buf1)
	scanner2 := bufio.NewScanner(&buf2)
	for _, scanner := range []*bufio.Scanner{scanner1, scanner2} {
		i := 0
		for scanner.Scan() {
			if i >= len(expectedSlice) {
				t.Fatal("Output longer then expected")
			}
			got := scanner.Text()
			expected := expectedSlice[i].String()
			expected = expected[:len(expected)-1] // Drop the newline
			i++
			if got != expected {
				t.Fatalf("Expected logger to write %d. %q, but got %q",
					i, expected, got)
			}
		}
	}
}

func TestCombineExistingName(t *testing.T) {
	defer reset()
	var buf1, buf2 bytes.Buffer
	name := "Test"
	log1, err := New(name, &buf1)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log2, err := New(name+"2", &buf2)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	_, err = Combine(name, log1, log2)
	if err == nil {
		t.Fatal("Expected error, but didn't get one")
	}

	errMsg := err.Error()
	expectedMsg := "logger: name " + name + " already taken"
	if errMsg != expectedMsg {
		t.Fatalf("Expected the error message to be %q, got %q, when combining "+
			"logger with the same name", errMsg, expectedMsg)
	}
}

func TestCombineNone(t *testing.T) {
	defer reset()
	name := "Test"
	_, err := Combine(name)
	if err == nil {
		t.Fatal("Expected error, but didn't get one")
	}

	errMsg := err.Error()
	expectedMsg := "logger: Combine requires atleast one logger"
	if errMsg != expectedMsg {
		t.Fatalf("Expected the error message to be %q, got %q, when combining "+
			"zero loggers", errMsg, expectedMsg)
	}
}

// Reset resets all global variable to the original.
func reset() {
	loggers = map[string]*Logger{}
}
