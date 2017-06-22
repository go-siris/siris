package main

import (
	"github.com/go-siris/siris"
)

// Follow these steps first:
// $ go get -u github.com/jteeuwen/go-bindata/...
// $ go-bindata ./assets/...
// $ go build
// $ ./embedding-files-into-app
// "physical" files are not used, you can delete the "assets" folder and run the example.

func newApp() *siris.Application {
	app := siris.New()

	app.StaticEmbedded("/static", "./assets", Asset, AssetNames)

	return app
}

func main() {
	app := newApp()

	// http://localhost:8080/static/css/bootstrap.min.css
	// http://localhost:8080/static/js/jquery-2.1.1.js
	// http://localhost:8080/static/favicon.ico
	app.Run(siris.Addr(":8080"))
}
