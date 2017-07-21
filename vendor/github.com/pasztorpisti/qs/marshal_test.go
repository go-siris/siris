package qs

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// MQSBytes implements the MarshalQS interface.
// The "M" prefix stands for Marshal to avoid name collisions with the
// Unmarshal tests.
type MQSBytes []byte

func (v MQSBytes) MarshalQS(opts *MarshalOptions) ([]string, error) {
	return []string{hex.EncodeToString(v)}, nil
}

type MEmbedded2 struct {
	EI int
}

type MEmbedded struct {
	MEmbedded2
}

// MTypes is used by the TestMarshalTypes test to check the marshaling of all
// supported types.
type MTypes struct {
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
	QS     MQSBytes
	QS2    MQSBytes

	MEmbedded
}

// MUnspecified is a struct that defines the MarshalPresence tag of its fields
// as MPUnspecified.
type MUnspecified struct {
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
	QS    MQSBytes

	MEmbedded
}

// MOmitEmpty is a struct that defines the MarshalPresence tag of its fields
// as OmitEmpty.
type MOmitEmpty struct {
	S string `qs:",omitempty"`
	B bool   `qs:",omitempty"`

	I   int   `qs:",omitempty"`
	I8  int8  `qs:",omitempty"`
	I16 int16 `qs:",omitempty"`
	I32 int32 `qs:",omitempty"`
	I64 int64 `qs:",omitempty"`

	U   uint   `qs:",omitempty"`
	U8  uint8  `qs:",omitempty"`
	U16 uint16 `qs:",omitempty"`
	U32 uint32 `qs:",omitempty"`
	U64 uint64 `qs:",omitempty"`

	F32 float32 `qs:",omitempty"`
	F64 float64 `qs:",omitempty"`

	Ptr *int `qs:",omitempty"`
	// Array: omitempty should have no effect
	Array [2]int   `qs:",omitempty"`
	Slice []int    `qs:",omitempty"`
	QS    MQSBytes `qs:",omitempty"`

	// MEmbedded: omitempty should have no effect
	MEmbedded `qs:",omitempty"`
}

// MKeepEmpty is a struct that defines the MarshalPresence tag of its fields
// as KeepEmpty.
type MKeepEmpty struct {
	S string `qs:",keepempty"`
	B bool   `qs:",keepempty"`

	I   int   `qs:",keepempty"`
	I8  int8  `qs:",keepempty"`
	I16 int16 `qs:",keepempty"`
	I32 int32 `qs:",keepempty"`
	I64 int64 `qs:",keepempty"`

	U   uint   `qs:",keepempty"`
	U8  uint8  `qs:",keepempty"`
	U16 uint16 `qs:",keepempty"`
	U32 uint32 `qs:",keepempty"`
	U64 uint64 `qs:",keepempty"`

	F32 float32 `qs:",keepempty"`
	F64 float64 `qs:",keepempty"`

	// Ptr: keepempty should have no effect
	Ptr   *int   `qs:",keepempty"`
	Array [2]int `qs:",keepempty"`
	// Slice: keepempty should have no effect
	Slice []int `qs:",keepempty"`
	// Slice: keepempty should have no effect
	QS MQSBytes `qs:",keepempty"`

	// MEmbedded: keepempty should have no effect
	MEmbedded `qs:",keepempty"`
}

func expectValues(values, expected url.Values) error {
	if len(values) > len(expected) {
		var unexpected []string
		for k := range values {
			if _, ok := expected[k]; !ok {
				unexpected = append(unexpected, k)
			}
		}
		return fmt.Errorf("unexpected keys: %v", strings.Join(unexpected, ", "))
	}
	for k, v := range expected {
		v2, ok := values[k]
		if !ok {
			return fmt.Errorf("expected key is missing: %q", k)
		}
		if len(v) != len(v2) {
			return fmt.Errorf("key(%q) == %#v, want %#v", k, v2, v)
		}
		for i, s := range v {
			if v2[i] != s {
				return fmt.Errorf("key(%q) == %#v, want %#v", k, v2, v)
			}
		}
	}
	return nil
}

