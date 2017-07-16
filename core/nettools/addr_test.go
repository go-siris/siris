// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nettools

import (
	"os"
	"testing"
)

func TestIsLoopbackHost(t *testing.T) {
	tests := []struct {
		host  string
		valid bool
	}{
		{"subdomain.127.0.0.1:8080", true},
		{"subdomain.127.0.0.1", true},
		{"subdomain.localhost:8080", true},
		{"subdomain.localhost", true},
		{"subdomain.127.0000.0000.1:8080", true},
		{"subdomain.127.0000.0000.1", true},
		{"subdomain.127.255.255.254:8080", true},
		{"subdomain.127.255.255.254", true},

		{"subdomain.0000:0:0000::01.1:8080", false},
		{"subdomain.0000:0:0000::01", false},
		{"subdomain.0000:0:0000::01.1:8080", false},
		{"subdomain.0000:0:0000::01", false},
		{"subdomain.0000:0000:0000:0000:0000:0000:0000:0001:8080", true},
		{"subdomain.0000:0000:0000:0000:0000:0000:0000:0001", false},

		{"subdomain.example:8080", false},
		{"subdomain.example", false},
		{"subdomain.example.com:8080", false},
		{"subdomain.example.com", false},
		{"subdomain.com", false},
		{"subdomain", false},
		{".subdomain", false},
		{"127.0.0.1.com", false},
	}

	for i, tt := range tests {
		if expected, got := tt.valid, IsLoopbackHost(tt.host); expected != got {
			t.Fatalf("[%d] expected %t but got %t for %s", i, expected, got, tt.host)
		}
	}

	if port := ResolvePort("example.com:8080"); port != 8080 {
		t.Fatalf("ResolvePort expected %t but got %t for %s", 8080, port, "example.com:8080")
	}

	if scheme := ResolveScheme(true); scheme != SchemeHTTPS {
		t.Fatalf("ResolvePort expected %s but got %s", SchemeHTTPS, scheme)
	}

	if scheme := ResolveScheme(false); scheme != SchemeHTTP {
		t.Fatalf("ResolvePort expected %s but got %s", SchemeHTTP, scheme)
	}

	if scheme := ResolveSchemeFromVHost("example.com:443"); scheme != SchemeHTTPS {
		t.Fatalf("ResolvePort expected %s but got %s for %s", SchemeHTTPS, scheme, "example.com:443")
	}

	if url := ResolveURL(SchemeHTTPS, ":https"); url != "https://localhost" {
		t.Fatalf("ResolvePort expected %s but got %t", "https://localhost", url)
	}

	hostname, _ := os.Hostname()

	_ = IsLoopbackSubdomain("127.0.0.1:8080")
	_ = IsLoopbackSubdomain("127.1.1.1:8080")
	_ = IsLoopbackSubdomain(hostname)

	_ = ResolveAddr("")
	_ = ResolveAddr(hostname)
	_ = ResolveAddr(":http")
	_ = ResolveAddr(":80")
	_ = ResolveAddr(":https")

	_ = ResolvePort(":https")

	_ = ResolveVHost(":8080")
	_ = ResolveVHost("0.0.0.0:8080")
	_ = ResolveVHost("localhost:8080")
	_ = ResolveVHost("www.go-siris.com:8080")

	_ = ResolveHostname("localhost:https")
	_ = ResolveHostname(":8080")
	_ = ResolveHostname("localhost")
}
