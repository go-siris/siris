package unis

import (
	"strings"
	"testing"
)

// Tests:
// If
// NewMatcher
// IsMail
// NewExclusivePrepender
func TestIsMail(t *testing.T) {
	mailToPreffixer := NewExclusivePrepender("mailto:")
	mailTo := If(IsMail, mailToPreffixer, ClearProcessor)
	// it checks if a string is an e-mail,if so then it runs the "mailToPreffixer"
	// otherwise it runs the ClearProcessor which clears the string and returns an empty string.

	tests := []originalAgainstResult{
		{"mail1@hotmail.com", "mailto:mail1@hotmail.com"},
		{"mail1hotmail.com", ""},
		{"@google.com", ""},
		{"mail@google", ""},
		{"mail2@google.com.dot", "mailto:mail2@google.com.dot"},
		{"mail2@google.com.dot.2", ""},
	}

	testOriginalAgainstResult(mailTo, tests, t)
}

func TestMatcherPanics(t *testing.T) {
	Logger = func(string) {} // disable the print of the error to the console.
	m := NewMatcher("\xf8\xa1\xa1\xa1\xa1")
	expected := false
	expectedErrMessagePrefix := "error parsing regexp: invalid UTF-8:"

	got, err := m.Valid("this_is_an_invalid_regexp_@expression")

	if err == nil {
		t.Fatalf("error should be filled with the correct message because the regexp is invalid")
	} else if err != nil && !strings.HasPrefix(err.Error(), expectedErrMessagePrefix) {
		t.Fatalf("expected error to starts with '%s' but got '%s'", expectedErrMessagePrefix, err.Error())
	}

	if expected != got {
		t.Fatalf("[0] expected false but got true")
	}

	got, _ = m.Valid("")
	if expected != got {
		t.Fatalf("[1] expected false but got true")
	}

	got, _ = m.Valid("string")
	if expected != got {
		t.Fatalf("[2] expected false but got true")
	}
}
