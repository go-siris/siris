package unis

import (
	"testing"
)

var (
	// invert if indiciator not found
	// because we need the first parameter to be the subdomain
	// even if empty, but the second parameter
	// should be the path, in order to normalize it
	// (because of the reason of subdomains shouldn't be normalized as path)
	subdomainDevider = NewInvertOnFailureDivider(NewDivider("./"))
)

func divideSubdomainPath(fullpath string) (path string, subdomain string) {
	subdomain, path = subdomainDevider.Divide(fullpath)
	return subdomain, normalizePath.Process(path)
}

func TestDivider(t *testing.T) {
	tests := []struct {
		original  string
		subdomain string
		path      string
	}{
		{"admin./users/42", "admin.", "/users/42"},
		{"//api/users\\42", "", "/api/users/42"},
		{"admin./users/\\42", "admin.", "/users/42"},
	}

	for i, tt := range tests {
		subdomain, path := divideSubdomainPath(tt.original)

		if expected, got := tt.subdomain, subdomain; expected != got {
			t.Fatalf("[%d] - expected subdomain '%s' but got '%s'", i, expected, got)
		}
		if expected, got := tt.path, path; expected != got {
			t.Fatalf("[%d] - expected path '%s' but got '%s'", i, expected, got)
		}
	}
}

func TestDivide(t *testing.T) {
	// it's being tested to TestDivider we have a .Divide function too

	part1, part2 := Divide("admin./users/42", "./")

	if expected, got := "admin.", part1; expected != got {
		t.Fatalf("[0] expected part1 to be '%s' but got '%s'", expected, got)
	}

	if expected, got := "/users/42", part2; expected != got {
		t.Fatalf("[1] expected part2 to be '%s' but got '%s'", expected, got)
	}
}
