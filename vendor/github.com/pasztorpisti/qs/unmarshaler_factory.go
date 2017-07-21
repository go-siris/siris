package qs

import (
	"net/url"
	"reflect"
)

// ValuesUnmarshaler can unmarshal a url.Values into a value.
type ValuesUnmarshaler interface {
	// UnmarshalValues unmarshals the given url.Values using opts into v.
	UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalOptions) error
}

// ValuesUnmarshalerFactory can create ValuesUnmarshaler objects for various types.
type ValuesUnmarshalerFactory interface {
	// ValuesUnmarshaler returns a ValuesUnmarshaler object for the given t
	// type and opts options.
	ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error)
}

// Unmarshaler can unmarshal a []string (which is the value type of the
// url.Values map) into a reflect.Value.
type Unmarshaler interface {
	// Unmarshal unmarshals the given []string using opts into v.
	//
	// If the query string doesn't contain a key for this field then Unmarshal
	// is called only if the UnmarshalPresence option of the field is Opt
	// and in that case a == nil. In such cases pointer like types (pointers,
	// arrays, maps) should initialise nil pointers with an empty object.
	// With Nil or Req options this Unmarshal method isn't called.
	//
	// The []string is the value type of the url.Values map. If your unmarshaler
	// expects only a single string value instead of an array then you can call
	// opts.SliceToString(a).
	Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error
}

// UnmarshalerFactory can create Unmarshaler objects for various types.
type UnmarshalerFactory interface {
	// Unmarshaler returns an Unmarshaler object for the given t type and opts
	// options.
	Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error)
}

// UnmarshalQS is an interface that can be implemented by any type that
// wants to handle its own unmarshaling instead of relying on the default
// unmarshaling provided by this package.
type UnmarshalQS interface {
	// UnmarshalQS is essentially the same as the Unmarshaler.Unmarshal
	// method without its v parameter.
	UnmarshalQS(a []string, opts *UnmarshalOptions) error
}

func newValuesUnmarshalerFactory() ValuesUnmarshalerFactory {
	return &valuesUnmarshalerFactory{
		KindSubRegistries: map[reflect.Kind]ValuesUnmarshalerFactory{
			reflect.Struct: valuesUnmarshalerFactoryFunc(newStructUnmarshaler),
			reflect.Map:    valuesUnmarshalerFactoryFunc(newMapUnmarshaler),
		},
	}
}

func newUnmarshalerFactory() UnmarshalerFactory {
	return &unmarshalerFactory{
		KindSubRegistries: map[reflect.Kind]UnmarshalerFactory{
			reflect.Ptr:   unmarshalerFactoryFunc(newPtrUnmarshaler),
			reflect.Array: unmarshalerFactoryFunc(newArrayUnmarshaler),
			reflect.Slice: unmarshalerFactoryFunc(newSliceUnmarshaler),
		},
		Kinds: map[reflect.Kind]Unmarshaler{
			reflect.String: primitiveUnmarshalerFunc(unmarshalString),
			reflect.Bool:   primitiveUnmarshalerFunc(unmarshalBool),

			reflect.Int:   primitiveUnmarshalerFunc(unmarshalInt),
			reflect.Int8:  primitiveUnmarshalerFunc(unmarshalInt),
			reflect.Int16: primitiveUnmarshalerFunc(unmarshalInt),
			reflect.Int32: primitiveUnmarshalerFunc(unmarshalInt),
			reflect.Int64: primitiveUnmarshalerFunc(unmarshalInt),

			reflect.Uint:   primitiveUnmarshalerFunc(unmarshalUint),
			reflect.Uint8:  primitiveUnmarshalerFunc(unmarshalUint),
			reflect.Uint16: primitiveUnmarshalerFunc(unmarshalUint),
			reflect.Uint32: primitiveUnmarshalerFunc(unmarshalUint),
			reflect.Uint64: primitiveUnmarshalerFunc(unmarshalUint),

			reflect.Float32: primitiveUnmarshalerFunc(unmarshalFloat),
			reflect.Float64: primitiveUnmarshalerFunc(unmarshalFloat),
		},
	}
}

type valuesUnmarshalerFactory struct {
	KindSubRegistries map[reflect.Kind]ValuesUnmarshalerFactory
}

func (p *valuesUnmarshalerFactory) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	if subFactory, ok := p.KindSubRegistries[t.Kind()]; ok {
		return subFactory.ValuesUnmarshaler(t, opts)
	}

	return nil, &unhandledTypeError{Type: t}
}

// unmarshalerFactory implements the UnmarshalerFactory interface.
type unmarshalerFactory struct {
	KindSubRegistries map[reflect.Kind]UnmarshalerFactory
	Kinds             map[reflect.Kind]Unmarshaler
}

var unmarshalQSInterfaceType = reflect.TypeOf((*UnmarshalQS)(nil)).Elem()

func (p *unmarshalerFactory) Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if reflect.PtrTo(t).Implements(unmarshalQSInterfaceType) {
		return unmarshalerFunc(unmarshalWithUnmarshalQS), nil
	}

	k := t.Kind()
	if subFactory, ok := p.KindSubRegistries[k]; ok {
		return subFactory.Unmarshaler(t, opts)
	}
	if unmarshaler, ok := p.Kinds[k]; ok {
		return unmarshaler, nil
	}

	return nil, &unhandledTypeError{Type: t}
}

// valuesUnmarshalerFactoryFunc implements the UnmarshalerFactory interface.
type valuesUnmarshalerFactoryFunc func(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error)

func (f valuesUnmarshalerFactoryFunc) ValuesUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	return f(t, opts)
}

// unmarshalerFactoryFunc implements the UnmarshalerFactory interface.
type unmarshalerFactoryFunc func(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error)

func (f unmarshalerFactoryFunc) Unmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	return f(t, opts)
}

// unmarshalerFunc implements the Unmarshaler interface.
type unmarshalerFunc func(v reflect.Value, a []string, opts *UnmarshalOptions) error

func (f unmarshalerFunc) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	return f(v, a, opts)
}

// primitiveUnmarshalerFunc implements the Unmarshaler interface.
type primitiveUnmarshalerFunc func(v reflect.Value, s string, opts *UnmarshalOptions) error

func (f primitiveUnmarshalerFunc) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	if a == nil {
		return nil
	}
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	return f(v, s, opts)
}
