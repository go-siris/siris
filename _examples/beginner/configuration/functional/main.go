package main

import (
	"github.com/go-siris/siris"
)

func main() {
	app := siris.New()

	// [...]

	// Good when you want to change some of the configuration's field.
	// I use that method :)
	app.Run(siris.Addr(":8080"), siris.WithoutBanner, siris.WithCharset("UTF-8"))

	// or before run:
	// app.Configure(siris.WithoutBanner, siris.WithCharset("UTF-8"))
	// app.Run(siris.Addr(":8080"))
}
