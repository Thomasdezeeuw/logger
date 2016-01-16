// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import (
	"reflect"
	"testing"
)

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
		{Tags{`tag"1"`}, `tag"1"`, `["tag\"1\""]`},
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

func TestTagsAppend(t *testing.T) {
	t.Parallel()

	tags := make(Tags, 2, 3)
	tags[0] = "tag1"
	tags[1] = "tag2"

	var tests = []struct {
		originalTags Tags
		addedTags    []string
		expected     Tags
	}{
		{Tags{}, []string{}, Tags{}},
		{make(Tags, 0, 1), []string{"tag1"}, Tags{"tag1"}},
		{Tags{}, []string{"tag1"}, Tags{"tag1"}},
		{Tags{"tag1"}, []string{"tag2"}, Tags{"tag1", "tag2"}},
		{Tags{"tag1"}, []string{"tag2", "tag3", "tag4", "tag5", "tag5", "tag6",
			"tag7", "tag8", "tag9", "tag10"}, Tags{"tag1", "tag2", "tag3", "tag4",
			"tag5", "tag5", "tag6", "tag7", "tag8", "tag9", "tag10"}},
		{tags, []string{"tag3"}, Tags{"tag1", "tag2", "tag3"}},
	}

	for _, test := range tests {
		var copyOriginalTags = make(Tags, len(test.originalTags))
		copy(copyOriginalTags, test.originalTags)

		got := test.originalTags.Append(test.addedTags...)

		if !reflect.DeepEqual(copyOriginalTags, test.originalTags) {
			t.Fatalf("Expected the original tags to be uneffected and get %v, but got %v",
				copyOriginalTags, test.originalTags)
		}

		if !reflect.DeepEqual(test.expected, got) {
			t.Fatalf("Expected %#v.Append(%v) to return %v, but got %v",
				test.originalTags, test.addedTags, test.expected, got)
		}
	}
}