// TestMarshalTypes tests the marshaling of all supported types.
func TestMarshalTypes(t *testing.T) {
	var i int = 42
	vs, err := MarshalValues(&MTypes{
		S:      "str",
		B:      true,
		B2:     false,
		I:      -1,
		I8:     -8,
		I16:    -16,
		I32:    -32,
		I64:    -64,
		U:      1,
		U8:     8,
		U16:    16,
		U32:    32,
		U64:    64,
		F32:    32.32,
		F64:    64.64,
		Ptr:    &i,
		Ptr2:   nil,
		Array:  [2]int{1, 2},
		Slice:  []int{3, 4},
		Slice2: nil,
		QS:     MQSBytes{1, 2, 3},
		MEmbedded: MEmbedded{
			MEmbedded2{
				EI: 33,
			},
		},
	})
	if err != nil {
		t.Error(err)
	} else {
		expected := url.Values{
			"s":     {"str"},
			"b":     {"true"},
			"b2":    {"false"},
			"i":     {"-1"},
			"i8":    {"-8"},
			"i16":   {"-16"},
			"i32":   {"-32"},
			"i64":   {"-64"},
			"u":     {"1"},
			"u8":    {"8"},
			"u16":   {"16"},
			"u32":   {"32"},
			"u64":   {"64"},
			"f32":   {"32.32"},
			"f64":   {"64.64"},
			"ptr":   {"42"},
			"array": {"1", "2"},
			"slice": {"3", "4"},
			"qs":    {"010203"},
			"qs2":   {""},
			"ei":    {"33"},
		}
		if err := expectValues(vs, expected); err != nil {
			t.Error(err)
		}
	}
}

