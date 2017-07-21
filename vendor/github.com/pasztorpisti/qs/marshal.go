package qs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

// MarshalPresence is an enum that controls the marshaling of empty fields.
// A field is empty if it has its zero value or it is an empty container.
type MarshalPresence int

const (
	// MPUnspecified can be used as the value of MarshalOptions.DefaultMarshalPresence
	// to tell the NewMarshaler function to use the value of the global
	// DefaultMarshalPresence variable.
	MPUnspecified MarshalPresence = iota

	// KeepEmpty marshals the values of empty fields into the marshal output.
	KeepEmpty

	// OmitEmpty doesn't marshal the values of empty fields into the marshal output.
	OmitEmpty
)

func (v MarshalPresence) String() string {
	switch v {
	case MPUnspecified:
		return "MPUnspecified"
	case KeepEmpty:
		return "keepempty"
	case OmitEmpty:
		return "omitempty"
	default:
		return fmt.Sprintf("MarshalPresence(%v)", int(v))
	}
}

// MarshalOptions is used as a parameter by the NewMarshaler function.
type MarshalOptions struct {
	// NameTransformer is used to transform struct field names into a query
	// string names when they aren't set explicitly in the struct field tag.
	// If this field is nil then NewMarshaler uses the DefaultNameTransformer
	// global variable.
	NameTransformer NameTransformFunc

	// ValuesMarshalerFactory is used by QSMarshaler to create ValuesMarshaler
	// objects for specific types. If this field is nil then NewMarshaler uses
	// the value of the DefaultValuesMarshalerFactory global variable.
	ValuesMarshalerFactory ValuesMarshalerFactory

	// MarshalerFactory is used by QSMarshaler to create Marshaler
	// objects for specific types. If this field is nil then NewMarshaler uses
	// the value of the DefaultMarshalerFactory global variable.
	MarshalerFactory MarshalerFactory

	// DefaultMarshalPresence is used for the marshaling of struct fields that
	// don't have an explicit MarshalPresence option set in their tags.
	// This option is used for every item when you marshal a map[string]WhateverType
	// instead of a struct because map items can't have a tag to override this.
	DefaultMarshalPresence MarshalPresence
}

// DefaultValuesMarshalerFactory is used by the NewMarshaler function when its
// MarshalOptions.ValuesMarshalerFactory parameter is nil.
var DefaultValuesMarshalerFactory = newValuesMarshalerFactory()

// DefaultMarshalerFactory is used by the NewMarshaler function when its
// MarshalOptions.MarshalerFactory parameter is nil. This variable is set
// to a factory object that handles most builtin types (arrays, pointers,
// bool, int, etc...). If a type implements the MarshalQS interface then this
// factory returns an marshaler object that allows instances of the given type
// to marshal themselves.
var DefaultMarshalerFactory = newMarshalerFactory()

// DefaultMarshalPresence is used by the NewMarshaler function when its
// MarshalOptions.DefaultMarshalPresence parameter is MPUnspecified.
var DefaultMarshalPresence = KeepEmpty

// DefaultMarshaler is the marshaler used by the Marshal, MarshalValues,
// CanMarshal and CanMarshalType functions.
var DefaultMarshaler = NewMarshaler(&MarshalOptions{})

// Marshal marshals an object into a query string. The type of the object must
// be supported by the ValuesMarshalerFactory of the marshaler. By default only
// structs and maps satisfy this condition without using a custom
// ValuesMarshalerFactory.
//
// If you use a map then the key type has to be string or a type with string as
// its underlying type and the map value type can be anything that can be used
// as a struct field for marshaling.
//
// A struct value is marshaled by adding its fields one-by-one to the query
// string. Only exported struct fields are marshaled. The struct field tag can
// contain qs package specific options in the following format:
//
//  FieldName bool `qs:"[name][,option1[,option2[...]]]"`
//
//  - If name is "-" then this field is skipped just like unexported fields.
//  - If name is omitted then it defaults to the snake_case of the FieldName.
//    The snake_case transformation can be replaced with a field name to query
//    string name converter function by creating a custom marshaler.
//  - For marshaling you can specify one of the keepempty and omitempty options.
//    If none of them is specified then the keepempty option is the default but
//    this default can be changed by using a custom marshaler object.
//
//  Examples:
//  FieldName bool `qs:"-"
//  FieldName bool `qs:"name_in_query_str"
//  FieldName bool `qs:"name_in_query_str,keepempty"
//  FieldName bool `qs:",omitempty"
//
// Anonymous struct fields are marshaled as if their inner exported fields were
// fields in the outer struct.
//
// Pointer fields are omitted when they are nil otherwise they are marshaled as
// the value pointed to.
//
// Items of array and slice fields are encoded by adding multiple items with the
// same key to the query string. E.g.: arr=[]byte{1, 2} is encoded as "arr=1&arr=2".
// You can change this behavior by creating a custom marshaler with its custom
// MarshalerFactory that provides your custom marshal logic for the given slice
// and/or array types.
//
// When a field is marshaled with the omitempty option then the field is skipped
// if it has the zero value of its type.
// A field is marshaled with the omitempty option when its tag explicitly
// specifies omitempty or when the tag contains neither omitempty nor keepempty
// but the marshaler's default marshal option is omitempty.
func Marshal(i interface{}) (string, error) {
	return DefaultMarshaler.Marshal(i)
}

