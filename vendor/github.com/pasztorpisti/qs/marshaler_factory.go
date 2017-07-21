package qs

import (
	"net/url"
	"reflect"
)

// ValuesMarshaler can marshal a value into a url.Values.
type ValuesMarshaler interface {
	// MarshalValues marshals the given v value using opts into a url.Values.
	MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error)
}

// ValuesMarshalerFactory can create ValuesMarshaler objects for various types.
type ValuesMarshalerFactory interface {
	// ValuesMarshaler returns a ValuesMarshaler object for the given t type and
	// opts options.
	ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error)
}

// Marshaler can marshal a reflect.Value into a []string.
type Marshaler interface {
	// Marshal marshals the given v value using opts into a []string.
	// Note that []string is the value type of the standard url.Values which is
	// a map[string][]string.
	Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error)
}

// MarshalerFactory can create Marshaler objects for various types.
type MarshalerFactory interface {
	// Marshaler returns a Marshaler object for the given t type and opts
	// options.
	Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error)
}

// MarshalQS is an interface that can be implemented by any type that
// wants to handle its own marshaling instead of relying on the default
// marshaling provided by this package.
type MarshalQS interface {
	// MarshalQS is essentially the same as the Marshaler.Marshal
	// method without its v parameter.
	MarshalQS(opts *MarshalOptions) ([]string, error)
}

func newValuesMarshalerFactory() ValuesMarshalerFactory {
	return &valuesMarshalerFactory{
		KindSubRegistries: map[reflect.Kind]ValuesMarshalerFactory{
			reflect.Struct: valuesMarshalerFactoryFunc(newStructMarshaler),
			reflect.Map:    valuesMarshalerFactoryFunc(newMapMarshaler),
		},
	}
}

func newMarshalerFactory() MarshalerFactory {
	return &marshalerFactory{
		KindSubRegistries: map[reflect.Kind]MarshalerFactory{
			reflect.Ptr:   marshalerFactoryFunc(newPtrMarshaler),
			reflect.Array: marshalerFactoryFunc(newArrayAndSliceMarshaler),
			reflect.Slice: marshalerFactoryFunc(newArrayAndSliceMarshaler),
		},
		Kinds: map[reflect.Kind]Marshaler{
			reflect.String: primitiveMarshalerFunc(marshalString),
			reflect.Bool:   primitiveMarshalerFunc(marshalBool),

			reflect.Int:   primitiveMarshalerFunc(marshalInt),
			reflect.Int8:  primitiveMarshalerFunc(marshalInt),
			reflect.Int16: primitiveMarshalerFunc(marshalInt),
			reflect.Int32: primitiveMarshalerFunc(marshalInt),
			reflect.Int64: primitiveMarshalerFunc(marshalInt),

			reflect.Uint:   primitiveMarshalerFunc(marshalUint),
			reflect.Uint8:  primitiveMarshalerFunc(marshalUint),
			reflect.Uint16: primitiveMarshalerFunc(marshalUint),
			reflect.Uint32: primitiveMarshalerFunc(marshalUint),
			reflect.Uint64: primitiveMarshalerFunc(marshalUint),

			reflect.Float32: primitiveMarshalerFunc(marshalFloat),
			reflect.Float64: primitiveMarshalerFunc(marshalFloat),
		},
	}
}

// valuesMarshalerFactory implements the ValuesMarshalerFactory interface.
type valuesMarshalerFactory struct {
	KindSubRegistries map[reflect.Kind]ValuesMarshalerFactory
}

func (p *valuesMarshalerFactory) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if subFactory, ok := p.KindSubRegistries[t.Kind()]; ok {
		return subFactory.ValuesMarshaler(t, opts)
	}

	return nil, &unhandledTypeError{Type: t}
}

// marshalerFactory implements the MarshalerFactory interface.
type marshalerFactory struct {
	Types             map[reflect.Type]Marshaler
	KindSubRegistries map[reflect.Kind]MarshalerFactory
	Kinds             map[reflect.Kind]Marshaler
}

var marshalQSInterfaceType = reflect.TypeOf((*MarshalQS)(nil)).Elem()

func (p *marshalerFactory) Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	if marshaler, ok := p.Types[t]; ok {
		return marshaler, nil
	}

	if t.Implements(marshalQSInterfaceType) {
		return marshalerFunc(marshalWithMarshalQS), nil
	}

	k := t.Kind()
	if subFactory, ok := p.KindSubRegistries[k]; ok {
		return subFactory.Marshaler(t, opts)
	}
	if marshaler, ok := p.Kinds[k]; ok {
		return marshaler, nil
	}

	return nil, &unhandledTypeError{Type: t}
}

// valuesMarshalerFactoryFunc implements the ValuesMarshalerFactory interface.
type valuesMarshalerFactoryFunc func(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error)

func (f valuesMarshalerFactoryFunc) ValuesMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	return f(t, opts)
}

// marshalerFactoryFunc implements the MarshalerFactory interface.
type marshalerFactoryFunc func(t reflect.Type, opts *MarshalOptions) (Marshaler, error)

func (f marshalerFactoryFunc) Marshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	return f(t, opts)
}

// marshalerFunc implements the Marshaler interface.
type marshalerFunc func(v reflect.Value, opts *MarshalOptions) ([]string, error)

func (f marshalerFunc) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	return f(v, opts)
}

// primitiveMarshalerFunc implements the Marshaler interface.
type primitiveMarshalerFunc func(v reflect.Value, opts *MarshalOptions) (string, error)

func (f primitiveMarshalerFunc) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	s, err := f(v, opts)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}
