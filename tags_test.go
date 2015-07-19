// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import "testing"

func TestTags(t *testing.T) {
	t.Parallel()

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
		got, gotBytes := test.tags.String(), string(test.tags.Bytes())

		if gotBytes != test.expected {
			t.Error(compareError("Tags%v.Bytes()", test.tags, test.expected, gotBytes))
		} else if got != test.expected {
			t.Error(compareError("Tags%v.String()", test.tags, test.expected, got))
		}
	}
}
