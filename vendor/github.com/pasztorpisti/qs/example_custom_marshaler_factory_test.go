package qs_test

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/pasztorpisti/qs"
)

// This example shows how to create QSMarshaler and QSUnmarshaler objects
// that have custom marshaler and unmarshaler factories to provide custom
// marshaling and unmarshaling for some types.
//
// In this example we change the default marshaling and unmarshaling of the
// []byte type and we compare our custom marshaler with the default one. You can
// not only change the behavior of already supported types (like []byte) but you
// can also add types that aren't supported by default.
// E.g.: time.Time, time.Duration.
//
// Builtin unnamed golang types (like []byte) can't implement the MarshalQS and
// UnmarshalQS interfaces to provide their own marshaling, this is why we have
// to create custom QSMarshaler and QSUnmarshaler with custom factories for them.
func Example_customMarshalerFactory() {
	customMarshaler := qs.NewMarshaler(&qs.MarshalOptions{
		MarshalerFactory: &marshalerFactory{qs.DefaultMarshalerFactory},
	})
	customUnmarshaler := qs.NewUnmarshaler(&qs.UnmarshalOptions{
		UnmarshalerFactory: &unmarshalerFactory{qs.DefaultUnmarshalerFactory},
	})

	performSliceTest("Default", qs.DefaultMarshaler, qs.DefaultUnmarshaler)
	performSliceTest("Custom", customMarshaler, customUnmarshaler)
	performTimeTest(customMarshaler, customUnmarshaler)

	// Output:
	// Default-Marshal-Result: a=0&a=1&a=2&b=3&b=4&b=5 <nil>
	// Default-Unmarshal-Result: len=2 a=[0 1 2] b=[3 4 5] <nil>
	// Custom-Marshal-Result: a=000102&b=030405 <nil>
	// Custom-Unmarshal-Result: len=2 a=[0 1 2] b=[3 4 5] <nil>
	// Time-Marshal-Result: time=2000-05-01T00%3A00%3A00Z <nil>
	// Time-Unmarshal-Result: len=1 time=2000-05-01T00:00:00Z <nil>
}

func performSliceTest(name string, m *qs.QSMarshaler, um *qs.QSUnmarshaler) {
	queryStr, err := m.Marshal(map[string][]byte{
		"a": {0, 1, 2},
		"b": {3, 4, 5},
	})
	fmt.Printf("%v-Marshal-Result: %v %v\n", name, queryStr, err)

	var query map[string][]byte
	err = um.Unmarshal(&query, queryStr)
	fmt.Printf("%v-Unmarshal-Result: len=%v a=%v b=%v %v\n",
		name, len(query), query["a"], query["b"], err)
}

func performTimeTest(m *qs.QSMarshaler, um *qs.QSUnmarshaler) {
	queryStr, err := m.Marshal(map[string]time.Time{
		"time": time.Date(2000, time.May, 1, 0, 0, 0, 0, time.UTC),
	})
	fmt.Printf("Time-Marshal-Result: %v %v\n", queryStr, err)

	var query map[string]time.Time
	err = um.Unmarshal(&query, queryStr)
	fmt.Printf("Time-Unmarshal-Result: len=%v time=%v %v\n",
		len(query), query["time"].Format(time.RFC3339), err)
}

var byteSliceType = reflect.TypeOf([]byte(nil))
var timeType = reflect.TypeOf((*time.Time)(nil)).Elem()

// marshalerFactory implements the MarshalerFactory interface and provides
// custom Marshaler for the []byte type.
type marshalerFactory struct {
	orig qs.MarshalerFactory
}

func (f *marshalerFactory) Marshaler(t reflect.Type, opts *qs.MarshalOptions) (qs.Marshaler, error) {
	switch t {
	case byteSliceType:
		return byteSliceMarshaler{}, nil
	case timeType:
		return timeMarshalerInstance, nil
	default:
		return f.orig.Marshaler(t, opts)
	}
}

// unmarshalerFactory implements the UnmarshalerFactory interface and provides
// custom Unmarshaler for the []byte type.
type unmarshalerFactory struct {
	orig qs.UnmarshalerFactory
}

func (f *unmarshalerFactory) Unmarshaler(t reflect.Type, opts *qs.UnmarshalOptions) (qs.Unmarshaler, error) {
	switch t {
	case byteSliceType:
		return byteSliceMarshaler{}, nil
	case timeType:
		return timeMarshalerInstance, nil
	default:
		return f.orig.Unmarshaler(t, opts)
	}
}

// byteSliceMarshaler implements the Marshaler and Unmarshaler interfaces to
// provide custom marshaling and unmarshaling for the []byte type.
type byteSliceMarshaler struct{}

func (byteSliceMarshaler) Marshal(v reflect.Value, opts *qs.MarshalOptions) ([]string, error) {
	return []string{hex.EncodeToString(v.Interface().([]byte))}, nil
}

func (byteSliceMarshaler) Unmarshal(v reflect.Value, a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(b))
	return nil
}

var timeMarshalerInstance = &timeMarshaler{
	layouts: []string{time.RFC3339},
}

// timeMarshaler implements the Marshaler and Unmarshaler interfaces to
// provide custom marshaling and unmarshaling for the time.Time type.
type timeMarshaler struct {
	layouts []string
}

func (o *timeMarshaler) Marshal(v reflect.Value, opts *qs.MarshalOptions) ([]string, error) {
	return []string{v.Interface().(time.Time).Format(o.layouts[0])}, nil
}

func (o *timeMarshaler) Unmarshal(v reflect.Value, a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	for _, layout := range o.layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			v.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return fmt.Errorf("unsupported time format: %v", s)
}
