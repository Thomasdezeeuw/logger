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

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		width    int
		expected string
	}{
		{1999, 4, "1999"}, // In case of time travel.
		{2009, 4, "2009"},
		{2010, 4, "2010"},
		{2015, 4, "2015"},
		{2016, 4, "2016"},
		{2017, 4, "2017"},
		{2018, 4, "2018"},
		{2019, 4, "2019"},
		{2020, 4, "2020"},
		{5000, 4, "5000"}, // At this point, all bets are off.
		{1, 2, "01"},
		{5, 2, "05"},
		{8, 2, "08"},
		{10, 2, "10"},
		{11, 2, "11"},
		{12, 2, "12"},
		{21, 2, "21"},
		{25, 2, "25"},
		{31, 2, "31"},
	}

	for _, test := range tests {
		var buf []byte
		itoa(&buf, test.input, test.width)

		if got := string(buf); got != test.expected {
			t.Errorf("Expected itoa(%v, %d, %d) to return %q, but got %q",
				&buf, test.input, test.width, test.expected, got)
		}
	}
}

func TestTagsBytes(t *testing.T) {
	tests := []struct {
		tags     Tags
		expected string
	}{
		{Tags{}, ""},
		{Tags{"tag1"}, "tag1"},
		{Tags{"tag1", "tag2"}, "tag1, tag2"},
		{Tags{"tag1", "tag2", "tag3"}, "tag1, tag2, tag3"},
	}

	for _, test := range tests {
		if got := string(test.tags.Bytes()); got != test.expected {
			t.Errorf("Expected Tags{%v}.Bytes() to return %q, got %q",
				test.tags, test.expected, got)
		}
	}
}

func TestTagsString(t *testing.T) {
	tests := []struct {
		tags     Tags
		expected string
	}{
		{Tags{}, ""},
		{Tags{"tag1"}, "tag1"},
		{Tags{"tag1", "tag2"}, "tag1, tag2"},
		{Tags{"tag1", "tag2", "tag3"}, "tag1, tag2, tag3"},
	}

	for _, test := range tests {
		if got := test.tags.String(); got != test.expected {
			t.Errorf("Expected Tags{%v}.String() to return %q, got %q",
				test.tags, test.expected, got)
		}
	}
}

func TestFormatMsg(t *testing.T) {
	const layout = "2006/01/02 15:04:05"

	tests := []struct {
		time     string
		lvl      string
		tags     Tags
		msg      string
		expected string
	}{
		{"2015/03/01 04:48:33", "FATAL", Tags{"tag1"}, "test message",
			"2015-03-01 04:48:33 [FATAL] tag1: test message\n"},
		{"1999/01/31 00:01:12", "INFO ", Tags{"tag1", "tag2"}, "test message2",
			"1999-01-31 00:01:12 [INFO ] tag1, tag2: test message2\n"},
		{"2020/12/16 23:48:59", "DEBUG", Tags{}, "test message3",
			"2020-12-16 23:48:59 [DEBUG] : test message3\n"},
	}

	for _, test := range tests {
		tm, err := time.Parse(layout, test.time)
		if err != nil {
			t.Fatal("Ugh, really...? " + err.Error())
		}

		got := formatMsg(tm, test.lvl, test.tags, test.msg)
		if got != test.expected {
			t.Errorf("Expected formatMsg(%v, %q, %q, %q) to return %q, but got %q",
				tm, test.lvl, test.tags.String(), test.msg, test.expected, got)
		}
	}
}

func TestNewLogger(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	size := 1024
	log, err := newLogger(name, size, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating logger: " + err.Error())
	}

	if log.Name != name {
		t.Errorf("Expected newLogger(%q, %d, %v) to have name %q, but got %q",
			name, size, buf, name, log.Name)
	}

	storedLogger, ok := loggers[name]
	if !ok {
		t.Errorf("Expected newLogger(%q, %d, %v) to store the logger in the "+
			"loggers map, but it didn't", name, size, buf)
	}

	if log != storedLogger {
		t.Errorf("Expected newLogger(%q, %d, %v) to store the logger and return "+
			"the same logger, but it didn't", name, size, buf)
	}
}

