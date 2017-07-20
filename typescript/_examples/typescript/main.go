package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"

	"github.com/go-siris/siris/typescript"
)

// NOTE: Some machines don't allow to install typescript automatically, so if you don't have typescript installed
// and the typescript adaptor doesn't works for you then follow the below steps:
// 1. close the siris server
// 2. open your terminal and execute: npm install -g typescript
// 3. start your siris server, it should be work, as expected, now.
func main() {
	app := siris.New()

	app.StaticWeb("/scripts", "./www/scripts") // serve the scripts

	app.Get("/", func(ctx context.Context) {
		ctx.ServeFile("./www/index.html", false)
	})

	ts := typescript.New()
	ts.Config.Dir = "./www/scripts"
	ts.Run(app.Logger().Infof)

	// http://localhost:8080
	app.Run(siris.Addr(":8080"))
}

// open http://localhost:8080
// go to ./www/scripts/app.ts
// make a change
// reload the http://localhost:8080 and you should see the changes
//
// what it does?
// - compiles the typescript files using default compiler options if not tsconfig found
// - watches for changes on typescript files, if a change then it recompiles the .ts to .js
//
// same as you used to do with gulp-like tools, but here I do my bests to help GO developers.
