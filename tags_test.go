// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import "testing"

func TestTags(t *testing.T) {
	var tagTests = []struct {
		tags     Tags
		expected string
	}{
		{Tags{}, ""},
		{Tags{"tag1"}, "tag1"},
		{Tags{"tag1", "tag2"}, "tag1, tag2"},
		{Tags{"tag1", "tag2", "tag3"}, "tag1, tag2, tag3"},
	}

	for _, test := range tagTests {
		got, gotB := test.tags.String(), test.tags.Bytes()

		if got != string(gotB) {
			t.Errorf("Tags.Bytes() and String() don't return the same value, got %q"+
				" and %q, want %q", got, string(gotB), test.expected)
		} else if got != test.expected {
			t.Errorf("Expected Tags.String() to return %q, got %q",
				test.expected, got)
		}
	}
}
