package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/core/nettools"
)

func main() {
	app := siris.New()

	l, err := nettools.UNIX("/tmpl/srv.sock", 0666) // see its code to see how you can manually create a new file listener, it's easy.
	if err != nil {
		panic(err)
	}

	app.Run(siris.Listener(l))
}
