// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
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
