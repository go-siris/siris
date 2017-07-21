package qs

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

// structUnmarshaler implements ValuesUnmarshaler.
type structUnmarshaler struct {
	Type           reflect.Type
	EmbeddedFields []embeddedFieldUnmarshaler
	Fields         []*fieldUnmarshaler
}

type embeddedFieldUnmarshaler struct {
	FieldIndex        int
	ValuesUnmarshaler ValuesUnmarshaler
}

type fieldUnmarshaler struct {
	FieldIndex  int
	Unmarshaler Unmarshaler
	Tag         parsedTag
}

// newStructUnmarshaler creates a struct unmarshaler for a specific struct type.
func newStructUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	if t.Kind() != reflect.Struct {
		return nil, &wrongKindError{Expected: reflect.Struct, Actual: t}
	}

	su := &structUnmarshaler{
		Type: t,
	}

	for i, numField := 0, t.NumField(); i < numField; i++ {
		sf := t.Field(i)
		vum, fum, err := newFieldUnmarshaler(sf, opts)
		if err != nil {
			return nil, fmt.Errorf("error creating unmarshaler for field %v of struct %v :: %v",
				sf.Name, t, err)
		}
		if vum != nil {
			su.EmbeddedFields = append(su.EmbeddedFields, embeddedFieldUnmarshaler{
				FieldIndex:        i,
				ValuesUnmarshaler: vum,
			})
		}
		if fum != nil {
			fum.FieldIndex = i
			su.Fields = append(su.Fields, fum)
		}
	}

	return su, nil
}

func newFieldUnmarshaler(sf reflect.StructField, opts *UnmarshalOptions) (vum ValuesUnmarshaler, fum *fieldUnmarshaler, err error) {
	skip, tag, err := getStructFieldInfo(sf, opts.NameTransformer, MPUnspecified, opts.DefaultUnmarshalPresence)
	if skip || err != nil {
		return
	}

	t := sf.Type
	if sf.Anonymous {
		vum, err = opts.ValuesUnmarshalerFactory.ValuesUnmarshaler(t, opts)
		if err == nil {
			// We can end up here for example in case of an embedded struct.
			return
		}
	}

	um, err := opts.UnmarshalerFactory.Unmarshaler(t, opts)
	if err != nil {
		return
	}
	fum = &fieldUnmarshaler{
		Unmarshaler: um,
		Tag:         tag,
	}
	return
}

func (p *structUnmarshaler) UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &wrongTypeError{Actual: t, Expected: p.Type}
	}

	// TODO: use a StructError error type in the function to generate
	// error messages prefixed with the name of the struct type.

	for _, fum := range p.Fields {
		a, ok := vs[fum.Tag.Name]
		if !ok {
			if fum.Tag.UnmarshalPresence == Req {
				return ReqError(fmt.Sprintf("missing required field %q in struct %v", fum.Tag.Name, t))
			}
			if fum.Tag.UnmarshalPresence == Nil {
				continue
			}
		}
		err := fum.Unmarshaler.Unmarshal(v.Field(fum.FieldIndex), a, opts)
		if err != nil {
			return fmt.Errorf("error unmarshaling url.Values entry %q :: %v", fum.Tag.Name, err)
		}
	}

	for _, ef := range p.EmbeddedFields {
		err := ef.ValuesUnmarshaler.UnmarshalValues(v.Field(ef.FieldIndex), vs, opts)
		if err != nil {
			if _, ok := err.(ReqError); ok {
				return ReqError(fmt.Sprintf("embedded field %q :: %v", t.Field(ef.FieldIndex).Name, err))
			}
			return fmt.Errorf("error unmarshaling embedded field %q :: %v", t.Field(ef.FieldIndex).Name, err)
		}
	}

	return nil
}

type mapUnmarshaler struct {
	Type            reflect.Type
	ElemType        reflect.Type
	ElemUnmarshaler Unmarshaler
}

func newMapUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (ValuesUnmarshaler, error) {
	if t.Kind() != reflect.Map {
		return nil, &wrongKindError{Expected: reflect.Map, Actual: t}
	}

	if t.Key() != stringType {
		return nil, fmt.Errorf("map key type is expected to be string: %v", t)
	}

	et := t.Elem()
	um, err := opts.UnmarshalerFactory.Unmarshaler(et, opts)
	if err != nil {
		// TODO: use a MapError error type in the function to generate
		// error messages prefixed with the name of the struct type.
		return nil, fmt.Errorf("error getting unmarshaler for map value type %v :: %v", et, err)
	}

	return &mapUnmarshaler{
		Type:            t,
		ElemType:        et,
		ElemUnmarshaler: um,
	}, nil
}

func (p *mapUnmarshaler) UnmarshalValues(v reflect.Value, vs url.Values, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &wrongTypeError{Actual: t, Expected: p.Type}
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(t))
	}

	for k, a := range vs {
		item := reflect.New(p.ElemType).Elem()
		err := p.ElemUnmarshaler.Unmarshal(item, a, opts)
		if err != nil {
			return fmt.Errorf("error unmarshaling key %q :: %v", k, err)
		}
		v.SetMapIndex(reflect.ValueOf(k), item)
	}

	return nil
}

type ptrUnmarshaler struct {
	Type            reflect.Type
	ElemType        reflect.Type
	ElemUnmarshaler Unmarshaler
}

func newPtrUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if t.Kind() != reflect.Ptr {
		return nil, &wrongKindError{Expected: reflect.Ptr, Actual: t}
	}
	et := t.Elem()
	eu, err := opts.UnmarshalerFactory.Unmarshaler(et, opts)
	if err != nil {
		return nil, err
	}
	return &ptrUnmarshaler{
		Type:            t,
		ElemType:        et,
		ElemUnmarshaler: eu,
	}, nil
}

