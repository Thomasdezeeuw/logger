package logger

import (
	"fmt"
	"testing"
)

func TestLogLevelString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lvl      LogLevel
		expected string
	}{
		{Debug, "Debug"},
		{Thumb, "Thumb"},
		{Info, "Info"},
		{Warn, "Warn"},
		{Error, "Error"},
		{Fatal, "Fatal"},
		{LogLevel(255), "LogLevel(255)"},
	}

	for _, test := range tests {
		got, gotBytes := test.lvl.String(), test.lvl.Bytes()

		if got != string(gotBytes) {
			t.Errorf("LogLevel.Bytes() and String() don't return the same value, got %q"+
				" and %q, want %q", got, string(gotBytes), test.expected)
		} else if got != test.expected {
			t.Errorf("Expected LogLevel.String() to return %q, got %q",
				test.expected, got)
		}
	}
}

func TestNewLogLevel(t *testing.T) {
	oldLogLevelNames := logLevelNames
	oldLogLevelIndices := logLevelIndices

	for i := 1; i <= 248; i++ {
		expected := fmt.Sprintf("myLogLevel%d", i)
		myLogLevel := NewLogLevel(expected)

		if got := myLogLevel.String(); got != expected {
			t.Fatalf("Expected Loglevel %d to return %s, but got %s", int(myLogLevel),
				expected, got)
		}
	}

	defer func() {
		recv := recover()
		if recv == nil {
			t.Fatal("Expected a panic after creating 248 log levels, but didn't get one")
		}

		got, ok := recv.(string)
		if !ok {
			t.Fatal("Expected the recoverd panic to be a string, but it's %v", recv)
		}

		expected := "ini: can't have more then 255 log levels"
		if got != expected {
			t.Fatal("Expected the recoverd panic to be %s, but got %s", expected, got)
		}

		logLevelNames = oldLogLevelNames
		logLevelIndices = oldLogLevelIndices
	}()

	NewLogLevel("myLogLevel249")
}
