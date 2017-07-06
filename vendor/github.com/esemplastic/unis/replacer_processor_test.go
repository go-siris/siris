package unis

import (
	"testing"
)

type replacements map[string]string

func TestReplacer(t *testing.T) {
	tests := []struct {
		original string
		result   string
		repl     replacements
	}{
		{"hello world!!!", "hello world", replacements{"!": ""}},
		{"c:\\path//", "c:/path/", replacements{"\\": "/", "//": "/"}},
	}

	for i, tt := range tests {
		replacer := NewReplacer(tt.repl)

		if expected, got := tt.result, replacer.Process(tt.original); expected != got {
			t.Fatalf("[%d] - expected '%s' but got '%s'", i, expected, got)
		}
	}

}
