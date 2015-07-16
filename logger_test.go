// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type msgWriter struct {
	msgs []Msg
}

func (mw *msgWriter) Write(msg Msg) error {
	mw.msgs = append(mw.msgs, msg)
	return nil
}

func (mw *msgWriter) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	const logName = "TestNew"
	mw := &msgWriter{}
	log, err := New(logName, mw)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	t1, logLevel, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMessages(t1, mw, logLevel); err != nil {
		t.Fatal(err)
	}
}

func TestNewExistingName(t *testing.T) {
	t.Parallel()

	const logName = "TestNewExistingName"
	_, err := New(logName, &msgWriter{})
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	expectedErr := "logger: name " + logName + " already taken"
	_, err = New(logName, &msgWriter{})
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
}

func TestNewFile(t *testing.T) {
	t.Parallel()

	const logName = "TestNewFile"
	path := filepath.Join(os.TempDir(), "LOGGER_TEST.log")
	defer os.Remove(path)
	log, err := NewFile(logName, path)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	t1, _, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal("Unexpected error opening log file: " + err.Error())
	}

	if err := checkMessagesString(t1, b); err != nil {
		t.Fatal(err)
	}
}

func TestNewConsole(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	oldStderr := stderr
	stderr = &buf

	const logName = "TestNewConsole"
	log, err := NewConsole(logName)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	t1, _, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMessagesString(t1, buf.Bytes()); err != nil {
		t.Fatal(err)
	}

	stderr = oldStderr
}

func TestGet(t *testing.T) {
	t.Parallel()

	const logName = "TestGet"
	log1, err := New(logName, &msgWriter{})
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	log2, err := Get(logName)
	if err != nil {
		t.Fatal("Unexpected error, getting the logger: " + err.Error())
	}

	if log1 != log2 {
		t.Fatal("Expected the created logger to be the same as the gotten logger")
	}
}

func TestGetNotFound(t *testing.T) {
	t.Parallel()

	const notLogName = "A logger which doesn't exists"
	expectedErr := "logger: no logger found with name " + notLogName
	_, err := Get(notLogName)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
}

func TestCombine(t *testing.T) {
	t.Parallel()

	const logName = "TestCombine"
	mw1 := &msgWriter{}
	log1, err := New(logName+"1", mw1)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	mw2 := &msgWriter{}
	log2, err := New(logName+"2", mw2)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}
	log2.ShowDebug = true

	log, err := Combine(logName, log1, log2)
	if err != nil {
		t.Fatal("Unexpected error, combining loggers: " + err.Error())
	}

	t1, logLevel, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMessages(t1, mw1, logLevel); err != nil {
		t.Fatal(err)
	} else if err := checkMessages(t1, mw2, logLevel); err != nil {
		t.Fatal(err)
	}
}

func TestCombineNone(t *testing.T) {
	t.Parallel()

	const logName = "TestCombineNone"
	expectedErr := "logger: Combine requires atleast one logger"
	_, err := Combine(logName)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
}

func TestCombineExistingName(t *testing.T) {
	t.Parallel()

	const logName = "TestCombineExistingName"
	log, err := New(logName, &msgWriter{})
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	expectedErr := "logger: name " + logName + " already taken"
	_, err = Combine(logName, log)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
}

// sendMessages is linked with checkMessages and checkMessagesString, any
// changes most be checked in those functions aswell.
func sendMessages(log *Logger) (time.Time, LogLevel, error) {
	tags := Tags{"test"}
	t1 := time.Now().UTC().Truncate(time.Second)
	myLogLevel := NewLogLevel("myLogLevel")
	msg := Msg{Level: myLogLevel, Tags: tags}

	log.Fatal(tags, errors.New("Fatal message1"))
	log.Fatal(tags, "Fatal message2")
	log.Fatal(tags, NewLogLevel("Fatal message3"))
	log.Error(tags, errors.New("Error message1"))
	log.Error(tags, errors.New("Error message2"))
	log.Error(tags, errors.New("Error message3"))
	log.Warn(tags, "Warn message1")
	log.Warn(tags, "Warn message2")
	log.Warn(tags, "Warn message3")
	log.Info(tags, "Info message1")
	log.Info(tags, "Info message2")
	log.Info(tags, "Info message3")
	log.Thumbstone(tags, "Thumb message1")
	log.Thumbstone(tags, "Thumb message2")
	log.Thumbstone(tags, "Thumb message3")
	log.ShowDebug = true
	log.Debug(tags, "Debug message1")
	log.Debug(tags, "Debug message2")
	log.Debug(tags, "Debug message3")
	log.ShowDebug = false
	log.Debug(tags, "Debug message4")
	msg.Msg = "myLogLevel message1"
	log.Message(msg)
	msg.Msg = "myLogLevel message2"
	log.Message(msg)
	msg.Msg = "myLogLevel message3"
	log.Message(msg)

	if err := log.Close(); err != nil {
		return t1, myLogLevel, errors.New("Unexpected error, closing logger: " + err.Error())
	}

	return t1, myLogLevel, nil
}

