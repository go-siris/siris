package main

// In this package I'll show you how to override the existing Context's functions and methods.
// You can easly navigate to the advanced/custom-context to see how you can add new functions
// to your own context (need a custom handler).
//
// This way is far easier to understand and it's faster when you want to override existing methods:
import (
	"reflect"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/sessions"
	"github.com/go-siris/siris/view"
)

// Create your own custom Context, put any fields you wanna need.
type MyContext struct {
	// Optional Part 1: embed (optional but required if you don't want to override all context's methods)
	context.Context // it's the context/context.go#context struct but you don't need to know it.
}

var _ context.Context = &MyContext{} // optionally: validate on compile-time if MyContext implements context.Context.

// Optional Part 2:
// The only one important if you will override the Context with an embedded context.Context inside it.
func (ctx *MyContext) Next() {
	context.Next(ctx)
}

// Override any context's method you want...
// [...]

func (ctx *MyContext) HTML(htmlContents string) (int, error) {
	ctx.Application().Logger().Info("Executing .HTML function from MyContext")

	ctx.ContentType("text/html")
	return ctx.WriteString(htmlContents)
}

func main() {
	app := siris.New()
	// Register a view engine on .html files inside the ./view/** directory.
	viewEngine := view.HTML("./view", ".html")
	app.AttachView(viewEngine)

	// Register the session manager.
	sessionManager := sessions.New(sessions.Config{
		Cookie: "myappcookieid",
	})
	app.AttachSessionManager(sessionManager)

	// The only one Required:
	// here is how you define how your own context will be created and acquired from the siris' generic context pool.
	app.ContextPool.Attach(func() context.Context {
		return &MyContext{
			// Optional Part 3:
			Context: context.NewContext(app),
		}
	})

	// register your route, as you normally do
	app.Handle("GET", "/", recordWhichContextJustForProofOfConcept, func(ctx context.Context) {
		// use the context's overridden HTML method.
		ctx.HTML("<h1> Hello from my custom context's HTML! </h1>")
	})

	// this will be executed by the MyContext.Context
	// if MyContext is not directly define the View function by itself.
	app.Handle("GET", "/hi/{firstname:alphabetical}", recordWhichContextJustForProofOfConcept, func(ctx context.Context) {
		firstname := ctx.Values().GetString("firstname")

		ctx.ViewData("firstname", firstname)
		ctx.Gzip(true)

		ctx.View("hi.html")
	})

	app.Run(siris.Addr(":8080"))
}

// should always print "($PATH) Handler is executing from 'MyContext'"
func recordWhichContextJustForProofOfConcept(ctx context.Context) {
	ctx.Application().Logger().Info("(%s) Handler is executing from: '%s'", ctx.Path(), reflect.TypeOf(ctx).Elem().Name())
	ctx.Next()
}
