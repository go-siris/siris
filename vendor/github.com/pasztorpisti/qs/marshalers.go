package qs

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

// structMarshaler implements ValuesMarshaler.
type structMarshaler struct {
	Type           reflect.Type
	EmbeddedFields []embeddedFieldMarshaler
	Fields         []*fieldMarshaler
}

type embeddedFieldMarshaler struct {
	FieldIndex      int
	ValuesMarshaler ValuesMarshaler
}

type fieldMarshaler struct {
	FieldIndex int
	Marshaler  Marshaler
	Tag        parsedTag
}

// newStructMarshaler creates a struct marshaler for a specific struct type.
func newStructMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if t.Kind() != reflect.Struct {
		return nil, &wrongKindError{Expected: reflect.Struct, Actual: t}
	}

	sm := &structMarshaler{
		Type: t,
	}

	for i, numField := 0, t.NumField(); i < numField; i++ {
		sf := t.Field(i)
		vm, fm, err := newFieldMarshaler(sf, opts)
		if err != nil {
			return nil, fmt.Errorf("error creating marshaler for field %v of struct %v :: %v",
				sf.Name, t, err)
		}
		if vm != nil {
			sm.EmbeddedFields = append(sm.EmbeddedFields, embeddedFieldMarshaler{
				FieldIndex:      i,
				ValuesMarshaler: vm,
			})
		}
		if fm != nil {
			fm.FieldIndex = i
			sm.Fields = append(sm.Fields, fm)
		}
	}

	return sm, nil
}

func newFieldMarshaler(sf reflect.StructField, opts *MarshalOptions) (vm ValuesMarshaler, fm *fieldMarshaler, err error) {
	skip, tag, err := getStructFieldInfo(sf, opts.NameTransformer, opts.DefaultMarshalPresence, UPUnspecified)
	if skip || err != nil {
		return
	}

	t := sf.Type
	if sf.Anonymous {
		vm, err = opts.ValuesMarshalerFactory.ValuesMarshaler(t, opts)
		if err == nil {
			// We can end up here for example in case of an embedded struct.
			return
		}
	}

	m, err := opts.MarshalerFactory.Marshaler(t, opts)
	if err != nil {
		return
	}
	fm = &fieldMarshaler{
		Marshaler: m,
		Tag:       tag,
	}
	return
}

func (p *structMarshaler) MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &wrongTypeError{Actual: t, Expected: p.Type}
	}

	// TODO: use a StructError error type in the function to generate
	// error messages prefixed with the name of the struct type.

	vs := make(url.Values, len(p.Fields))

	for _, fm := range p.Fields {
		fv := v.Field(fm.FieldIndex)
		if fm.Tag.MarshalPresence == OmitEmpty && isEmpty(fv) {
			continue
		}
		a, err := fm.Marshaler.Marshal(fv, opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling url.Values entry %q :: %v", fm.Tag.Name, err)
		}
		if len(a) != 0 {
			vs[fm.Tag.Name] = a
		}
	}

	for _, ef := range p.EmbeddedFields {
		evs, err := ef.ValuesMarshaler.MarshalValues(v.Field(ef.FieldIndex), opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling embedded field %q :: %v", v.Type().Field(ef.FieldIndex).Name, err)
		}
		for k, a := range evs {
			vs[k] = a
		}
	}

	return vs, nil
}

func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0.0
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

type mapMarshaler struct {
	Type          reflect.Type
	ElemMarshaler Marshaler
}

func newMapMarshaler(t reflect.Type, opts *MarshalOptions) (ValuesMarshaler, error) {
	if t.Kind() != reflect.Map {
		return nil, &wrongKindError{Expected: reflect.Map, Actual: t}
	}

	if t.Key() != stringType {
		return nil, fmt.Errorf("map key type is expected to be string: %v", t)
	}

	et := t.Elem()
	m, err := opts.MarshalerFactory.Marshaler(et, opts)
	if err != nil {
		// TODO: use a MapError error type in the function to generate
		// error messages prefixed with the name of the struct type.
		return nil, fmt.Errorf("error getting marshaler for map value type %v :: %v", et, err)
	}

	return &mapMarshaler{
		Type:          t,
		ElemMarshaler: m,
	}, nil
}

