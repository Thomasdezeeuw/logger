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
	const logName = "TestNew"
	mw := &msgWriter{}
	log, err := New(logName, mw)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	t1, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMessages(t1, mw); err != nil {
		t.Fatal(err)
	}
}

func TestNewExistingName(t *testing.T) {
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
	const logName = "TestNewFile"
	path := filepath.Join(os.TempDir(), "LOGGER_TEST.log")
	defer os.Remove(path)
	log, err := NewFile(logName, path)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	t1, err := sendMessages(log)
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
	var buf bytes.Buffer
	oldStderr := stderr
	stderr = &buf

	const logName = "TestNewConsole"
	log, err := NewConsole(logName)
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	t1, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMessagesString(t1, buf.Bytes()); err != nil {
		t.Fatal(err)
	}

	stderr = oldStderr
}

func TestGet(t *testing.T) {
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

	log, err := Combine(logName, log1, log2)
	if err != nil {
		t.Fatal("Unexpected error, combining loggers: " + err.Error())
	}

	t1, err := sendMessages(log)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMessages(t1, mw1); err != nil {
		t.Fatal(err)
	} else if err := checkMessages(t1, mw2); err != nil {
		t.Fatal(err)
	}
}

func TestCombineNone(t *testing.T) {
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
func sendMessages(log *Logger) (time.Time, error) {
	log.ShowDebug = true
	tags := Tags{"test"}
	t1 := time.Now().Truncate(time.Second)
	log.Fatal(tags, errors.New("Fatal message1"))
	log.Fatal(tags, errors.New("Fatal message2"))
	log.Fatal(tags, "Fatal message3")
	log.Error(tags, errors.New("Error message1"))
	log.Error(tags, errors.New("Error message2"))
	log.Error(tags, errors.New("Error message3"))
	log.Info(tags, "Info message1")
	log.Info(tags, "Info message2")
	log.Info(tags, "Info message3")
	log.Debug(tags, "Debug message1")
	log.Debug(tags, "Debug message2")
	log.Debug(tags, "Debug message3")
	log.Thumbstone("Thumb message1")
	log.Thumbstone("Thumb message2")
	log.Thumbstone("Thumb message3")

	log.ShowDebug = false
	log.Debug(tags, "Debug message4")

	if err := log.Close(); err != nil {
		return t1, errors.New("Unexpected error, closing logger: " + err.Error())
	}

	return t1, nil
}

// checkMessages is linked to sendMessages.
func checkMessages(t1 time.Time, mw *msgWriter) error {
	if len(mw.msgs) != 15 {
		return fmt.Errorf("Expected 15 messages, but got %d", len(mw.msgs))
	}

	tags := Tags{"test"}
	for i, msg := range mw.msgs {
		var expectedLevel string
		if i < 3 {
			expectedLevel = FatalLevel
		} else if i >= 3 && i < 6 {
			expectedLevel = ErrorLevel
		} else if i >= 6 && i < 9 {
			expectedLevel = InfoLevel
		} else if i >= 9 && i < 12 {
			expectedLevel = DebugLevel
		} else {
			expectedLevel = ThumbLevel
		}

		expectedMsg := expectedLevel[:1] + strings.ToLower(expectedLevel[1:])
		expectedMsg = strings.TrimSpace(expectedMsg) + " message"
		expectedMsg += fmt.Sprintf("%d", i%3+1)

		if expectedLevel == FatalLevel {
			msg.Msg = msg.Msg[:14] // trim stack trace from message.
		} else if expectedLevel == ThumbLevel {
			tags = Tags{"thumbstone"}
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

	// Remove the stack traces from the output and only keep the message lines.
	var got string
	s := bufio.NewScanner(bytes.NewReader(gotBytes))
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, t1Str[:4]) {
			got += line + "\n"
		}
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("Unexpected scanning error: %s", err.Error())
	}

	// not the prettiest solution, but good enough...
	expected := t1Str + " [FATAL] test: Fatal message1\n"
	expected += t1Str + " [FATAL] test: Fatal message2\n"
	expected += t1Str + " [FATAL] test: Fatal message3\n"
	expected += t1Str + " [ERROR] test: Error message1\n"
	expected += t1Str + " [ERROR] test: Error message2\n"
	expected += t1Str + " [ERROR] test: Error message3\n"
	expected += t1Str + " [INFO ] test: Info message1\n"
	expected += t1Str + " [INFO ] test: Info message2\n"
	expected += t1Str + " [INFO ] test: Info message3\n"
	expected += t1Str + " [DEBUG] test: Debug message1\n"
	expected += t1Str + " [DEBUG] test: Debug message2\n"
	expected += t1Str + " [DEBUG] test: Debug message3\n"
	expected += t1Str + " [THUMB] thumbstone: Thumb message1\n"
	expected += t1Str + " [THUMB] thumbstone: Thumb message2\n"
	expected += t1Str + " [THUMB] thumbstone: Thumb message3\n"

	if got != expected {
		return fmt.Errorf("Expected the log file to contain: \n%s\nbut got: \n%s",
			expected, got)
	}
	return nil
}