// MarshalValues is the same as Marshal but returns a url.Values instead of a
// query string.
func MarshalValues(i interface{}) (url.Values, error) {
	return DefaultMarshaler.MarshalValues(i)
}

// CheckMarshal returns an error if the type of the given object can't be
// marshaled into a url.Values or query string. By default only maps and structs
// can be marshaled into query strings given that all of their fields or values
// can be marshaled to []string (which is the value type of the url.Values map).
//
// It performs the check on the type of the object without traversing or
// marshaling the object.
func CheckMarshal(i interface{}) error {
	return DefaultMarshaler.CheckMarshal(i)
}

// CheckMarshalType returns an error if the given type can't be marshaled
// into a url.Values or query string. By default only maps and structs
// can be marshaled int query strings given that all of their fields or values
// can be marshaled to []string (which is the value type of the url.Values map).
func CheckMarshalType(t reflect.Type) error {
	return DefaultMarshaler.CheckMarshalType(t)
}

// QSMarshaler objects can be created by calling NewMarshaler and they can be
// used to marshal structs or maps into query strings or url.Values.
type QSMarshaler struct {
	opts *MarshalOptions
}

// NewMarshaler returns a new QSMarshaler object.
func NewMarshaler(opts *MarshalOptions) *QSMarshaler {
	return &QSMarshaler{
		opts: prepareMarshalOptions(*opts),
	}
}

// Marshal marshals a given object into a query string.
// See the documentation of the global Marshal func.
func (p *QSMarshaler) Marshal(i interface{}) (string, error) {
	values, err := p.MarshalValues(i)
	if err != nil {
		return "", err
	}
	return values.Encode(), nil
}

// MarshalValues marshals a given object into a url.Values.
// See the documentation of the global MarshalValues func.
func (p *QSMarshaler) MarshalValues(i interface{}) (url.Values, error) {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return nil, errors.New("received an empty interface")
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer of type %T", i)
		}
		v = v.Elem()
	}

	vum, err := p.opts.ValuesMarshalerFactory.ValuesMarshaler(v.Type(), p.opts)
	if err != nil {
		return nil, err
	}
	return vum.MarshalValues(v, p.opts)
}

// CheckMarshal check whether the type of the given object supports
// marshaling into query strings.
// See the documentation of the global CheckMarshal func.
func (p *QSMarshaler) CheckMarshal(i interface{}) error {
	return p.CheckMarshalType(reflect.TypeOf(i))
}

// CheckMarshalType check whether the given type supports marshaling into
// query strings. See the documentation of the global CheckMarshalType func.
func (p *QSMarshaler) CheckMarshalType(t reflect.Type) error {
	if t == nil {
		return errors.New("nil type")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	_, err := p.opts.ValuesMarshalerFactory.ValuesMarshaler(t, p.opts)
	return err
}

func prepareMarshalOptions(opts MarshalOptions) *MarshalOptions {
	if opts.NameTransformer == nil {
		opts.NameTransformer = DefaultNameTransform
	}
	if opts.ValuesMarshalerFactory == nil {
		opts.ValuesMarshalerFactory = DefaultValuesMarshalerFactory
	}
	if opts.MarshalerFactory == nil {
		opts.MarshalerFactory = DefaultMarshalerFactory
	}
	if opts.DefaultMarshalPresence == MPUnspecified {
		if DefaultMarshalPresence == MPUnspecified {
			opts.DefaultMarshalPresence = KeepEmpty
		} else {
			opts.DefaultMarshalPresence = DefaultMarshalPresence
		}
	}
	return &opts
}
