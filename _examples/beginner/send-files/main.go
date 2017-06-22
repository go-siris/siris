package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

func main() {
	app := siris.New()

	app.Get("/", func(ctx context.Context) {
		file := "./files/first.zip"
		ctx.SendFile(file, "c.zip")
	})

	app.Run(siris.Addr(":8080"))
}
