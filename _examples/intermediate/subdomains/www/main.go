package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

func newApp() *siris.Application {
	app := siris.New()

	app.Get("/", info)
	app.Get("/about", info)
	app.Get("/contact", info)

	usersAPI := app.Party("/api/users")
	{
		usersAPI.Get("/", info)
		usersAPI.Get("/{id:int}", info)

		usersAPI.Post("/", info)

		usersAPI.Put("/{id:int}", info)
	}

	www := app.Party("www.")
	{
		// get all routes that are registered so far, including all "Parties":
		currentRoutes := app.GetRoutes()
		// register them to the www subdomain/vhost as well:
		for _, r := range currentRoutes {
			if _, err := www.Handle(r.Method, r.Path, r.Handlers...); err != nil {
				app.Log("%s for www. failed: %v", r.Path, err)
			}
		}
	}

	return app
}

func main() {
	app := newApp()
	// http://go-siris.com
	// http://go-siris.com/about
	// http://go-siris.com/contact
	// http://go-siris.com/api/users
	// http://go-siris.com/api/users/42

	// http://www.go-siris.com
	// http://www.go-siris.com/about
	// http://www.go-siris.com/contact
	// http://www.go-siris.com/api/users
	// http://www.go-siris.com/api/users/42
	if err := app.Run(siris.Addr("go-siris.com:80")); err != nil {
		panic(err)
	}
}

func info(ctx context.Context) {
	method := ctx.Method()
	subdomain := ctx.Subdomain()
	path := ctx.Path()

	ctx.Writef("\nInfo\n\n")
	ctx.Writef("Method: %s\nSubdomain: %s\nPath: %s", method, subdomain, path)
}
