package qs

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// UQSBytes implements the MarshalQS interface.
// The "U" prefix stands for Unmarshal to avoid name collisions with the
// Marshal tests.
type UQSBytes []byte

func (p *UQSBytes) UnmarshalQS(a []string, opts *UnmarshalOptions) error {
	// a is nil when the query string doesn't contain any items for this
	// field and the UnmarshalPresence of this field is Opt.
	if a == nil {
		*p = []byte{}
		return nil
	}
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*p = b
	return nil
}

type UEmbedded2 struct {
	EI int
}

type UEmbedded struct {
	UEmbedded2
}

// UTypes is used by the TestMarshalTypes test to check the marshaling of all
// supported types.
type UTypes struct {
	S  string
	B  bool
	B2 bool

	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64

	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64

	F32 float32
	F64 float64

	Ptr    *int
	Ptr2   *int
	Array  [2]int
	Slice  []int
	Slice2 []int
	QS     UQSBytes
	QS2    UQSBytes

	UEmbedded
}

// UUnspecified is a struct that defines the UnmarshalPresence tag of its fields
// as UPUnspecified.
type UUnspecified struct {
	S string
	B bool

	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64

	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64

	F32 float32
	F64 float64

	Ptr   *int
	Array [2]int
	Slice []int
	QS    UQSBytes

	UEmbedded
}

// UNil is a struct that defines the UnmarshalPresence tag of its fields as Nil.
type UNil struct {
	S string `qs:",nil"`
	B bool   `qs:",nil"`

	I   int   `qs:",nil"`
	I8  int8  `qs:",nil"`
	I16 int16 `qs:",nil"`
	I32 int32 `qs:",nil"`
	I64 int64 `qs:",nil"`

	U   uint   `qs:",nil"`
	U8  uint8  `qs:",nil"`
	U16 uint16 `qs:",nil"`
	U32 uint32 `qs:",nil"`
	U64 uint64 `qs:",nil"`

	F32 float32 `qs:",nil"`
	F64 float64 `qs:",nil"`

	Ptr *int `qs:",nil"`
	// Array: nil should have no effect
	Array [2]int   `qs:",nil"`
	Slice []int    `qs:",nil"`
	QS    UQSBytes `qs:",nil"`

	// UEmbedded: nil should have no effect
	UEmbedded `qs:",nil"`
}

// UOpt is a struct that defines the UnmarshalPresence tag of its fields as Opt.
type UOpt struct {
	S string `qs:",opt"`
	B bool   `qs:",opt"`

	I   int   `qs:",opt"`
	I8  int8  `qs:",opt"`
	I16 int16 `qs:",opt"`
	I32 int32 `qs:",opt"`
	I64 int64 `qs:",opt"`

	U   uint   `qs:",opt"`
	U8  uint8  `qs:",opt"`
	U16 uint16 `qs:",opt"`
	U32 uint32 `qs:",opt"`
	U64 uint64 `qs:",opt"`

	F32 float32 `qs:",opt"`
	F64 float64 `qs:",opt"`

	Ptr   *int     `qs:",opt"`
	Array [2]int   `qs:",opt"`
	Slice []int    `qs:",opt"`
	QS    UQSBytes `qs:",opt"`

	// UEmbedded: opt should have no effect
	UEmbedded `qs:",opt"`
}

// UReq is a struct that defines the UnmarshalPresence tag of its fields as Req.
type UReq struct {
	S string `qs:",req"`
	B bool   `qs:",req"`

	I   int   `qs:",req"`
	I8  int8  `qs:",req"`
	I16 int16 `qs:",req"`
	I32 int32 `qs:",req"`
	I64 int64 `qs:",req"`

	U   uint   `qs:",req"`
	U8  uint8  `qs:",req"`
	U16 uint16 `qs:",req"`
	U32 uint32 `qs:",req"`
	U64 uint64 `qs:",req"`

	F32 float32 `qs:",req"`
	F64 float64 `qs:",req"`

	Ptr   *int     `qs:",req"`
	Array [2]int   `qs:",req"`
	Slice []int    `qs:",req"`
	QS    UQSBytes `qs:",req"`

	// UEmbedded: opt should have no effect
	UEmbedded `qs:",req"`
}

