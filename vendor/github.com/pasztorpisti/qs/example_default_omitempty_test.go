package qs_test

import (
	"fmt"

	"github.com/pasztorpisti/qs"
)

// A struct field tag can mark a field with one of the keepempty and omitempty
// options for marshaling. If you don't use any of these options in the
// tag then the default marshaler uses keepempty as the default. This example
// creates a custom marshaler that uses omitempty as the default option.
// Similarly, you can change UnmarshalOptions.DefaultUnmarshalPresence to
// one of the Nil/Opt/Req options when calling NewUnmarshaler but this example
// doesn't demonstrate that.
func Example_defaultOmitEmpty() {
	customMarshaler := qs.NewMarshaler(&qs.MarshalOptions{
		DefaultMarshalPresence: qs.OmitEmpty,
	})

	type Query struct {
		Default   string
		KeepEmpty string `qs:",keepempty"`
		OmitEmpty string `qs:",omitempty"`
	}

	empty := &Query{}
	full := &Query{
		Default:   "Default",
		KeepEmpty: "KeepEmpty",
		OmitEmpty: "OmitEmpty",
	}

	queryStr, err := qs.Marshal(empty)
	fmt.Println("DefaultKeepEmpty-EmptyStruct:", queryStr, err)

	queryStr, err = customMarshaler.Marshal(empty)
	fmt.Println("DefaultOmitEmpty-EmptyStruct:", queryStr, err)

	queryStr, err = qs.Marshal(full)
	fmt.Println("DefaultKeepEmpty-FullStruct:", queryStr, err)

	queryStr, err = customMarshaler.Marshal(full)
	fmt.Println("DefaultOmitEmpty-FullStruct:", queryStr, err)

	// Output:
	// DefaultKeepEmpty-EmptyStruct: default=&keep_empty= <nil>
	// DefaultOmitEmpty-EmptyStruct: keep_empty= <nil>
	// DefaultKeepEmpty-FullStruct: default=Default&keep_empty=KeepEmpty&omit_empty=OmitEmpty <nil>
	// DefaultOmitEmpty-FullStruct: default=Default&keep_empty=KeepEmpty&omit_empty=OmitEmpty <nil>
}
