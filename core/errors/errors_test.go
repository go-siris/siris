// black-box testing
package errors_test

import (
	"fmt"
	"testing"

	"github.com/go-siris/siris/core/errors"
	"github.com/stretchr/testify/assert"
)

var errMessage = "User with mail: %s already exists"
var errUserAlreadyExists = errors.New(errMessage)
var userMail = "user1@mail.go"
var expectedUserAlreadyExists = "User with mail: user1@mail.go already exists"

func ExampleError() {

	fmt.Print(errUserAlreadyExists.Format(userMail).Append("Please change your mail addr"))

	// Output:
	// User with mail: user1@mail.go already exists
	// Please change your mail addr
}

func do(method string, testErr errors.Error, expectingMsg string, t *testing.T) {
	formattedErr := func() error {
		return testErr.Format(userMail)
	}()

	if formattedErr.Error() != expectingMsg {
		t.Fatalf("error %s failed, expected:\n%s got:\n%s", method, expectingMsg, formattedErr.Error())
	}
}

func TestFormat(t *testing.T) {
	expected := errors.Prefix + expectedUserAlreadyExists
	do("Format Test", errUserAlreadyExists, expected, t)
}

func TestAppendErr(t *testing.T) {

	errChangeMailMsg := "Please change your mail addr"
	errChangeMail := fmt.Errorf(errChangeMailMsg) // test go standard error
	errAppended := errUserAlreadyExists.AppendErr(errChangeMail)
	expectedErrorMessage := errUserAlreadyExists.Format(userMail).Error() + "\n" + errChangeMailMsg

	do("Append Test Standard error type", errAppended, expectedErrorMessage, t)
}

func TestAppendError(t *testing.T) {
	errors.Prefix = "error: "

	errChangeMailMsg := "Please change your mail addr"
	errChangeMail := errors.New(errChangeMailMsg)

	errAppended := errUserAlreadyExists.AppendErr(errChangeMail)
	expectedErrorMessage := errUserAlreadyExists.Format(userMail).Error() + "\n" + errChangeMail.Error()

	do("Append Test Error type", errAppended, expectedErrorMessage, t)
}

func TestAppend(t *testing.T) {
	errors.Prefix = "error: "

	errChangeMailMsg := "Please change your mail addr"
	expectedErrorMessage := errUserAlreadyExists.Format(userMail).Error() + "\n" + errChangeMailMsg
	errAppended := errUserAlreadyExists.Append(errChangeMailMsg)
	do("Append Test string Message", errAppended, expectedErrorMessage, t)
}

func TestEqual(t *testing.T) {
	newE := errUserAlreadyExists
	if !errUserAlreadyExists.Equal(newE) {
		t.Fatalf("error %s failed, expected:\n%s got:\n%s", "Equal", newE.Error(), errUserAlreadyExists.Error())
	}
}

func TestEmpty(t *testing.T) {
	newE := errors.New("")
	if newE.Empty() {
		t.Fatal("error Empty failed, expected:\ntrue got:\nfalse")
	}
}
func TestNotEmpty(t *testing.T) {
	newE := errors.New("123")
	if !newE.NotEmpty() {
		t.Fatal("error NotEmpty failed, expected:\nfalse got:\ntrue")
	}
}

func TestNewLine(t *testing.T) {
	err := errors.New(errMessage)
	expected := errors.Prefix + expectedUserAlreadyExists
	do("NewLine Test", err, expected, t)
}

func TestNewLineAppend(t *testing.T) {
	err := errors.New(errMessage)
	err.AppendErr(errUserAlreadyExists)
	expected := errors.Prefix + expectedUserAlreadyExists
	_ = err.HasStack()
	do("NewLine Test", err, expected, t)
}

func TestPrefix(t *testing.T) {
	errors.Prefix = "MyPrefix: "

	errUpdatedPrefix := errors.New(errMessage)
	expected := errors.Prefix + expectedUserAlreadyExists
	do("Prefix Test with "+errors.Prefix, errUpdatedPrefix, expected, t)
}

func TestWith(t *testing.T) {
	errUpdatedPrefix := errors.New(errMessage)
	errUpdatedPrefix.With(nil)
	expected := errors.Prefix + expectedUserAlreadyExists
	do("With Test", errUpdatedPrefix, expected, t)
}

func TestWith2(t *testing.T) {
	errors.Prefix = "MyPrefix: "

	userErr := errors.New(userMail)
	errUpdatedPrefix := errors.New(errMessage)
	errUpdatedPrefix.With(userErr)
	expected := errors.Prefix + expectedUserAlreadyExists
	do("With Test2", errUpdatedPrefix, expected, t)
}

func TestPanic(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()

	errUpdatedPrefix := errors.New(errMessage)
	errUpdatedPrefix.Panic()
}

func TestPanic2(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()

	errUpdatedPrefix := errors.New(errMessage)
	errUpdatedPrefix.Panicf(userMail)
}
