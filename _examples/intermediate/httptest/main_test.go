package main

import (
	"testing"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/httptest"
)

// $ cd $GOPATH/src/github.com/go-siris/siris/_examples/intermediate/httptest
// $ go test -v
func TestNewApp(t *testing.T) {
	app := newApp()
	e := httptest.New(t, app)

	// redirects to /admin without basic auth
	e.GET("/").Expect().Status(siris.StatusUnauthorized)
	// without basic auth
	e.GET("/admin").Expect().Status(siris.StatusUnauthorized)

	// with valid basic auth
	e.GET("/admin").WithBasicAuth("myusername", "mypassword").Expect().
		Status(siris.StatusOK).Body().Equal("/admin myusername:mypassword")
	e.GET("/admin/profile").WithBasicAuth("myusername", "mypassword").Expect().
		Status(siris.StatusOK).Body().Equal("/admin/profile myusername:mypassword")
	e.GET("/admin/settings").WithBasicAuth("myusername", "mypassword").Expect().
		Status(siris.StatusOK).Body().Equal("/admin/settings myusername:mypassword")

	// with invalid basic auth
	e.GET("/admin/settings").WithBasicAuth("invalidusername", "invalidpassword").
		Expect().Status(siris.StatusUnauthorized)

}
