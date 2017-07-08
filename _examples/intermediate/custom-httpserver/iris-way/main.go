package main

import (
	"net/http"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

func main() {
	app := siris.New()

	app.Get("/", func(ctx context.Context) {
		ctx.Writef("Hello from the server")
	})

	app.Get("/mypath", func(ctx context.Context) {
		ctx.Writef("Hello from %s", ctx.Path())
	})

	srv := &http.Server{Addr: ":8080" /* Any custom fields here: Handler and ErrorLog are set to the server automatically */}
	// http://localhost:8080/
	// http://localhost:8080/mypath
	app.Run(siris.Server(srv)) // same as app.Run(siris.Addr(":8080"))

	// More:
	// see "multi" if you need to use more than one server at the same app.
	//
	// for a custom listener use: siris.Listener(net.Listener) or
	// siris.TLS(cert,key) or siris.AutoTLS(), see "custom-listener" example for those.
}
