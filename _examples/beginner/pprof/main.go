package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"

	"github.com/go-siris/siris/middleware/pprof"
)

func main() {
	app := siris.New()

	app.Get("/", func(ctx context.Context) {
		ctx.HTML("<h1> Please click <a href='/debug/pprof'>here</a>")
	})

	app.Any("/debug/pprof/{action:path}", pprof.New())
	//                              ___________
	app.Run(siris.Addr(":8080"))
}
