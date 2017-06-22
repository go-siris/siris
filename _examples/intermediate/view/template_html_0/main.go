package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/view"
)

func main() {
	app := siris.New() // defaults to these

	// - standard html  | view.HTML(...)
	// - django         | view.Django(...)
	// - pug(jade)      | view.Pug(...)
	// - handlebars     | view.Handlebars(...)
	// - amber          | view.Amber(...)

	tmpl := view.HTML("./templates", ".html")
	tmpl.Reload(true) // reload templates on each request (development mode)
	// default template funcs are:
	//
	// - {{ urlpath "mynamedroute" "pathParameter_ifneeded" }}
	// - {{ render "header.html" }}
	// - {{ render_r "header.html" }} // partial relative path to current page
	// - {{ yield }}
	// - {{ current }}
	tmpl.AddFunc("greet", func(s string) string {
		return "Greetings " + s + "!"
	})
	app.AttachView(tmpl)

	app.Get("/", hi)

	// http://localhost:8080
	app.Run(siris.Addr(":8080"), siris.WithCharset("UTF-8")) // defaults to that but you can change it.
}

func hi(ctx context.Context) {
	ctx.ViewData("Title", "Hi Page")
	ctx.ViewData("Name", "Siris") // {{.Name}} will render: Siris
	// ctx.ViewData("", myCcustomStruct{})
	ctx.View("hi.html")
}