func TestNewLoggerExisting(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	size := 1024
	_, err := newLogger(name, size, &buf)
	if err != nil {
		t.Fatal("Unexpected error, creating logger: " + err.Error())
	}

	_, err = newLogger(name, size, &buf)
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
	size := 1024
	log, err := New(name, size, &buf)
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

	expectedSlice := []string{
		formatMsg(now, "THUMB", Tags{"thumbstone"}, msg),
		formatMsg(now, "INFO ", tags, msg),
		formatMsg(now, "INFO ", tags, msg),
		formatMsg(now, "ERROR", tags, msg),
		formatMsg(now, "FATAL", tags, msg),
	}

	scanner := bufio.NewScanner(&buf)
	i := 0
	for scanner.Scan() {
		if i >= len(expectedSlice) {
			break
		}
		got := scanner.Text()
		expected := expectedSlice[i]
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
	size := 1024
	log, err := New(name, size, &buf)
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

		expectedSlice := []string{
			formatMsg(now, "DEBUG", tags, msg),
			formatMsg(now, "DEBUG", tags, msg),
			formatMsg(now, "FATAL", tags, msg),
		}

		scanner := bufio.NewScanner(&buf)
		i := 0
		for scanner.Scan() {
			if i >= len(expectedSlice) {
				break
			}
			got := scanner.Text()
			expected := expectedSlice[i]
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
	size := 1024
	log, err := New(name, size, &buf)
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

	expectedSlice := []string{
		formatMsg(now, "FATAL", tags, msg),
	}

	scanner := bufio.NewScanner(&buf)
	i := 0
	for scanner.Scan() {
		if i >= len(expectedSlice) {
			break
		}
		got := scanner.Text()
		expected := expectedSlice[i]
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
	size := 1024
	log, err := New(name, size, &buf)
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

	expectedSlice := []string{
		formatMsg(now, "FATAL", tags, msg),
	}

	scanner := bufio.NewScanner(&buf)
	i := 0
	for scanner.Scan() {
		if i >= len(expectedSlice) {
			break
		}
		got := scanner.Text()
		expected := expectedSlice[i]
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
	size := 1024
	log, err := newLogger(name, size, &buf)
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

func TestNew(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	size := 1024
	log, err := New(name, size, &buf)
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

	expected := formatMsg(t1, "INFO ", tags, msg)
	expected += formatMsg(t2, "INFO ", tags, msg)
	got := buf.String()
	if got != expected {
		t.Fatalf("Expected logger to write %q, but got %q", expected, got)
	}
}

func TestNewExisting(t *testing.T) {
	defer reset()
	var buf bytes.Buffer
	name := "Test"
	size := 1024
	log, err := New(name, size, &buf)
	if err != nil {
		t.Fatal("Unexpected error, when create a new logger: " + err.Error())
	}

	_, err = New(name, size, &buf)
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
	size := 1024
	log, err := NewFile(name, path, size)
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

	expected := formatMsg(t1, "INFO ", tags, msg)
	expected += formatMsg(t2, "INFO ", tags, msg)
	if got != expected {
		t.Fatalf("Expected logger to write %q, but got %q", expected, got)
	}
}

func TestCombine(t *testing.T) {
	defer reset()
	var buf1, buf2 bytes.Buffer
	name := "Test"
	size := 1024
	log1, err := New(name+"1", size, &buf1)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log2, err := New(name+"2", size, &buf2)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	log, err := Combine(name, size, log1, log2)
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
	expectedSlice := []string{
		formatMsg(now, "INFO ", tags, msg),
		formatMsg(now, "INFO ", tags, msg),
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
			expected := expectedSlice[i]
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
	size := 1024
	log1, err := New(name, size, &buf1)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}
	log2, err := New(name+"2", size, &buf2)
	if err != nil {
		t.Fatal("Unexpected error, creating a logger: " + err.Error())
	}

	_, err = Combine(name, size, log1, log2)
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
	size := 1024
	_, err := Combine(name, size)
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
