// Copyright 2016 Tim Heckman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpforwarded_test

import (
	"github.com/theckman/httpforwarded"
	. "gopkg.in/check.v1"
)

func (*TestSuite) TestFormat(c *C) {
	var out string

	out = httpforwarded.Format(nil)
	c.Check(out, Equals, "")

	params := map[string][]string{
		"for":   []string{"192.0.2.1", "192.0.2.4"},
		"by":    []string{"192.0.2.200", "192.0.2.202"},
		"proto": []string{"http"},
	}

	out = httpforwarded.Format(params)
	c.Check(out, Equals, "by=192.0.2.200, by=192.0.2.202; for=192.0.2.1, for=192.0.2.4; proto=http")
}
