package unis

import (
	"testing"
)

func TestTargetedJoiner(t *testing.T) {
	tests := []originalAgainstResult{
		{"/api/users/42", "/api/users/42"},
		{"//api/users\\42", "/api/users/42"},
		{"api\\////users/", "/api/users"},
	}

	testOriginalAgainstResult(normalizePath, tests, t)
}
