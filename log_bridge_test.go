package logger

import (
	"log"
	"reflect"
	"testing"
	"time"
)

func TestSetLogOutput(t *testing.T) {
	defer reset()

	tags := Tags{"TestSetLogOutput", "log"}
	ew := eventWriter{}
	Start(&ew)
	BridgeLogPgk(tags)

	t1 := time.Now()

	defer func() {
		if recv := recover(); recv == nil {
			t.Fatalf("Expected an panic to occur, but it didn't")
		}

		if err := Close(); err != nil {
			t.Fatal("Unexpected error calling close: ", err.Error())
		}

		expected := []Event{
			{Type: LogEvent, Timestamp: t1, Tags: tags, Message: "Log message"},
			{Type: LogEvent, Timestamp: t1, Tags: tags, Message: "Log formatted message"},
			{Type: LogEvent, Timestamp: t1, Tags: tags, Message: "Log message newline"},
			{Type: LogEvent, Timestamp: t1, Tags: tags, Message: "Panic message"},
		}

		if len(ew.events) != len(expected) {
			t.Fatalf("Expected to have %d events, but got %d",
				len(expected), len(ew.events))
		}

		const margin = 10 * time.Millisecond
		for i, event := range ew.events {
			expectedEvent := expected[i]

			// Can't mock time in the log package, so we have a truncate it.
			if !event.Timestamp.Truncate(margin).Equal(expectedEvent.Timestamp.Truncate(margin)) {
				t.Errorf("Expected event #%d to be %v, but got %v", i, expectedEvent, event)
				continue
			}
			event.Timestamp = expectedEvent.Timestamp

			if expected, got := expectedEvent, event; !reflect.DeepEqual(expected, got) {
				t.Errorf("Expected event #%d to be %v, but got %v", i, expected, got)
			}
		}
	}()

	log.Print("Log message")
	log.Printf("Log %s message", "formatted")
	log.Println("Log message newline")
	log.Panic("Panic message")
}