func TestDefaultKeepEmpty(t *testing.T) {
	var i int = 42
	// default presence: keepempty, struct presence: unspecified, fields: nozero
	{
		vs, err := MarshalValues(&MUnspecified{
			S:     "str",
			B:     true,
			I:     -1,
			I8:    -8,
			I16:   -16,
			I32:   -32,
			I64:   -64,
			U:     1,
			U8:    8,
			U16:   16,
			U32:   32,
			U64:   64,
			F32:   32.32,
			F64:   64.64,
			Ptr:   &i,
			Array: [2]int{1, 2},
			Slice: []int{3, 4},
			QS:    MQSBytes{1, 2, 3},
			MEmbedded: MEmbedded{
				MEmbedded2{
					EI: 33,
				},
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {"str"},
				"b":     {"true"},
				"i":     {"-1"},
				"i8":    {"-8"},
				"i16":   {"-16"},
				"i32":   {"-32"},
				"i64":   {"-64"},
				"u":     {"1"},
				"u8":    {"8"},
				"u16":   {"16"},
				"u32":   {"32"},
				"u64":   {"64"},
				"f32":   {"32.32"},
				"f64":   {"64.64"},
				"ptr":   {"42"},
				"array": {"1", "2"},
				"slice": {"3", "4"},
				"qs":    {"010203"},
				"ei":    {"33"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: keepempty, struct presence: unspecified, fields: zero
	{
		vs, err := MarshalValues(&MUnspecified{})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {""},
				"b":     {"false"},
				"i":     {"0"},
				"i8":    {"0"},
				"i16":   {"0"},
				"i32":   {"0"},
				"i64":   {"0"},
				"u":     {"0"},
				"u8":    {"0"},
				"u16":   {"0"},
				"u32":   {"0"},
				"u64":   {"0"},
				"f32":   {"0"},
				"f64":   {"0"},
				"array": {"0", "0"},
				"qs":    {""},
				"ei":    {"0"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: keepempty, struct presence: keepempty, fields: nozero
	{
		vs, err := MarshalValues(&MKeepEmpty{
			S:     "str",
			B:     true,
			I:     -1,
			I8:    -8,
			I16:   -16,
			I32:   -32,
			I64:   -64,
			U:     1,
			U8:    8,
			U16:   16,
			U32:   32,
			U64:   64,
			F32:   32.32,
			F64:   64.64,
			Ptr:   &i,
			Array: [2]int{1, 2},
			Slice: []int{3, 4},
			QS:    MQSBytes{1, 2, 3},
			MEmbedded: MEmbedded{
				MEmbedded2{
					EI: 33,
				},
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {"str"},
				"b":     {"true"},
				"i":     {"-1"},
				"i8":    {"-8"},
				"i16":   {"-16"},
				"i32":   {"-32"},
				"i64":   {"-64"},
				"u":     {"1"},
				"u8":    {"8"},
				"u16":   {"16"},
				"u32":   {"32"},
				"u64":   {"64"},
				"f32":   {"32.32"},
				"f64":   {"64.64"},
				"ptr":   {"42"},
				"array": {"1", "2"},
				"slice": {"3", "4"},
				"qs":    {"010203"},
				"ei":    {"33"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: keepempty, struct presence: keepempty, fields: zero
	{
		vs, err := MarshalValues(&MKeepEmpty{})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {""},
				"b":     {"false"},
				"i":     {"0"},
				"i8":    {"0"},
				"i16":   {"0"},
				"i32":   {"0"},
				"i64":   {"0"},
				"u":     {"0"},
				"u8":    {"0"},
				"u16":   {"0"},
				"u32":   {"0"},
				"u64":   {"0"},
				"f32":   {"0"},
				"f64":   {"0"},
				"array": {"0", "0"},
				"qs":    {""},
				"ei":    {"0"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: keepempty, struct presence: omitempty, fields: nozero
	{
		vs, err := MarshalValues(&MOmitEmpty{
			S:     "str",
			B:     true,
			I:     -1,
			I8:    -8,
			I16:   -16,
			I32:   -32,
			I64:   -64,
			U:     1,
			U8:    8,
			U16:   16,
			U32:   32,
			U64:   64,
			F32:   32.32,
			F64:   64.64,
			Ptr:   &i,
			Array: [2]int{1, 2},
			Slice: []int{3, 4},
			QS:    MQSBytes{1, 2, 3},
			MEmbedded: MEmbedded{
				MEmbedded2{
					EI: 33,
				},
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {"str"},
				"b":     {"true"},
				"i":     {"-1"},
				"i8":    {"-8"},
				"i16":   {"-16"},
				"i32":   {"-32"},
				"i64":   {"-64"},
				"u":     {"1"},
				"u8":    {"8"},
				"u16":   {"16"},
				"u32":   {"32"},
				"u64":   {"64"},
				"f32":   {"32.32"},
				"f64":   {"64.64"},
				"ptr":   {"42"},
				"array": {"1", "2"},
				"slice": {"3", "4"},
				"qs":    {"010203"},
				"ei":    {"33"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: keepempty, struct presence: omitempty, fields: zero
	{
		vs, err := MarshalValues(&MOmitEmpty{})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"array": {"0", "0"},
				"ei":    {"0"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestDefaultOmitEmpty(t *testing.T) {
	marshaler := NewMarshaler(&MarshalOptions{
		DefaultMarshalPresence: OmitEmpty,
	})

	var i int = 42
	// default presence: omitempty, struct presence: unspecified, fields: nozero
	{
		vs, err := marshaler.MarshalValues(&MUnspecified{
			S:     "str",
			B:     true,
			I:     -1,
			I8:    -8,
			I16:   -16,
			I32:   -32,
			I64:   -64,
			U:     1,
			U8:    8,
			U16:   16,
			U32:   32,
			U64:   64,
			F32:   32.32,
			F64:   64.64,
			Ptr:   &i,
			Array: [2]int{1, 2},
			Slice: []int{3, 4},
			QS:    MQSBytes{1, 2, 3},
			MEmbedded: MEmbedded{
				MEmbedded2{
					EI: 33,
				},
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {"str"},
				"b":     {"true"},
				"i":     {"-1"},
				"i8":    {"-8"},
				"i16":   {"-16"},
				"i32":   {"-32"},
				"i64":   {"-64"},
				"u":     {"1"},
				"u8":    {"8"},
				"u16":   {"16"},
				"u32":   {"32"},
				"u64":   {"64"},
				"f32":   {"32.32"},
				"f64":   {"64.64"},
				"ptr":   {"42"},
				"array": {"1", "2"},
				"slice": {"3", "4"},
				"qs":    {"010203"},
				"ei":    {"33"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: omitempty, struct presence: unspecified, fields: zero
	{
		vs, err := marshaler.MarshalValues(&MUnspecified{})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"array": {"0", "0"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: omitempty, struct presence: keepempty, fields: nozero
	{
		vs, err := marshaler.MarshalValues(&MKeepEmpty{
			S:     "str",
			B:     true,
			I:     -1,
			I8:    -8,
			I16:   -16,
			I32:   -32,
			I64:   -64,
			U:     1,
			U8:    8,
			U16:   16,
			U32:   32,
			U64:   64,
			F32:   32.32,
			F64:   64.64,
			Ptr:   &i,
			Array: [2]int{1, 2},
			Slice: []int{3, 4},
			QS:    MQSBytes{1, 2, 3},
			MEmbedded: MEmbedded{
				MEmbedded2{
					EI: 33,
				},
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {"str"},
				"b":     {"true"},
				"i":     {"-1"},
				"i8":    {"-8"},
				"i16":   {"-16"},
				"i32":   {"-32"},
				"i64":   {"-64"},
				"u":     {"1"},
				"u8":    {"8"},
				"u16":   {"16"},
				"u32":   {"32"},
				"u64":   {"64"},
				"f32":   {"32.32"},
				"f64":   {"64.64"},
				"ptr":   {"42"},
				"array": {"1", "2"},
				"slice": {"3", "4"},
				"qs":    {"010203"},
				"ei":    {"33"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: omitempty, struct presence: keepempty, fields: zero
	{
		vs, err := marshaler.MarshalValues(&MKeepEmpty{})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {""},
				"b":     {"false"},
				"i":     {"0"},
				"i8":    {"0"},
				"i16":   {"0"},
				"i32":   {"0"},
				"i64":   {"0"},
				"u":     {"0"},
				"u8":    {"0"},
				"u16":   {"0"},
				"u32":   {"0"},
				"u64":   {"0"},
				"f32":   {"0"},
				"f64":   {"0"},
				"array": {"0", "0"},
				"qs":    {""},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: omitempty, struct presence: omitempty, fields: nozero
	{
		vs, err := marshaler.MarshalValues(&MOmitEmpty{
			S:     "str",
			B:     true,
			I:     -1,
			I8:    -8,
			I16:   -16,
			I32:   -32,
			I64:   -64,
			U:     1,
			U8:    8,
			U16:   16,
			U32:   32,
			U64:   64,
			F32:   32.32,
			F64:   64.64,
			Ptr:   &i,
			Array: [2]int{1, 2},
			Slice: []int{3, 4},
			QS:    MQSBytes{1, 2, 3},
			MEmbedded: MEmbedded{
				MEmbedded2{
					EI: 33,
				},
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"s":     {"str"},
				"b":     {"true"},
				"i":     {"-1"},
				"i8":    {"-8"},
				"i16":   {"-16"},
				"i32":   {"-32"},
				"i64":   {"-64"},
				"u":     {"1"},
				"u8":    {"8"},
				"u16":   {"16"},
				"u32":   {"32"},
				"u64":   {"64"},
				"f32":   {"32.32"},
				"f64":   {"64.64"},
				"ptr":   {"42"},
				"array": {"1", "2"},
				"slice": {"3", "4"},
				"qs":    {"010203"},
				"ei":    {"33"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	// default presence: omitempty, struct presence: omitempty, fields: zero
	{
		vs, err := marshaler.MarshalValues(&MOmitEmpty{})
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"array": {"0", "0"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestMarshalMap(t *testing.T) {
	// A map is marshallable like a struct. The key type has to be string and
	// it acts as the "struct field name". The value type has to be something
	// that you would be able to use as a struct field.
	m := map[string]MQSBytes{
		"a": []byte{0, 1, 2},
		"b": []byte{3, 4, 5},
		"c": []byte{},
	}

	{
		// default presence: keepempty
		vs, err := MarshalValues(m)
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"a": {"000102"},
				"b": {"030405"},
				"c": {""},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}

	{
		marshaler := NewMarshaler(&MarshalOptions{
			DefaultMarshalPresence: OmitEmpty,
		})

		// default presence: omitempty
		vs, err := marshaler.MarshalValues(m)
		if err != nil {
			t.Error(err)
		} else {
			expected := url.Values{
				"a": {"000102"},
				"b": {"030405"},
			}
			if err := expectValues(vs, expected); err != nil {
				t.Error(err)
			}
		}
	}
}

type MIgnoredFields struct {
	// unexported/private fields are ignored automatically.
	unexported int
	Ignored    int `qs:"-"`
	Ignored2   int `qs:"-,keepempty"`
	Used       int
}

func TestMIgnoredFields(t *testing.T) {
	vs, err := MarshalValues(&MIgnoredFields{
		unexported: 1,
		Ignored:    2,
		Ignored2:   3,
		Used:       4,
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := url.Values{
		"used": {"4"},
	}
	if err := expectValues(vs, expected); err != nil {
		t.Error(err)
	}
}

func TestMarshalNonPointer(t *testing.T) {
	// An instance of MOmitEmpty is passed by value.
	vs, err := MarshalValues(MOmitEmpty{})
	if err != nil {
		t.Fatal(err)
	}
	expected := url.Values{
		"array": {"0", "0"},
		"ei":    {"0"},
	}
	if err := expectValues(vs, expected); err != nil {
		t.Error(err)
	}
}

type MNonMarshalable struct {
	FuncArray []func()
}

func TestCheckMarshal(t *testing.T) {
	if err := CheckMarshal(&MTypes{}); err != nil {
		t.Errorf("unexpected error :: %v", err)
	}
	if err := CheckMarshal(MTypes{}); err != nil {
		t.Errorf("unexpected error :: %v", err)
	}

	if err := CheckMarshal(&MNonMarshalable{}); err == nil {
		t.Error("unexpected success")
	}
	if err := CheckMarshal(MNonMarshalable{}); err == nil {
		t.Error("unexpected success")
	}
}

func TestCheckMarshalType(t *testing.T) {
	ptrTypeOK := reflect.TypeOf((*MTypes)(nil))

	if err := CheckMarshalType(ptrTypeOK); err != nil {
		t.Errorf("unexpected error :: %v", err)
	}
	if err := CheckMarshalType(ptrTypeOK.Elem()); err != nil {
		t.Errorf("unexpected error :: %v", err)
	}

	ptrTypeNotOK := reflect.TypeOf((*MNonMarshalable)(nil))

	if err := CheckMarshalType(ptrTypeNotOK); err == nil {
		t.Error("unexpected success")
	}
	if err := CheckMarshalType(ptrTypeNotOK.Elem()); err == nil {
		t.Error("unexpected success")
	}
}
