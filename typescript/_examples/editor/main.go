package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"

	"github.com/go-siris/siris/typescript/editor"
)

func main() {
	app := siris.Default()
	app.StaticWeb("/scripts", "./www/scripts") // serve the scripts
	// when you edit a typescript file from the alm-tools
	// it compiles it to javascript, have fun!

	app.Get("/", func(ctx context.Context) {
		ctx.ServeFile("./www/index.html", false)
	})

	editorConfig := editor.Config{
		Hostname:   "localhost",
		Port:       4444,
		WorkingDir: "./www/scripts/", // "/path/to/the/client/side/directory/",
		Username:   "myusername",
		Password:   "mypassword",
	}
	e := editor.New(editorConfig)
	e.Run(app.Logger().Infof) // start the editor's server

	// http://localhost:8080
	// http://localhost:4444
	app.Run(siris.Addr(":8080"))
	e.Stop()
}
