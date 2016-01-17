// Copyright (C) 2015-2016 Thomas de Zeeuw.
//
// Licensed under the MIT license that can be found in the LICENSE file.

package util

import (
	"errors"
	"testing"
)

type stringer int

func (stringer) String() string {
	return "string123"
}

func TestInterfaceToString(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected string
	}{
		{"string", "string"},
		{stringer(123), "string123"},
		{[]byte("string"), "string"},
		{errors.New("string"), "string"},
		{123, "123"},
	}

	for _, test := range tests {
		got := InterfaceToString(test.value)

		if got != test.expected {
			t.Fatalf("Expected InterfaceToString(%#v) to return %s, but got %s",
				test.value, test.expected, got)
		}
	}
}

func TestInterfacesToString(t *testing.T) {
	tests := []struct {
		values   []interface{}
		expected string
	}{
		{[]interface{}{"string"}, "string"},
		{[]interface{}{"string", 123}, "string 123"},
		{[]interface{}{stringer(123), []byte("string"), errors.New("string")}, "string123 string string"},
	}

	for _, test := range tests {
		got := InterfacesToString(test.values)

		if got != test.expected {
			t.Fatalf("Expected InterfaceToString(%#v) to return %s, but got %s",
				test.values, test.expected, got)
		}
	}
}
