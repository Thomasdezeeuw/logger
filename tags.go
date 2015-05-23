// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

const defaultTagsSize = 50

// Tags are keywords usefull in searching logs. Examples of these are:
//	"file.go", "myFn" // indicating the location of the log operation.
//	"user:$id" // indicating a user is logged in (usefull in user specific bugs)
type Tags []string

// String creates a comma separated list from the tags in string.
func (tags *Tags) String() string {
	return string(tags.Bytes())
}

// Bytes creates a comma separated list from the tags in bytes.
func (tags *Tags) Bytes() []byte {
	buf := make([]byte, 0, defaultTagsSize)

	// Add each tag in the form of "tag, "
	for _, tag := range *tags {
		buf = append(buf, tag...)
		buf = append(buf, ',')
		buf = append(buf, ' ')
	}

	// Drop the last ", "
	if len(buf) > 2 {
		buf = buf[:len(buf)-2]
	}

	return buf
}
