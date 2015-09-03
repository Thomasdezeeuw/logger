// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed onder the MIT license that can be found in the LICENSE file.

package logger

import "fmt"

// Tags are keywords usefull in searching through logs, for example:
//
//	tags := []Tags{"file.go", "myFn", "user:$user_id", "input:$input"}
//
// With this information you can lookup any logs for a specific user reporting
// a problem. Then you can find which function, in which file, is throwing the
// error.
type Tags []string

// String creates a comma separated list from the tags in string.
func (tags Tags) String() string {
	return string(tags.Bytes())
}

// Bytes does the same as Tags.String, but returns a byte slice.
func (tags Tags) Bytes() []byte {
	var buf []byte

	// Add each tag in the form of "tag, "
	for _, tag := range tags {
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

func (tags Tags) MarshalJSON() ([]byte, error) {
	if len(tags) == 0 {
		return []byte("[]"), nil
	}

	str := "["
	for _, tag := range tags {
		str += fmt.Sprintf("%q, ", tag)
	}
	str = str[:len(str)-2] + "]"
	return []byte(str), nil
}
