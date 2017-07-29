// black-box testing
package basicauth_test

import (
	"testing"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/httptest"
	"github.com/go-siris/siris/middleware/basicauth"
)

func buildApp() *siris.Application {
	app := siris.New()

	authConfig := basicauth.Config{
		Users: map[string]string{"myusername": "mypassword"},
	}

	authentication := basicauth.New(authConfig)

	app.Get("/", func(ctx context.Context) { ctx.Redirect("/admin") })

	// to party

	needAuth := app.Party("/admin", authentication)
	{
		//http://localhost:8080/admin
		needAuth.Get("/", h)
		// http://localhost:8080/admin/profile
		needAuth.Get("/profile", h)

		// http://localhost:8080/admin/settings
		needAuth.Get("/settings", h)
	}

	return app
}

func h(ctx context.Context) {
	username, password, _ := ctx.Request().BasicAuth()
	// third parameter it will be always true because the middleware
	// makes sure for that, otherwise this handler will not be executed.

	ctx.Writef("%s %s:%s", ctx.Path(), username, password)
}
func TestBasicAuth(t *testing.T) {
	app := buildApp()
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
