package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/view"
)

// same as embedding-single-page-application but without go-bindata, the files are "physical" stored in the
// current system directory.

var page = struct {
	Title string
}{"Welcome"}

func newApp() *siris.Application {
	app := siris.New()
	app.AttachView(view.HTML("./public", ".html"))

	app.Get("/", func(ctx context.Context) {
		ctx.ViewData("Page", page)
		ctx.View("index.html")
	})

	// or just serve index.html as it is:
	// app.Get("/", func(ctx context.Context) {
	// 	ctx.ServeFile("index.html", false)
	// })

	assetHandler := app.StaticHandler("./public", false, false)
	app.SPA(assetHandler)

	return app
}

func main() {
	app := newApp()

	// http://localhost:8080
	// http://localhost:8080/index.html
	// http://localhost:8080/app.js
	// http://localhost:8080/css/main.css
	app.Run(siris.Addr(":8080"))
}
