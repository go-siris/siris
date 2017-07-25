// Copyright 2016 Tim Heckman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpforwarded_test

import (
	"github.com/theckman/httpforwarded"
	. "gopkg.in/check.v1"
)

func (*TestSuite) TestParse(c *C) {
	testParseMisc(c)
	testParseSingleParam(c)
	testParseMultiParam(c)
	testParseMultiLine(c)
	testParseMultiParamValue(c)
	testParseAllTheThings(c)
}

func testParseMisc(c *C) {
	var err error
	var vals map[string][]string

	vals, err = httpforwarded.Parse(nil)
	c.Check(err, IsNil)
	c.Check(vals, IsNil)

	var values []string

	vals, err = httpforwarded.Parse(values)
	c.Check(err, IsNil)
	c.Check(vals, IsNil)
}

func testParseSingleParam(c *C) {
	var err error
	var params map[string][]string

	values := []string{"for=192.0.2.1"}

	params, err = httpforwarded.Parse(values)
	c.Assert(err, IsNil)
	c.Check(len(params), Equals, 1)

	forVal, ok := params["for"]
	c.Assert(ok, Equals, true)
	c.Assert(len(forVal), Equals, 1)
	c.Check(forVal[0], Equals, "192.0.2.1")
}

func testParseMultiParam(c *C) {
	var err error
	var params map[string][]string

	values := []string{"for=192.0.2.1; proto=http"}

	params, err = httpforwarded.Parse(values)
	c.Assert(err, IsNil)
	c.Check(len(params), Equals, 2)

	forVal, ok := params["for"]
	c.Assert(ok, Equals, true)
	c.Assert(len(forVal), Equals, 1)
	c.Check(forVal[0], Equals, "192.0.2.1")

	proto, ok := params["proto"]
	c.Assert(ok, Equals, true)
	c.Assert(len(proto), Equals, 1)
	c.Check(proto[0], Equals, "http")
}

func testParseMultiLine(c *C) {
	var err error
	var params map[string][]string

	values := []string{
		"for=192.0.2.1",
		"proto=http; by=192.0.2.200",
	}

	params, err = httpforwarded.Parse(values)
	c.Assert(err, IsNil)
	c.Check(len(params), Equals, 3)

	forVal, ok := params["for"]
	c.Assert(ok, Equals, true)
	c.Assert(len(forVal), Equals, 1)
	c.Check(forVal[0], Equals, "192.0.2.1")

	proto, ok := params["proto"]
	c.Assert(ok, Equals, true)
	c.Assert(len(proto), Equals, 1)
	c.Check(proto[0], Equals, "http")

	by, ok := params["by"]
	c.Assert(ok, Equals, true)
	c.Assert(len(by), Equals, 1)
	c.Check(by[0], Equals, "192.0.2.200")
}

func testParseMultiParamValue(c *C) {
	var err error
	var params map[string][]string

	values := []string{"for=192.0.2.1, for=192.0.2.4; proto=http"}

	params, err = httpforwarded.Parse(values)
	c.Assert(err, IsNil)
	c.Check(len(params), Equals, 2)

	forVal, ok := params["for"]
	c.Assert(ok, Equals, true)
	c.Assert(len(forVal), Equals, 2)
	c.Check(forVal[0], Equals, "192.0.2.1")
	c.Check(forVal[1], Equals, "192.0.2.4")

	proto, ok := params["proto"]
	c.Assert(ok, Equals, true)
	c.Assert(len(proto), Equals, 1)
	c.Check(proto[0], Equals, "http")
}

