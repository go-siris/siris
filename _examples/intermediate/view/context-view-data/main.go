// this example will show you how you can set per-request data for a template outside of the main handler which calls
// the .Render, via middleware.
//
// Remember: .Render has the "binding" argument which can be used to send data to the template at any case.
package main

import (
	"time"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/view"
)

const (
	DefaultTitle  = "My Awesome Site"
	DefaultLayout = "layouts/layout.html"
)

func main() {
	app := siris.New()
	// output startup banner and error logs on os.Stdout

	// set the view engine target to ./templates folder
	app.AttachView(view.HTML("./templates", ".html").Reload(true))

	app.Use(func(ctx context.Context) {
		// set the title, current time and a layout in order to be used if and when the next handler(s) calls the .Render function
		ctx.ViewData("Title", DefaultTitle)
		now := time.Now().Format(ctx.Application().ConfigurationReadOnly().GetTimeFormat())
		ctx.ViewData("CurrentTime", now)
		ctx.ViewLayout(DefaultLayout)

		ctx.Next()
	})

	app.Get("/", func(ctx context.Context) {
		ctx.ViewData("BodyMessage", "a sample text here... set by the route handler")
		if err := ctx.View("index.html"); err != nil {
			ctx.Application().Log(err.Error())
		}
	})

	app.Get("/about", func(ctx context.Context) {
		ctx.ViewData("Title", "My About Page")
		ctx.ViewData("BodyMessage", "about text here... set by the route handler")

		// same file, just to keep things simple.
		if err := ctx.View("index.html"); err != nil {
			ctx.Application().Log(err.Error())
		}
	})

	// http://localhost:8080
	// http://localhost:8080/about
	app.Run(siris.Addr(":8080"))
}

// Notes: ViewData("", myCustomStruct{}) will set this myCustomStruct value as a root binding data,
// so any View("other", "otherValue") will probably fail.
// To clear the binding data: ctx.Set(ctx.Application().ConfigurationReadOnly().GetViewDataContextKey(), nil)
