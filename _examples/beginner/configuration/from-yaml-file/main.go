package main

import (
	"github.com/go-siris/siris"
)

func main() {
	app := siris.New()

	// [...]

	// Good when you have two configurations, one for development and a different one for production use.
	app.Run(siris.Addr(":8080"), siris.WithConfiguration(siris.YAML("./configs/iris.yml")))

	// or before run:
	// app.Configure(siris.WithConfiguration(siris.YAML("./configs/iris.yml")))
	// app.Run(siris.Addr(":8080"))
}