func testParseAllTheThings(c *C) {
	var err error
	var params map[string][]string

	values := []string{
		"for=192.0.2.1; proto=http",
		"by=192.0.2.200",
		"for=192.0.2.4, for=192.0.2.10; by=192.0.2.202",
	}

	params, err = httpforwarded.Parse(values)
	c.Assert(err, IsNil)
	c.Check(len(params), Equals, 3)

	forVal, ok := params["for"]
	c.Assert(ok, Equals, true)
	c.Assert(len(forVal), Equals, 3)
	c.Check(forVal[0], Equals, "192.0.2.1")
	c.Check(forVal[1], Equals, "192.0.2.4")
	c.Check(forVal[2], Equals, "192.0.2.10")

	proto, ok := params["proto"]
	c.Assert(ok, Equals, true)
	c.Assert(len(proto), Equals, 1)
	c.Check(proto[0], Equals, "http")

	by, ok := params["by"]
	c.Assert(ok, Equals, true)
	c.Assert(len(by), Equals, 2)
	c.Check(by[0], Equals, "192.0.2.200")
	c.Check(by[1], Equals, "192.0.2.202")
}

func (*TestSuite) TestParseParameter(c *C) {
	testParseParameterMisc(c)
	testParseParameterSingleParam(c)
	testParseParameterMultiParam(c)
	testParseParameterMultiLine(c)
	testParseParameterAllTheThings(c)
}

func testParseParameterMisc(c *C) {
	var err error
	var vals []string

	vals, err = httpforwarded.ParseParameter("for", nil)
	c.Check(err, IsNil)
	c.Check(vals, IsNil)

	var values []string

	vals, err = httpforwarded.ParseParameter("for", values)
	c.Check(err, IsNil)
	c.Check(vals, IsNil)

	values = []string{"for=192.0.2.1"}

	vals, err = httpforwarded.ParseParameter("", values)
	c.Check(err, ErrorMatches, `paramName must not be ""`)
	c.Check(vals, IsNil)
}

func testParseParameterSingleParam(c *C) {
	var err error
	var vals []string

	values := []string{"for=192.0.2.1"}

	vals, err = httpforwarded.ParseParameter("for", values)
	c.Assert(err, IsNil)
	c.Check(len(vals), Equals, 1)
	c.Check(vals[0], Equals, "192.0.2.1")
}

func testParseParameterMultiParam(c *C) {
	var err error
	var vals []string

	values := []string{"for=192.0.2.1, for=192.0.2.4; proto=http"}

	vals, err = httpforwarded.ParseParameter("for", values)
	c.Assert(err, IsNil)
	c.Check(len(vals), Equals, 2)
	c.Check(vals[0], Equals, "192.0.2.1")
	c.Check(vals[1], Equals, "192.0.2.4")

	vals, err = httpforwarded.ParseParameter("proto", values)
	c.Assert(err, IsNil)
	c.Check(len(vals), Equals, 1)
	c.Check(vals[0], Equals, "http")
}

func testParseParameterMultiLine(c *C) {
	var err error
	var vals []string

	values := []string{"for=192.0.2.1", "for=192.0.2.4"}

	vals, err = httpforwarded.ParseParameter("for", values)
	c.Assert(err, IsNil)
	c.Check(len(vals), Equals, 2)
	c.Check(vals[0], Equals, "192.0.2.1")
	c.Check(vals[1], Equals, "192.0.2.4")
}

func testParseParameterAllTheThings(c *C) {
	var err error
	var vals []string

	values := []string{
		"for=192.0.2.1; proto=http",
		"by=192.0.2.200",
		"for=192.0.2.4, for=192.0.2.10; by=192.0.2.202",
	}

	vals, err = httpforwarded.ParseParameter("for", values)
	c.Assert(err, IsNil)
	c.Check(len(vals), Equals, 3)
	c.Check(vals[0], Equals, "192.0.2.1")
	c.Check(vals[1], Equals, "192.0.2.4")
	c.Check(vals[2], Equals, "192.0.2.10")

	vals, err = httpforwarded.ParseParameter("by", values)
	c.Assert(err, IsNil)
	c.Check(len(vals), Equals, 2)
	c.Check(vals[0], Equals, "192.0.2.200")
	c.Check(vals[1], Equals, "192.0.2.202")
}
