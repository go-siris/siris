package qs_test

import (
	"encoding/hex"
	"fmt"

	"github.com/pasztorpisti/qs"
)

// This example shows how to implement the MarshalQS and UnmarshalQS interfaces
// with a custom type that wants to handle its own marshaling and unmarshaling.
func Example_selfMarshalingType() {
	// Using the []byte type with its default marshaling.
	defaultMarshaling()

	// Using the Byte array type that implements custom marshaling.
	customMarshaling()

	// Output:
	// Default-Marshal-Result: a=0&a=1&a=2&b=3&b=4&b=5 <nil>
	// Default-Unmarshal-Result: len=2 a=[0 1 2] b=[3 4 5] <nil>
	// Custom-Marshal-Result: a=000102&b=030405 <nil>
	// Custom-Unmarshal-Result: len=2 a=[0 1 2] b=[3 4 5] <nil>
}

func defaultMarshaling() {
	queryStr, err := qs.Marshal(map[string][]byte{
		"a": {0, 1, 2},
		"b": {3, 4, 5},
	})
	fmt.Println("Default-Marshal-Result:", queryStr, err)

	var query map[string][]byte
	err = qs.Unmarshal(&query, queryStr)
	fmt.Printf("Default-Unmarshal-Result: len=%v a=%v b=%v %v\n",
		len(query), query["a"], query["b"], err)
}

func customMarshaling() {
	queryStr, err := qs.Marshal(map[string]Bytes{
		"a": {0, 1, 2},
		"b": {3, 4, 5},
	})
	fmt.Println("Custom-Marshal-Result:", queryStr, err)

	var query map[string]Bytes
	err = qs.Unmarshal(&query, queryStr)
	fmt.Printf("Custom-Unmarshal-Result: len=%v a=%v b=%v %v\n",
		len(query), query["a"], query["b"], err)
}

// Bytes implements the MarshalQS and the UnmarshalQS interfaces to marshal
// itself as a hex string instead of the usual array serialisation used in
// standard query strings.
//
// The default marshaler marshals the []byte{4, 2} array to the "key=4&key=2"
// query string. In contrast the Bytes{4, 2} array is marshaled as "key=0402".
type Bytes []byte

func (b Bytes) MarshalQS(opts *qs.MarshalOptions) ([]string, error) {
	return []string{hex.EncodeToString(b)}, nil
}

func (b *Bytes) UnmarshalQS(a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}
	*b, err = hex.DecodeString(s)
	return err
}

// Compile time check: Bytes implements the qs.MarshalQS interface.
var _ qs.MarshalQS = Bytes{}

// Compile time check: *Bytes implements the qs.UnmarshalQS interface.
var _ qs.UnmarshalQS = &Bytes{}
