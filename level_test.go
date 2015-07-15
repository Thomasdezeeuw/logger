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
	}

	for _, test := range tests {
		got := test.lvl.String()
		if got != test.expected {
			t.Fatalf("Expected Loglevel %d to return %s, but got %s", int(test.lvl),
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

	logLevelNames = oldLogLevelNames
	logLevelIndices = oldLogLevelIndices
}
