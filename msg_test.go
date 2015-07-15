// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"testing"
	"time"
)

func TestItoa(t *testing.T) {
	t.Parallel()

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

func TestMsg(t *testing.T) {
	t.Parallel()

	var now = time.Now()
	var tStr = now.Format("2006-01-02 15:04:05")

	var msgTests = []struct {
		msg      Msg
		expected string
	}{
		{Msg{Fatal, "Message", Tags{}, now},
			tStr + " [Fatal] : Message\n"},
		{Msg{Error, "Message", Tags{"tag1"}, now},
			tStr + " [Error] tag1: Message\n"},
		{Msg{Info, "Message", Tags{"tag1", "tag2"}, now},
			tStr + " [Info] tag1, tag2: Message\n"},
		{Msg{Debug, "Message", Tags{"tag1", "tag2", "tag3"}, now},
			tStr + " [Debug] tag1, tag2, tag3: Message\n"},
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
