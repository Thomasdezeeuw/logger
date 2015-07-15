// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"testing"
	"time"
)

func TestMsg(t *testing.T) {
	t.Parallel()

	var now = time.Now()
	var tStr = now.UTC().Format("2006-01-02 15:04:05")

	var msgTests = []struct {
		msg      Msg
		expected string
	}{
		{Msg{Fatal, "Message", Tags{}, now},
			tStr + " [Fatal] : Message"},
		{Msg{Error, "Message", Tags{"tag1"}, now},
			tStr + " [Error] tag1: Message"},
		{Msg{Info, "Message", Tags{"tag1", "tag2"}, now},
			tStr + " [Info] tag1, tag2: Message"},
		{Msg{Debug, "Message", Tags{"tag1", "tag2", "tag3"}, now},
			tStr + " [Debug] tag1, tag2, tag3: Message"},
	}

	for _, test := range msgTests {
		got, gotBytes := test.msg.String(), test.msg.Bytes()

		if got != string(gotBytes) {
			t.Errorf("Msg.Bytes() and String() don't return the same value, got %q"+
				" and %q, want %q", got, string(gotBytes), test.expected)
		} else if got != test.expected {
			t.Errorf("Expected Msg.String() to return %q, got %q",
				test.expected, got)
		}
	}
}

func TestLogLevelString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lvl      LogLevel
		expected string
	}{
		{Debug, "Debug"},
		{Thumb, "Thumb"},
		{Info, "Info"},
		{Warn, "Warn"},
		{Error, "Error"},
		{Fatal, "Fatal"},
	}

	for _, test := range tests {
		got := test.lvl.String()
		if got != test.expected {
			t.Fatalf("Expected Loglevel %d to return %s, but got %s", int(test.lvl),
				test.expected, got)
		}
	}
}

func TestNewLogLevel(t *testing.T) {
	for i := 1; i <= 248; i++ {
		expected := fmt.Sprintf("myLogLevel%d", i)
		myLogLevel := NewLogLevel(expected)

		if got := myLogLevel.String(); got != expected {
			t.Fatalf("Expected Loglevel %d to return %s, but got %s", int(myLogLevel),
				expected, got)
		}
	}
}
