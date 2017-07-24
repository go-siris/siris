// Copyright 2016 Tim Heckman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpforwarded_test

import (
	"fmt"

	"github.com/theckman/httpforwarded"
)

func ExampleParse() {
	// mock value of HTTP headers
	headers := []string{"for=192.0.2.1; proto=http"}

	// parse the fields in to one map
	params, _ := httpforwarded.Parse(headers)

	// print the origin IP address and protocolg
	fmt.Printf("origin: %s | protocol: %s", params["for"][0], params["proto"][0])
	// output: origin: 192.0.2.1 | protocol: http
}

func ExampleFormat() {
	// build a parameter map
	params := map[string][]string{
		"for":   []string{"192.0.2.1", "192.0.2.4"},
		"proto": []string{"http"},
	}

	// format the parameter map
	val := httpforwarded.Format(params)

	fmt.Print(val)
	// output: for=192.0.2.1, for=192.0.2.4; proto=http
}

func ExampleParseParameter() {
	// mock value of HTTP headers
	headers := []string{"for=192.0.2.1, for=192.0.2.42; proto=http"}

	// parse the header fields while only extracting 'for' parameter
	values, _ := httpforwarded.ParseParameter("for", headers)

	// print the slice of values
	fmt.Printf("%#v", values)
	// output: []string{"192.0.2.1", "192.0.2.42"}
}