func (p *mapMarshaler) MarshalValues(v reflect.Value, opts *MarshalOptions) (url.Values, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &wrongTypeError{Actual: t, Expected: p.Type}
	}

	vlen := v.Len()
	if vlen == 0 {
		return nil, nil
	}

	vs := make(url.Values, vlen)
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		if opts.DefaultMarshalPresence == OmitEmpty && isEmpty(val) {
			continue
		}
		keyStr := key.String()
		a, err := p.ElemMarshaler.Marshal(val, opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling key %q :: %v", keyStr, err)
		}
		vs[keyStr] = a
	}
	return vs, nil
}

type ptrMarshaler struct {
	Type          reflect.Type
	ElemMarshaler Marshaler
}

func newPtrMarshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	if t.Kind() != reflect.Ptr {
		return nil, &wrongKindError{Expected: reflect.Ptr, Actual: t}
	}
	et := t.Elem()
	em, err := opts.MarshalerFactory.Marshaler(et, opts)
	if err != nil {
		return nil, err
	}
	return &ptrMarshaler{
		Type:          t,
		ElemMarshaler: em,
	}, nil
}

func (p *ptrMarshaler) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &wrongTypeError{Actual: t, Expected: p.Type}
	}
	if v.IsNil() {
		return nil, nil
	}
	return p.ElemMarshaler.Marshal(v.Elem(), opts)
}

type arrayAndSliceMarshaler struct {
	Type          reflect.Type
	ElemMarshaler Marshaler
}

func newArrayAndSliceMarshaler(t reflect.Type, opts *MarshalOptions) (Marshaler, error) {
	k := t.Kind()
	if k != reflect.Array && k != reflect.Slice {
		return nil, &wrongKindError{Expected: reflect.Array, Actual: t}
	}

	em, err := opts.MarshalerFactory.Marshaler(t.Elem(), opts)
	if err != nil {
		return nil, err
	}
	return &arrayAndSliceMarshaler{
		Type:          t,
		ElemMarshaler: em,
	}, nil
}

func (p *arrayAndSliceMarshaler) Marshal(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	t := v.Type()
	if t != p.Type {
		return nil, &wrongTypeError{Actual: t, Expected: p.Type}
	}

	vlen := v.Len()
	if vlen == 0 {
		return nil, nil
	}

	a := make([]string, vlen)
	for i := 0; i < vlen; i++ {
		a2, err := p.ElemMarshaler.Marshal(v.Index(i), opts)
		if err != nil {
			return nil, fmt.Errorf("error marshaling array/slice index %v :: %v", i, err)
		}
		if len(a2) != 1 {
			return nil, fmt.Errorf("marshaler returned a slice of length %v for array/slice index %v", len(a2), i)
		}
		a[i] = a2[0]
	}
	return a, nil
}

func marshalString(v reflect.Value, opts *MarshalOptions) (string, error) {
	if v.Kind() != reflect.String {
		return "", &wrongKindError{Expected: reflect.String, Actual: v.Type()}
	}
	return v.String(), nil
}

func marshalBool(v reflect.Value, opts *MarshalOptions) (string, error) {
	if v.Kind() != reflect.Bool {
		return "", &wrongKindError{Expected: reflect.Struct, Actual: v.Type()}
	}
	return strconv.FormatBool(v.Bool()), nil
}

func marshalInt(v reflect.Value, opts *MarshalOptions) (string, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	default:
		return "", &wrongKindError{Expected: reflect.Int, Actual: v.Type()}
	}
}

func marshalUint(v reflect.Value, opts *MarshalOptions) (string, error) {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	default:
		return "", &wrongKindError{Expected: reflect.Uint, Actual: v.Type()}
	}
}

func marshalFloat(v reflect.Value, opts *MarshalOptions) (string, error) {
	var bitSize int

	switch v.Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	default:
		return "", &wrongKindError{Expected: reflect.Float32, Actual: v.Type()}
	}

	return strconv.FormatFloat(v.Float(), 'f', -1, bitSize), nil
}

func marshalWithMarshalQS(v reflect.Value, opts *MarshalOptions) ([]string, error) {
	marshalQS, ok := v.Interface().(MarshalQS)
	if !ok {
		return nil, fmt.Errorf("expected a type that implements MarshalQS, got %v", v.Type())
	}
	return marshalQS.MarshalQS(opts)
}
