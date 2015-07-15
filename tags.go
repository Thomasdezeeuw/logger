// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

// Tags are keywords usefull in searching the logs, for example:
//
//	tags := []Tags{"file.go", "myFn", "user:$id", "input:$input"}
//
// With this information you can lookup any logs for a specific user reporting
// problems. Then you can find which function, in which file, is throwing the
// error.
type Tags []string

// String creates a comma separated list from the tags in string.
func (tags *Tags) String() string {
	return string(tags.Bytes())
}

// Bytes does the same as Tags.String, but returns a byte slice.
func (tags *Tags) Bytes() []byte {
	var buf []byte

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
