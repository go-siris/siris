package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/view"
)

type mypage struct {
	Title   string
	Message string
}

func main() {
	app := siris.New()

	app.AttachView(view.HTML("./templates", ".html").Layout("layout.html"))
	// TIP: append .Reload(true) to reload the templates on each request.

	app.Get("/", func(ctx context.Context) {
		ctx.Gzip(true)
		ctx.ViewData("", mypage{"My Page title", "Hello world!"})
		ctx.View("mypage.html")
		// Note that: you can pass "layout" : "otherLayout.html" to bypass the config's Layout property
		// or view.NoLayout to disable layout on this render action.
		// third is an optional parameter
	})

	// http://localhost:8080
	app.Run(siris.Addr(":8080"))
}
