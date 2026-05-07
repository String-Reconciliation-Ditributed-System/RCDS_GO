package log

import (
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	if err := SetLogLevel([]byte("debug")); err != nil {
		t.Fatalf("expected debug level to parse: %v", err)
	}
	if err := SetLogLevel([]byte("fatal")); err != nil {
		t.Fatalf("expected legacy fatal level to parse: %v", err)
	}
	if err := SetLogLevel([]byte("definitely-not-a-level")); err == nil {
		t.Fatal("expected invalid level to error")
	}
}
