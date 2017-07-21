package qs_test

import (
	"fmt"

	"github.com/pasztorpisti/qs"
)

type Query struct {
	Search     string
	Page       int
	PageSize   int
	Categories []string `qs:"category"`
}

// This example demonstrates the usage of qs.Marshal and qs.Unmarshal.
func Example() {
	queryStr, err := qs.Marshal(&Query{
		Search:     "my search",
		Page:       2,
		PageSize:   50,
		Categories: []string{"c1", "c2"},
	})
	fmt.Println("Marshal-Result:", queryStr, err)

	var q Query
	err = qs.Unmarshal(&q, queryStr)
	fmt.Println("Unmarshal-Result:", q, err)

	// Output:
	// Marshal-Result: category=c1&category=c2&page=2&page_size=50&search=my+search <nil>
	// Unmarshal-Result: {my search 2 50 [c1 c2]} <nil>
}