// checkMessages is linked to sendMessages.
func checkMessages(t1 time.Time, mw *msgWriter, myLogLevel LogLevel) error {
	if nMsgs := 21; len(mw.msgs) != nMsgs {
		return fmt.Errorf("Expected %d messages, but got %d", nMsgs, len(mw.msgs))
	}

	for i, msg := range mw.msgs {
		var expectedLevel = myLogLevel
		if i < 3 {
			expectedLevel = Fatal
		} else if i >= 3 && i < 6 {
			expectedLevel = Error
		} else if i >= 6 && i < 9 {
			expectedLevel = Warn
		} else if i >= 9 && i < 12 {
			expectedLevel = Info
		} else if i >= 12 && i < 15 {
			expectedLevel = Thumb
		} else if i >= 15 && i < 18 {
			expectedLevel = Debug
		}

		expectedMsg := expectedLevel.String()
		expectedMsg = strings.TrimSpace(expectedMsg) + " message"
		expectedMsg += fmt.Sprintf("%d", i%3+1)

		if expectedLevel == Fatal {
			msg.Data = nil // Drop the stack trace, it's never the same.
		}

		tags := Tags{"test"}
		if expectedLevel == Thumb {
			tags = Tags{"thumbstone", "test"}
		}

		if msg.Level != expectedLevel {
			return fmt.Errorf("Expected msg.Level to be %q, but got %q",
				expectedLevel, msg.Level)
		} else if msg.Msg != expectedMsg {
			return fmt.Errorf("Expected msg.Msg to be %q, but got %q",
				expectedMsg, msg.Msg)
		} else if !reflect.DeepEqual(msg.Tags, tags) {
			return fmt.Errorf("Expected msg.Tags to be %q, but got %q",
				tags.String(), msg.Tags.String())
		} else if !msg.Timestamp.Truncate(time.Second).Equal(t1) {
			return fmt.Errorf("Expected msg.Timestamp to be %v, but got %v",
				t1, msg.Timestamp)
		}
	}

	return nil
}

// checkMessagesString is linked to sendMessages.
func checkMessagesString(t1 time.Time, gotBytes []byte) error {
	t1Str := t1.Format("2006-01-02 15:04:05")

	// not the prettiest solution, but good enough...
	expectedLines := []string{
		t1Str + " [Fatal] test: Fatal message1",
		t1Str + " [Fatal] test: Fatal message2",
		t1Str + " [Fatal] test: Fatal message3",
		t1Str + " [Error] test: Error message1",
		t1Str + " [Error] test: Error message2",
		t1Str + " [Error] test: Error message3",
		t1Str + " [Warn] test: Warn message1",
		t1Str + " [Warn] test: Warn message2",
		t1Str + " [Warn] test: Warn message3",
		t1Str + " [Info] test: Info message1",
		t1Str + " [Info] test: Info message2",
		t1Str + " [Info] test: Info message3",
		t1Str + " [Thumb] thumbstone, test: Thumb message1",
		t1Str + " [Thumb] thumbstone, test: Thumb message2",
		t1Str + " [Thumb] thumbstone, test: Thumb message3",
		t1Str + " [Debug] test: Debug message1",
		t1Str + " [Debug] test: Debug message2",
		t1Str + " [Debug] test: Debug message3",
		t1Str + " [myLogLevel] test: myLogLevel message1",
		t1Str + " [myLogLevel] test: myLogLevel message2",
		t1Str + " [myLogLevel] test: myLogLevel message3",
	}

	i := 0
	s := bufio.NewScanner(bytes.NewReader(gotBytes))
	for s.Scan() {
		got := s.Text()
		expected := expectedLines[i]

		if !strings.HasPrefix(got, t1Str) {
			continue
		} else if got[21:26] == Fatal.String() {
			// Trim the stacktrace data, it's never the same.
			got = got[:48]
		}

		if got != expected {
			return fmt.Errorf("Error comparing line %d\nExpected: %q\nbut got:  %q",
				i, expected, got)
		}
		i++
	}

	if i != len(expectedLines) {
		return errors.New("Didn't get the same amount of lines as expected")
	} else if err := s.Err(); err != nil {
		return fmt.Errorf("Unexpected scanning error: %s", err.Error())
	}
	return nil
}