// comparisonResults is a utility class to gather the results of multiple
// comparison for reporting the differences after finishing the comparisons.
type comparisonResults struct {
	errors []string
}

func (p *comparisonResults) finish() error {
	if len(p.errors) == 0 {
		return nil
	}
	return errors.New(strings.Join(p.errors, "\n"))
}

func (p *comparisonResults) compare(name string, value, want interface{}) {
	if !compareValues(value, want) {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			value = v.Elem().Interface()
		}
		p.errors = append(p.errors, fmt.Sprintf("%v == %#v, want %#v", name, value, want))
	}
}

func compareValues(value, want interface{}) bool {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return false
	}

	w := reflect.ValueOf(want)
	switch v.Kind() {
	case reflect.Array:
		if !w.IsValid() || w.IsNil() {
			panic("array can't be compared to nil")
		}
		if v.Len() != w.Len() {
			return false
		}
		for i, vlen := 0, v.Len(); i < vlen; i++ {
			if v.Index(i).Interface() != w.Index(i).Interface() {
				return false
			}
		}
		return true
	case reflect.Slice:
		if !w.IsValid() || w.IsNil() {
			return v.IsNil()
		}
		if v.IsNil() {
			return false
		}
		if v.Len() != w.Len() {
			return false
		}
		for i, vlen := 0, v.Len(); i < vlen; i++ {
			if v.Index(i).Interface() != w.Index(i).Interface() {
				return false
			}
		}
		return true
	case reflect.Ptr:
		if v.IsNil() {
			return !w.IsValid() || value == want
		}
		v = v.Elem()
		return v.Interface() == want
	default:
		return w.IsValid() && value == want
	}
}

func TestUnmarshalTypes(t *testing.T) {
	queryString := strings.Join([]string{
		"s=str",
		"b=true&b2=false",
		"i=-1&i8=-8&i16=-16&i32=-32&i64=-64",
		"u=1&u8=8&u16=16&u32=32&u64=64",
		"f32=32.32&f64=64.64",
		"ptr=42",
		"array=1&array=2",
		"slice=3&slice=4",
		"qs=010203",
		"ei=33",
	}, "&")

	var us UTypes
	err := Unmarshal(&us, queryString)
	if err != nil {
		t.Error(err)
	} else {
		var cr comparisonResults
		cr.compare("s", us.S, "str")
		cr.compare("b", us.B, true)
		cr.compare("b2", us.B2, false)
		cr.compare("i", us.I, -1)
		cr.compare("i8", us.I8, int8(-8))
		cr.compare("i16", us.I16, int16(-16))
		cr.compare("i32", us.I32, int32(-32))
		cr.compare("i64", us.I64, int64(-64))
		cr.compare("u", us.U, uint(1))
		cr.compare("u8", us.U8, uint8(8))
		cr.compare("u16", us.U16, uint16(16))
		cr.compare("u32", us.U32, uint32(32))
		cr.compare("u64", us.U64, uint64(64))
		cr.compare("f32", us.F32, float32(32.32))
		cr.compare("f64", us.F64, 64.64)
		cr.compare("ptr", us.Ptr, 42)
		cr.compare("ptr2", us.Ptr2, 0)
		cr.compare("array", us.Array, []int{1, 2})
		cr.compare("slice", us.Slice, []int{3, 4})
		cr.compare("slice2", us.Slice2, []int{})
		cr.compare("qs", us.QS, []byte{1, 2, 3})
		cr.compare("qs2", us.QS2, []int{})
		cr.compare("ei", us.EI, 33)
		if err := cr.finish(); err != nil {
			t.Error(err)
		}
	}
}

