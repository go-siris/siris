package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

func main() {
	app := siris.New()
	app.Handle("GET", "/", func(ctx context.Context) {
		ctx.HTML("<b> Hello world! </b>")
	})
	app.Run(siris.Addr(":8080"))
}
