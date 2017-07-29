package main

import (
	basicauth "github.com/go-siris/middleware-basicauth"
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

func newApp() *siris.Application {
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

func main() {
	app := newApp()
	app.Run(siris.Addr(":8080"))
}
