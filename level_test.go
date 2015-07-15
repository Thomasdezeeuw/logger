package logger

import "testing"

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
