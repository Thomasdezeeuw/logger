// Copyright (C) 2015 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

import "strconv"

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
	if len(tags) == 0 {
		return []byte{}
	}

	// Add each tag in the form of "tag, ".
	var buf []byte
	for _, tag := range tags {
		buf = append(buf, tag...)
		buf = append(buf, ',')
		buf = append(buf, ' ')
	}

	// Drop the last ", ".
	buf = buf[:len(buf)-2]
	return buf
}

func (tags Tags) MarshalJSON() ([]byte, error) {
	if len(tags) == 0 {
		return []byte("[]"), nil
	}

	// Add each tag in the form of `"tag", `
	buf := []byte("[")
	for _, tag := range tags {
		qoutedTag := strconv.Quote(tag)
		buf = append(buf, qoutedTag...)
		buf = append(buf, ',')
		buf = append(buf, ' ')
	}

	// Drop the last "," and a closing bracket.
	buf = append(buf[:len(buf)-2], ']')
	return buf, nil
}
