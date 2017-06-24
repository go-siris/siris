package unis

import (
	"path"
	"strings"
	"testing"
)

func newPathNormalizer() Processor {
	slash := "/"
	replacer := NewReplacer(map[string]string{
		`\`:  slash,
		`//`: slash,
	})

	suffixRemover := NewSuffixRemover(slash)
	slashPrepender := NewTargetedJoiner(0, slash[0])

	toLower := ProcessorFunc(strings.ToLower)
	pathCleaner := ProcessorFunc(path.Clean)
	return NewChain(
		replacer,
		suffixRemover,
		slashPrepender,
		pathCleaner,
		toLower,
	)
}

var normalizePath = newPathNormalizer()

func TestChain(t *testing.T) {
	tests := []originalAgainstResult{
		{"/api/users/42", "/api/users/42"},
		{"//api/users\\42", "/api/users/42"},
		{"api\\////users/", "/api/users"},
	}

	testOriginalAgainstResult(normalizePath, tests, t)
}
