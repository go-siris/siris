package logger

import (
	"bytes"
	"testing"
)

func TestNewFromWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	expected := "hello"
	logger := NewFromWriter(buf)

	logger(expected)
	if got := buf.String(); expected != got {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}