func (p *ptrUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &wrongTypeError{Actual: t, Expected: p.Type}
	}
	if v.IsNil() {
		v.Set(reflect.New(p.ElemType))
	}
	return p.ElemUnmarshaler.Unmarshal(v.Elem(), a, opts)
}

type arrayUnmarshaler struct {
	Type            reflect.Type
	ElemUnmarshaler Unmarshaler
	Len             int
}

func newArrayUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if t.Kind() != reflect.Array {
		return nil, &wrongKindError{Expected: reflect.Array, Actual: t}
	}

	eu, err := opts.UnmarshalerFactory.Unmarshaler(t.Elem(), opts)
	if err != nil {
		return nil, err
	}
	return &arrayUnmarshaler{
		Type:            t,
		ElemUnmarshaler: eu,
		Len:             t.Len(),
	}, nil
}

func (p *arrayUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &wrongTypeError{Actual: t, Expected: p.Type}
	}

	if a == nil {
		return nil
	}
	if len(a) != p.Len {
		return fmt.Errorf("array length == %v, want %v", len(a), p.Len)
	}
	for i := range a {
		err := p.ElemUnmarshaler.Unmarshal(v.Index(i), a[i:i+1], opts)
		if err != nil {
			return fmt.Errorf("error unmarshaling array index %v :: %v", i, err)
		}
	}
	return nil
}

type sliceUnmarshaler struct {
	Type            reflect.Type
	ElemUnmarshaler Unmarshaler
}

func newSliceUnmarshaler(t reflect.Type, opts *UnmarshalOptions) (Unmarshaler, error) {
	if t.Kind() != reflect.Slice {
		return nil, &wrongKindError{Expected: reflect.Slice, Actual: t}
	}

	eu, err := opts.UnmarshalerFactory.Unmarshaler(t.Elem(), opts)
	if err != nil {
		return nil, err
	}
	return &sliceUnmarshaler{
		Type:            t,
		ElemUnmarshaler: eu,
	}, nil
}

func (p *sliceUnmarshaler) Unmarshal(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	t := v.Type()
	if t != p.Type {
		return &wrongTypeError{Actual: t, Expected: p.Type}
	}

	if v.IsNil() {
		v.Set(reflect.MakeSlice(t, len(a), len(a)))
	}

	for i := range a {
		err := p.ElemUnmarshaler.Unmarshal(v.Index(i), a[i:i+1], opts)
		if err != nil {
			return fmt.Errorf("error unmarshaling slice index %v :: %v", i, err)
		}
	}

	return nil
}

// unmarshalString can unmarshal an ini file entry into a value with an
// underlying type (kind) of string.
func unmarshalString(v reflect.Value, s string, opts *UnmarshalOptions) error {
	if v.Kind() != reflect.String {
		return &wrongKindError{Expected: reflect.String, Actual: v.Type()}
	}
	v.SetString(s)
	return nil
}

// unmarshalBool can unmarshal an ini file entry into a value with an
// underlying type (kind) of bool.
func unmarshalBool(v reflect.Value, s string, opts *UnmarshalOptions) error {
	if v.Kind() != reflect.Bool {
		return &wrongKindError{Expected: reflect.Struct, Actual: v.Type()}
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	v.SetBool(b)
	return nil
}

// unmarshalInt can unmarshal an ini file entry into a signed integer value
// with an underlying type (kind) of int, int8, int16, int32 or int64.
func unmarshalInt(v reflect.Value, s string, opts *UnmarshalOptions) error {
	var bitSize int

	switch v.Kind() {
	case reflect.Int:
	case reflect.Int8:
		bitSize = 8
	case reflect.Int16:
		bitSize = 16
	case reflect.Int32:
		bitSize = 32
	case reflect.Int64:
		bitSize = 64
	default:
		return &wrongKindError{Expected: reflect.Int, Actual: v.Type()}
	}

	i, err := strconv.ParseInt(s, 0, bitSize)
	if err != nil {
		return err
	}

	v.SetInt(i)
	return nil
}

// unmarshalUint can unmarshal an ini file entry into an unsigned integer value
// with an underlying type (kind) of uint, uint8, uint16, uint32 or uint64.
func unmarshalUint(v reflect.Value, s string, opts *UnmarshalOptions) error {
	var bitSize int

	switch v.Kind() {
	case reflect.Uint:
	case reflect.Uint8:
		bitSize = 8
	case reflect.Uint16:
		bitSize = 16
	case reflect.Uint32:
		bitSize = 32
	case reflect.Uint64:
		bitSize = 64
	default:
		return &wrongKindError{Expected: reflect.Uint, Actual: v.Type()}
	}

	i, err := strconv.ParseUint(s, 0, bitSize)
	if err != nil {
		return err
	}

	v.SetUint(i)
	return nil
}

func unmarshalFloat(v reflect.Value, s string, opts *UnmarshalOptions) error {
	var bitSize int

	switch v.Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	default:
		return &wrongKindError{Expected: reflect.Float32, Actual: v.Type()}
	}

	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return err
	}

	v.SetFloat(f)
	return nil
}

func unmarshalWithUnmarshalQS(v reflect.Value, a []string, opts *UnmarshalOptions) error {
	if !v.CanAddr() {
		return fmt.Errorf("expected and addressable value, got %v", v)
	}
	unmarshalQS, ok := v.Addr().Interface().(UnmarshalQS)
	if !ok {
		return fmt.Errorf("expected a type that implements UnmarshalQS, got %v", v.Type())
	}
	return unmarshalQS.UnmarshalQS(a, opts)
}
