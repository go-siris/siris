package main

import (
	"net/http"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/core/handlerconv"
)

func main() {
	app := siris.New()
	irisMiddleware := handlerconv.FromStd(nativeTestMiddleware)
	app.Use(sirisMiddleware)

	// Method GET: http://localhost:8080/
	app.Get("/", func(ctx context.Context) {
		ctx.HTML("Home")
	})

	// Method GET: http://localhost:8080/ok
	app.Get("/ok", func(ctx context.Context) {
		ctx.HTML("<b>Hello world!</b>")
	})

	// http://localhost:8080
	// http://localhost:8080/ok
	app.Run(siris.Addr(":8080"))
}

func nativeTestMiddleware(w http.ResponseWriter, r *http.Request) {
	println("Request path: " + r.URL.Path)
}
