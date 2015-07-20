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
	"runtime"
	"strings"
	"testing"
	"time"
)

func init() {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can't get the current file name, required for testing")
	}

	expectedMsgs[7].Msg = "Function myFunction called by github.com/" +
		"Thomasdezeeuw/logger.sendMessages, from file " + filePath + " on line 308"
}

// todo: test combine with different log levels.
// todo: check log.Errors.
// todo: send and check message setting SetMinLogLevel().
// todo: Check if logWriter() (running in goroutine) is closed.
// todo: Check if combinedLogWriter() (running in goroutine) is closed.
// todo: test with bad msgWriter, return errors.
// todo: test with bad io.Writer return errors and short writes.
// todo: test NewFile with bad filepath.

type msgWriter struct {
	msgs   []Msg
	closed bool
}

func (mw *msgWriter) Write(msg Msg) error {
	mw.msgs = append(mw.msgs, msg)
	return nil
}

func (mw *msgWriter) Close() error {
	mw.closed = true
	return nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	mw := &msgWriter{}
	log, err := New("TestNew", mw)
	if err != nil {
		t.Fatal("Unexpected error creating a new logger: " + err.Error())
	}

	// Send test messages.
	t1 := sendMessages(log)

	// Make sure all messages are written.
	if err := log.Close(); err != nil {
		t.Fatal("Unexpected error closing the logger: " + err.Error())
	} else if !mw.closed {
		t.Fatal("We closed the logger, but the msg writer isn't closed")
	}

	// Check the messages.
	if err := checkMessages(t1, mw, Debug); err != nil {
		t.Fatal(err)
	}
}

func TestNewExistingName(t *testing.T) {
	t.Parallel()
	const logName = "TestNewExistingName"
	const expectedErr = "logger: name " + logName + " already taken"

	_, err := New(logName, &msgWriter{})
	if err != nil {
		t.Fatal("Unexpected error creating a new logger: " + err.Error())
	}

	_, err = New(logName, &msgWriter{})
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
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
	const expectedErr = "logger: no logger found with name " + notLogName

	_, err := Get(notLogName)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
}

func TestNewConsole(t *testing.T) {
	t.Parallel()
	const logName = "TestNewConsole"

	var buf bytes.Buffer
	oldStderr := stderr
	stderr = &buf

	log, err := NewConsole(logName)
	if err != nil {
		t.Fatal("Unexpected error creating a new logger: " + err.Error())
	}

	t1 := sendMessages(log)

	if err := log.Close(); err != nil {
		t.Fatal("Unexpected error closing the logger: " + err.Error())
	}

	if err := checkMessagesString(t1, buf.String(), Debug); err != nil {
		t.Fatal(err)
	}

	stderr = oldStderr
}

func TestNewWriter(t *testing.T) {
	t.Parallel()
	const logName = "TestNewWriterTruncated"

	var buf bytes.Buffer

	log, err := NewWriter(logName, &buf)
	if err != nil {
		t.Fatal("Unexpected error creating a new logger: " + err.Error())
	}

	t1 := sendMessages(log)

	if err := log.Close(); err != nil {
		t.Fatal("Unexpected error closing the logger: " + err.Error())
	}

	if err := checkMessagesString(t1, buf.String(), Debug); err != nil {
		t.Fatal(err)
	}
}

func TestNewFile(t *testing.T) {
	t.Parallel()
	const logName = "TestNewFile"

	filePath := filepath.Join(os.TempDir(), "LOGGER_TEST.log")
	defer os.Remove(filePath)
	log, err := NewFile(logName, filePath)
	if err != nil {
		t.Fatal("Unexpected error creating a new logger: " + err.Error())
	}

	t1 := sendMessages(log)

	if err := log.Close(); err != nil {
		t.Fatal("Unexpected error closing the logger: " + err.Error())
	}

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal("Unexpected error opening log file: " + err.Error())
	}

	if err := checkMessagesString(t1, string(bytes), Debug); err != nil {
		t.Fatal(err)
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

	log, err := Combine(logName, log1, log2)
	if err != nil {
		t.Fatal("Unexpected error, combining loggers: " + err.Error())
	}

	t1 := sendMessages(log)

	if err := log.Close(); err != nil {
		t.Fatal("Unexpected error closing the logger: " + err.Error())
	} else if !mw1.closed || !mw2.closed {
		t.Fatal("We closed the logger, but the underlying msg writers aren't closed")
	}

	if err := checkMessages(t1, mw1, Debug); err != nil {
		t.Fatal(err)
	} else if err := checkMessages(t1, mw2, Debug); err != nil {
		t.Fatal(err)
	}
}

func TestCombineNone(t *testing.T) {
	t.Parallel()
	const logName = "TestCombineNone"
	const expectedErr = "logger: Combine requires atleast one logger"

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
	const expectedErr = "logger: name " + logName + " already taken"

	log, err := New(logName, &msgWriter{})
	if err != nil {
		t.Fatal("Unexpected error, creating a new logger: " + err.Error())
	}

	_, err = Combine(logName, log)
	if err == nil {
		t.Fatal("Expected an error when creating a logger with the same name")
	} else if err.Error() != expectedErr {
		t.Fatalf("Expected the error to be %q, but got %q",
			expectedErr, err.Error())
	}
}

