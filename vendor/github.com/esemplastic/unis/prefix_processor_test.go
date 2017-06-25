package unis

import (
	"testing"
)

func TestPrefixRemover(t *testing.T) {
	tests := []originalAgainstResult{
		{"/api/users/42", "api/users/42"},
		{"//api/users/42/", "api/users/42/"},
		{"api/users/", "api/users/"},
	}

	prefixRemover := NewPrefixRemover("/")

	testOriginalAgainstResult(prefixRemover, tests, t)
}

func TestPrepender(t *testing.T) {
	tests := []originalAgainstResult{
		{"/api/users/42", "/api/users/42"},
		{"//api/users\\42", "//api/users\\42"},
		{"api\\////users/", "/api\\////users/"},
	}
	prepender := NewPrepender("/")

	testOriginalAgainstResult(prepender, tests, t)
}
func TestExclusivePrepender(t *testing.T) {
	tests := []originalAgainstResult{
		{"/api/users/42", "/api/users/42"},
		// the only difference from simple Prepender is that this ExclusivePrepender
		// will make sure that we have only one slash as a prefix.
		{"//api/users\\42", "/api/users\\42"},
		{"api\\////users/", "/api\\////users/"},
	}
	prepender := NewExclusivePrepender("/")

	testOriginalAgainstResult(prepender, tests, t)
}
