// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"testing"
	"time"
)

type stringer struct{}

func (s *stringer) String() string {
	return "data"
}

func TestMsg(t *testing.T) {
	t.Parallel()

	var now = time.Now()
	var tStr = now.UTC().Format("2006-01-02 15:04:05")

	var msgTests = []struct {
		msg      Msg
		expected string
	}{
		{Msg{Fatal, "Message1", Tags{}, now, nil},
			tStr + " [Fatal] : Message1"},
		{Msg{Error, "Message2", Tags{"tag1"}, now, "data"},
			tStr + " [Error] tag1: Message2, data"},
		{Msg{Warn, "Message3", Tags{"tag1"}, now, &stringer{}},
			tStr + " [Warn] tag1: Message3, data"},
		{Msg{Info, "Message4", Tags{"tag1", "tag2"}, now, []byte("data")},
			tStr + " [Info] tag1, tag2: Message4, data"},
		{Msg{Thumb, "Message5", Tags{"tag1", "tag2", "tag3"}, now, errors.New("error data")},
			tStr + " [Thumb] tag1, tag2, tag3: Message5, error data"},
		{Msg{Debug, "Message6", Tags{"tag1", "tag2", "tag3"}, now, 0},
			tStr + " [Debug] tag1, tag2, tag3: Message6, 0"},
	}

	for _, test := range msgTests {
		got, gotBytes := test.msg.String(), test.msg.Bytes()

		if got != string(gotBytes) {
			t.Errorf("Expected %#v.Bytes() and .String() to return the same value, "+
				"but got %q and %q, and want %q", got, string(gotBytes), test.expected)
		} else if got != test.expected {
			t.Errorf("Expected %#v to return %q, got %q",
				test.msg, test.expected, got)
		}
	}
}