func TestDefaultOpt(t *testing.T) {
	queryString := strings.Join([]string{
		"s=str",
		"b=true",
		"i=-1&i8=-8&i16=-16&i32=-32&i64=-64",
		"u=1&u8=8&u16=16&u32=32&u64=64",
		"f32=32.32&f64=64.64",
		"ptr=42",
		"array=1&array=2",
		"slice=3&slice=4",
		"qs=010203",
		"ei=33",
	}, "&")

	// default presence: opt, struct presence: unspecified, queryString: nozero
	{
		var us UUnspecified
		err := Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: unspecified, queryString: zero
	{
		var us UUnspecified
		err := Unmarshal(&us, "")
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "")
			cr.compare("b", us.B, false)
			cr.compare("i", us.I, 0)
			cr.compare("i8", us.I8, int8(0))
			cr.compare("i16", us.I16, int16(0))
			cr.compare("i32", us.I32, int32(0))
			cr.compare("i64", us.I64, int64(0))
			cr.compare("u", us.U, uint(0))
			cr.compare("u8", us.U8, uint8(0))
			cr.compare("u16", us.U16, uint16(0))
			cr.compare("u32", us.U32, uint32(0))
			cr.compare("u64", us.U64, uint64(0))
			cr.compare("f32", us.F32, float32(0.0))
			cr.compare("f64", us.F64, 0.0)
			cr.compare("ptr", us.Ptr, 0)
			cr.compare("array", us.Array, []int{0, 0})
			cr.compare("slice", us.Slice, []int{})
			cr.compare("qs", us.QS, []byte{})
			cr.compare("ei", us.EI, 0)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: opt, queryString: nozero
	{
		var us UOpt
		err := Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: opt, queryString: zero
	{
		var us UOpt
		err := Unmarshal(&us, "")
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "")
			cr.compare("b", us.B, false)
			cr.compare("i", us.I, 0)
			cr.compare("i8", us.I8, int8(0))
			cr.compare("i16", us.I16, int16(0))
			cr.compare("i32", us.I32, int32(0))
			cr.compare("i64", us.I64, int64(0))
			cr.compare("u", us.U, uint(0))
			cr.compare("u8", us.U8, uint8(0))
			cr.compare("u16", us.U16, uint16(0))
			cr.compare("u32", us.U32, uint32(0))
			cr.compare("u64", us.U64, uint64(0))
			cr.compare("f32", us.F32, float32(0.0))
			cr.compare("f64", us.F64, 0.0)
			cr.compare("ptr", us.Ptr, 0)
			cr.compare("array", us.Array, []int{0, 0})
			cr.compare("slice", us.Slice, []int{})
			cr.compare("qs", us.QS, []byte{})
			cr.compare("ei", us.EI, 0)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: nil, queryString: nozero
	{
		var us UNil
		err := Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: nil, queryString: zero
	{
		var us UNil
		err := Unmarshal(&us, "")
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "")
			cr.compare("b", us.B, false)
			cr.compare("i", us.I, 0)
			cr.compare("i8", us.I8, int8(0))
			cr.compare("i16", us.I16, int16(0))
			cr.compare("i32", us.I32, int32(0))
			cr.compare("i64", us.I64, int64(0))
			cr.compare("u", us.U, uint(0))
			cr.compare("u8", us.U8, uint8(0))
			cr.compare("u16", us.U16, uint16(0))
			cr.compare("u32", us.U32, uint32(0))
			cr.compare("u64", us.U64, uint64(0))
			cr.compare("f32", us.F32, float32(0.0))
			cr.compare("f64", us.F64, 0.0)
			cr.compare("ptr", us.Ptr, nil)
			cr.compare("array", us.Array, []int{0, 0})
			cr.compare("slice", us.Slice, nil)
			cr.compare("qs", us.QS, nil)
			cr.compare("ei", us.EI, 0)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: req, queryString: nozero
	{
		var us UReq
		err := Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: opt, struct presence: req, queryString: zero
	{
		var us UReq
		err := Unmarshal(&us, "")
		if err == nil {
			t.Error("unexpected success")
		} else if _, ok := err.(ReqError); !ok {
			t.Errorf("expected a ReqError :: %v", err)
		}
	}
}

func TestDefaultNil(t *testing.T) {
	queryString := strings.Join([]string{
		"s=str",
		"b=true",
		"i=-1&i8=-8&i16=-16&i32=-32&i64=-64",
		"u=1&u8=8&u16=16&u32=32&u64=64",
		"f32=32.32&f64=64.64",
		"ptr=42",
		"array=1&array=2",
		"slice=3&slice=4",
		"qs=010203",
		"ei=33",
	}, "&")

	unmarshaler := NewUnmarshaler(&UnmarshalOptions{
		DefaultUnmarshalPresence: Nil,
	})

	// default presence: nil, struct presence: unspecified, queryString: nozero
	{
		var us UUnspecified
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: unspecified, queryString: zero
	{
		var us UUnspecified
		err := unmarshaler.Unmarshal(&us, "")
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "")
			cr.compare("b", us.B, false)
			cr.compare("i", us.I, 0)
			cr.compare("i8", us.I8, int8(0))
			cr.compare("i16", us.I16, int16(0))
			cr.compare("i32", us.I32, int32(0))
			cr.compare("i64", us.I64, int64(0))
			cr.compare("u", us.U, uint(0))
			cr.compare("u8", us.U8, uint8(0))
			cr.compare("u16", us.U16, uint16(0))
			cr.compare("u32", us.U32, uint32(0))
			cr.compare("u64", us.U64, uint64(0))
			cr.compare("f32", us.F32, float32(0.0))
			cr.compare("f64", us.F64, 0.0)
			cr.compare("ptr", us.Ptr, nil)
			cr.compare("array", us.Array, []int{0, 0})
			cr.compare("slice", us.Slice, nil)
			cr.compare("qs", us.QS, nil)
			cr.compare("ei", us.EI, 0)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: opt, queryString: nozero
	{
		var us UOpt
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: opt, queryString: zero
	{
		var us UOpt
		err := unmarshaler.Unmarshal(&us, "")
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "")
			cr.compare("b", us.B, false)
			cr.compare("i", us.I, 0)
			cr.compare("i8", us.I8, int8(0))
			cr.compare("i16", us.I16, int16(0))
			cr.compare("i32", us.I32, int32(0))
			cr.compare("i64", us.I64, int64(0))
			cr.compare("u", us.U, uint(0))
			cr.compare("u8", us.U8, uint8(0))
			cr.compare("u16", us.U16, uint16(0))
			cr.compare("u32", us.U32, uint32(0))
			cr.compare("u64", us.U64, uint64(0))
			cr.compare("f32", us.F32, float32(0.0))
			cr.compare("f64", us.F64, 0.0)
			cr.compare("ptr", us.Ptr, 0)
			cr.compare("array", us.Array, []int{0, 0})
			cr.compare("slice", us.Slice, []int{})
			cr.compare("qs", us.QS, []byte{})
			cr.compare("ei", us.EI, 0)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: nil, queryString: nozero
	{
		var us UNil
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: nil, queryString: zero
	{
		var us UNil
		err := unmarshaler.Unmarshal(&us, "")
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "")
			cr.compare("b", us.B, false)
			cr.compare("i", us.I, 0)
			cr.compare("i8", us.I8, int8(0))
			cr.compare("i16", us.I16, int16(0))
			cr.compare("i32", us.I32, int32(0))
			cr.compare("i64", us.I64, int64(0))
			cr.compare("u", us.U, uint(0))
			cr.compare("u8", us.U8, uint8(0))
			cr.compare("u16", us.U16, uint16(0))
			cr.compare("u32", us.U32, uint32(0))
			cr.compare("u64", us.U64, uint64(0))
			cr.compare("f32", us.F32, float32(0.0))
			cr.compare("f64", us.F64, 0.0)
			cr.compare("ptr", us.Ptr, nil)
			cr.compare("array", us.Array, []int{0, 0})
			cr.compare("slice", us.Slice, nil)
			cr.compare("qs", us.QS, nil)
			cr.compare("ei", us.EI, 0)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: req, queryString: nozero
	{
		var us UReq
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: nil, struct presence: req, queryString: zero
	{
		var us UReq
		err := unmarshaler.Unmarshal(&us, "")
		if err == nil {
			t.Error("unexpected success")
		} else if _, ok := err.(ReqError); !ok {
			t.Errorf("expected a ReqError :: %v", err)
		}
	}
}

func TestDefaultReq(t *testing.T) {
	queryString := strings.Join([]string{
		"s=str",
		"b=true",
		"i=-1&i8=-8&i16=-16&i32=-32&i64=-64",
		"u=1&u8=8&u16=16&u32=32&u64=64",
		"f32=32.32&f64=64.64",
		"ptr=42",
		"array=1&array=2",
		"slice=3&slice=4",
		"qs=010203",
		"ei=33",
	}, "&")

	unmarshaler := NewUnmarshaler(&UnmarshalOptions{
		DefaultUnmarshalPresence: Req,
	})

	// default presence: req, struct presence: unspecified, queryString: nozero
	{
		var us UUnspecified
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: req, struct presence: unspecified, queryString: zero
	{
		var us UUnspecified
		err := unmarshaler.Unmarshal(&us, "")
		if err == nil {
			t.Error("unexpected success")
		} else if _, ok := err.(ReqError); !ok {
			t.Errorf("expected a ReqError :: %v", err)
		}
	}

	// default presence: req, struct presence: opt, queryString: nozero
	{
		var us UOpt
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: req, struct presence: opt, queryString: zero
	{
		var us UOpt
		err := unmarshaler.Unmarshal(&us, "")
		if err == nil {
			t.Error("unexpected success")
		} else if _, ok := err.(ReqError); !ok {
			t.Errorf("expected a ReqError :: %v", err)
		}
	}

	// default presence: req, struct presence: nil, queryString: nozero
	{
		var us UNil
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: req, struct presence: nil, queryString: zero
	{
		var us UNil
		err := unmarshaler.Unmarshal(&us, "")
		if err == nil {
			t.Error("unexpected success")
		} else if _, ok := err.(ReqError); !ok {
			t.Errorf("expected a ReqError :: %v", err)
		}
	}

	// default presence: req, struct presence: req, queryString: nozero
	{
		var us UReq
		err := unmarshaler.Unmarshal(&us, queryString)
		if err != nil {
			t.Error(err)
		} else {
			var cr comparisonResults
			cr.compare("s", us.S, "str")
			cr.compare("b", us.B, true)
			cr.compare("i", us.I, -1)
			cr.compare("i8", us.I8, int8(-8))
			cr.compare("i16", us.I16, int16(-16))
			cr.compare("i32", us.I32, int32(-32))
			cr.compare("i64", us.I64, int64(-64))
			cr.compare("u", us.U, uint(1))
			cr.compare("u8", us.U8, uint8(8))
			cr.compare("u16", us.U16, uint16(16))
			cr.compare("u32", us.U32, uint32(32))
			cr.compare("u64", us.U64, uint64(64))
			cr.compare("f32", us.F32, float32(32.32))
			cr.compare("f64", us.F64, 64.64)
			cr.compare("ptr", us.Ptr, 42)
			cr.compare("array", us.Array, []int{1, 2})
			cr.compare("slice", us.Slice, []int{3, 4})
			cr.compare("qs", us.QS, []byte{1, 2, 3})
			cr.compare("ei", us.EI, 33)
			if err := cr.finish(); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: req, struct presence: req, queryString: zero
	{
		var us UReq
		err := unmarshaler.Unmarshal(&us, "")
		if err == nil {
			t.Error("unexpected success")
		} else if _, ok := err.(ReqError); !ok {
			t.Errorf("expected a ReqError :: %v", err)
		}
	}
}

func TestUnmarshalMap(t *testing.T) {
	// Req should be ingored and shouldn't be a problem in case of map unmarshaling.
	unmarshaler := NewUnmarshaler(&UnmarshalOptions{
		DefaultUnmarshalPresence: Req,
	})

	{
		// m is nil, Unmarshal will have to create a new map
		var m map[string]UQSBytes
		err := unmarshaler.Unmarshal(&m, "a=000102&b=030405")
		if err != nil {
			t.Error(err)
		} else if len(m) != 2 {
			t.Errorf("map should have 2 keys - 'a' and 'b': %v", m)
		} else {
			want := UQSBytes{0, 1, 2}
			if a, ok := m["a"]; !ok {
				t.Errorf("'a' is missing: %v", m)
			} else if !compareValues(a, want) {
				t.Errorf("a == %#v, want %#v", a, want)
			}

			want = UQSBytes{3, 4, 5}
			if b, ok := m["b"]; !ok {
				t.Errorf("'b' is missing: %v", m)
			} else if !compareValues(b, want) {
				t.Errorf("b == %#v, want %#v", b, want)
			}
		}
	}

	{
		// m isn't nil, Unmarshal shouldn't create the map and it should
		// overwrite "a" but should leave "x" untouched.
		m := map[string]UQSBytes{
			"a": {9},
			"x": {9},
		}
		// here the query string contains an extra "&c="
		err := unmarshaler.Unmarshal(&m, "a=000102&b=030405&c=")
		if err != nil {
			t.Error(err)
		} else if len(m) != 4 {
			t.Errorf("map should have 2 keys - 'a', 'b' and 'c': %v", m)
		} else {
			want := UQSBytes{0, 1, 2}
			if a, ok := m["a"]; !ok {
				t.Errorf("'a' is missing: %v", m)
			} else if !compareValues(a, want) {
				t.Errorf("a == %#v, want %#v", a, want)
			}

			want = UQSBytes{3, 4, 5}
			if b, ok := m["b"]; !ok {
				t.Errorf("'b' is missing: %v", m)
			} else if !compareValues(b, want) {
				t.Errorf("b == %#v, want %#v", b, want)
			}

			want = UQSBytes{}
			if c, ok := m["c"]; !ok {
				t.Errorf("'c' is missing: %v", m)
			} else if !compareValues(c, want) {
				t.Errorf("c == %#v, want %#v", c, want)
			}

			want = UQSBytes{9}
			if x, ok := m["x"]; !ok {
				t.Errorf("'x' is missing: %v", x)
			} else if !compareValues(x, want) {
				t.Errorf("x == %#v, want %#v", x, want)
			}
		}
	}
}

type UIgnoredFields struct {
	// unexported/private fields are ignored automatically.
	unexported int
	Ignored    int `qs:"-"`
	Ignored2   int `qs:"-,req"`
	Used       int
}

func TestUIgnoredFields(t *testing.T) {
	var uif UIgnoredFields
	err := UnmarshalValues(&uif, url.Values{
		"unexported": {"1"},
		"ignored":    {"2"},
		"ignored2":   {"3"},
		"used":       {"4"},
		"-":          {"5"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var cr comparisonResults
	cr.compare("unexported", uif.unexported, 0)
	cr.compare("ignored", uif.Ignored, 0)
	cr.compare("ignored2", uif.Ignored2, 0)
	cr.compare("used", uif.Used, 4)
	if err := cr.finish(); err != nil {
		t.Error(err)
	}
}

type UNonMarshalable struct {
	FuncArray []func()
}

func TestCheckUnmarshal(t *testing.T) {
	if err := CheckUnmarshal(&UTypes{}); err != nil {
		t.Errorf("unexpected error :: %v", err)
	}
	if err := CheckUnmarshal(UTypes{}); err == nil {
		t.Error("unexpected success")
	}

	if err := CheckUnmarshal(&UNonMarshalable{}); err == nil {
		t.Error("unexpected success")
	}
	if err := CheckUnmarshal(UNonMarshalable{}); err == nil {
		t.Error("unexpected success")
	}
}

func TestCheckUnmarshalType(t *testing.T) {
	ptrTypeOK := reflect.TypeOf((*UTypes)(nil))

	if err := CheckUnmarshalType(ptrTypeOK); err != nil {
		t.Errorf("unexpected error :: %v", err)
	}
	if err := CheckUnmarshalType(ptrTypeOK.Elem()); err == nil {
		t.Error("unexpected success")
	}

	ptrTypeNotOK := reflect.TypeOf((*UNonMarshalable)(nil))

	if err := CheckUnmarshalType(ptrTypeNotOK); err == nil {
		t.Error("unexpected success")
	}
	if err := CheckUnmarshalType(ptrTypeNotOK.Elem()); err == nil {
		t.Error("unexpected success")
	}
}
