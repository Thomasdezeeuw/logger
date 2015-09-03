// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import "testing"

func TestTags(t *testing.T) {
	t.Parallel()

	var tagTests = []struct {
		tags         Tags
		expected     string
		expectedJSON string
	}{
		{Tags{}, "", "[]"},
		{Tags{"tag1"}, "tag1", `["tag1"]`},
		{Tags{"tag1", "tag2"}, "tag1, tag2", `["tag1", "tag2"]`},
		{Tags{"tag1", "tag2", "tag3"}, "tag1, tag2, tag3", `["tag1", "tag2", "tag3"]`},
	}

	for _, test := range tagTests {
		got, gotBytes := test.tags.String(), string(test.tags.Bytes())
		if gotBytes != test.expected {
			t.Errorf("Expected %#v.Bytes() to return %q, but got %q",
				test.tags, test.expected, gotBytes)
		} else if got != test.expected {
			t.Errorf("Expected %#v.String() to return %q, but got %q",
				test.tags, test.expected, got)
		}

		if json, err := test.tags.MarshalJSON(); err != nil {
			t.Errorf("Unexpected error marshaling %v into json: %s", test.tags, err.Error())
		} else if got := string(json); got != test.expectedJSON {
			t.Errorf("Expected %#v.MarshalJSON() to return %q, but got %q",
				test.tags, test.expectedJSON, got)
		}
	}
}
