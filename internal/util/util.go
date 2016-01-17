// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package util

import "fmt"

// InterfaceToString converts a interface{} variable to a string.
func InterfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case []byte:
		return string(v)
	case error:
		return v.Error()
	}
	return fmt.Sprintf("%v", value)
}

// InterfacesToString converts mulitple empty interfaces into a single string.
func InterfacesToString(value []interface{}) string {
	var str string
	for _, v := range value {
		str += InterfaceToString(v)
		str += " "
	}
	return str[:len(str)-1]
}
