package unis

import (
	"testing"
)

type originalAgainstResult struct {
	original string
	result   string
}

func testOriginalAgainstResult(p Processor, tests []originalAgainstResult, t *testing.T) {
	for i, tt := range tests {
		if expected, got := tt.result, p.Process(tt.original); expected != got {
			t.Fatalf("%s[%d] - expected '%s' but got '%s'", getTestName(t), i, expected, got)
		}
	}
}