type MessageData struct {
	i   int
	str string
}

var (
	defaultTags           = Tags{"test"}
	defaultCustomLogLevel = NewLogLevel("myLogLevel")

	debugMsg    = "Debug message"
	infoMsg     = "Info message"
	warnMsg     = "Warn message"
	errorMsg    = "Error message"
	fatalMsg1   = "Fatal message1"
	fatalMsg2   = "Fatal message2"
	fatalMsg3   = "0" // Actual integer used in logging.
	messageMsg  = "myLogLevel message"
	messageData = MessageData{100, "hello"}
)

func sendMessages(log *Logger) time.Time {
	t1 := time.Now().Truncate(time.Second)
	log.Debug(defaultTags, debugMsg)
	log.Info(defaultTags, infoMsg)
	log.Warn(defaultTags, warnMsg)
	log.Error(defaultTags, errors.New(errorMsg))
	log.Fatal(defaultTags, errors.New(fatalMsg1))
	log.Fatal(defaultTags, fatalMsg2)
	log.Fatal(defaultTags, 0)
	func() { // fake a unused function, to have a consistent caller.
		log.Thumbstone(defaultTags, "myFunction")
	}()
	log.Message(Msg{Level: defaultCustomLogLevel, Msg: messageMsg,
		Tags: defaultTags, Data: messageData})

	return t1
}

var expectedMsgs = []Msg{
	Msg{Level: Debug, Msg: debugMsg, Tags: defaultTags},
	Msg{Level: Info, Msg: infoMsg, Tags: defaultTags},
	Msg{Level: Warn, Msg: warnMsg, Tags: defaultTags},
	Msg{Level: Error, Msg: errorMsg, Tags: defaultTags},
	Msg{Level: Fatal, Msg: fatalMsg1, Tags: defaultTags},
	Msg{Level: Fatal, Msg: fatalMsg2, Tags: defaultTags},
	Msg{Level: Fatal, Msg: fatalMsg3, Tags: defaultTags},
	Msg{Level: Thumb, Tags: defaultTags}, // Msg added by init.
	Msg{Level: defaultCustomLogLevel, Msg: messageMsg, Tags: defaultTags,
		Data: messageData},
}

func checkMessages(t1 time.Time, mw *msgWriter, minLevel LogLevel) error {
	// Get the message we can expect.
	if minLevel > Fatal {
		minLevel += 2 // We send 3 fatal messages.
	}
	var msgs = make([]Msg, len(expectedMsgs)-int(minLevel))
	copy(msgs, expectedMsgs[int(minLevel):])

	if len(mw.msgs) != len(msgs) {
		return fmt.Errorf("Expected to get %d messages, but got %d messages",
			len(msgs), len(mw.msgs))
	}

	for i, got := range mw.msgs {
		expected := msgs[i]

		// Add our timestamp to the expected message and truncate the time of the
		// actual message, because 1 second is accurate enough.
		expected.Timestamp = t1
		got.Timestamp = got.Timestamp.Truncate(time.Second)

		if expected.Level == Fatal {
			bytes, ok := got.Data.([]byte)
			if !ok {
				return fmt.Errorf("Expected message %d to have a stack trace, but it "+
					"didn't, got: %v", got.Data)
			} else if len(bytes) < 10 {
				return fmt.Errorf("The expected stacktrace seems empty, got: %s",
					string(bytes))
			}

			// Can't compare the strack trace any better.
			got.Data = nil
		}

		if !reflect.DeepEqual(got, expected) {
			return fmt.Errorf("Expected message %d to be %v, but got %v",
				i, expected, got)
		}
	}

	return nil
}

func checkMessagesString(t1 time.Time, gotString string, minLevel LogLevel) error {
	// Get the message we can expect.
	if minLevel > Fatal {
		minLevel += 2 // We send 3 fatal messages.
	}
	var msgs = make([]Msg, len(expectedMsgs)-int(minLevel))
	copy(msgs, expectedMsgs[int(minLevel):])

	i := 0
	s := bufio.NewScanner(strings.NewReader(gotString))
	t1Str := t1.UTC().Format(TimeFormat)

	for s.Scan() {
		got := s.Text()

		if !strings.HasPrefix(got, t1Str) {
			// Likely a stacktrace from Logger.Fatal.
			continue
		}

		if i >= len(msgs) {
			return fmt.Errorf("Unexpected log message #%d: %s", i+1, got)
		}
		expectedMsg := msgs[i]
		expectedMsg.Timestamp = t1
		expected := expectedMsg.String()

		if expectedMsg.Level == Fatal {
			// Trim the stacktrace data.
			i := strings.LastIndex(got, ",")
			if i != -1 {
				got = got[:i]
			}
		}

		if got != expected {
			return errors.New(compareError("Log message #%d", i+1, expected, got))
		}
		i++
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("Unexpected scanning error: %s", err.Error())
	}

	if i != len(msgs) {
		return fmt.Errorf("Expected %d log messages, but got %d", len(msgs), i)
	}

	return nil
}

func compareError(format string, v interface{}, expected, got string) string {
	return fmt.Sprintf("String compare error trying to compare: "+
		format+"\nexpected: %q\ngot:      %q\n", v, expected, got)
}
