package qs

import (
	"fmt"
	"reflect"
)

// ReqError is returned when a struct field marked with the 'req' option isn't
// in the unmarshaled url.Values or query string.
type ReqError string

func (e ReqError) Error() string {
	return string(e)
}

type wrongTypeError struct {
	Actual   reflect.Type
	Expected reflect.Type
}

func (e *wrongTypeError) Error() string {
	return fmt.Sprintf("received type %v, want %v", e.Actual, e.Expected)
}

type wrongKindError struct {
	Actual   reflect.Type
	Expected reflect.Kind
}

func (e *wrongKindError) Error() string {
	return fmt.Sprintf("received type %v of kind %v, want kind %v",
		e.Actual, e.Actual.Kind(), e.Expected)
}

type unhandledTypeError struct {
	Type reflect.Type
}

func (e *unhandledTypeError) Error() string {
	return fmt.Sprintf("unhandled type: %v", e.Type)
}
