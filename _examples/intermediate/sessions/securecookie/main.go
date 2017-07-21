package main

// $ go run main.go

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/sessions"
)

func main() {
	app := siris.New()

	app.AttachSessionManager("cookie", &sessions.ManagerConfig{
		CookieName:      "go-session-id",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     7200,
		Domain:          "example.com",
		EnableSetCookie: true,
		ProviderConfig:  "{\"cookieName\":\"go-session-id\",\"domain\":\"example.com\",\"securityKey\":\"siriscookiesecretkey\"}",
	})

	app.Get("/", func(ctx context.Context) {
		ctx.Writef("You should navigate to the /set, /get, /delete, /clear, /regenerate, /destroy instead")
	})
	app.Get("/set", func(ctx context.Context) {

		//set session values
		ctx.Session().Set("name", "siris")

		//test if set here
		ctx.Writef("All ok session set to: %s", ctx.Session().Get("name"))
	})

	app.Get("/get", func(ctx context.Context) {
		// get a specific key, as string, if no found returns just an empty string
		name := ctx.Session().Get("name")

		ctx.Writef("The name on the /set was: %s", name)
	})

	app.Get("/delete", func(ctx context.Context) {
		// delete a specific key
		ctx.Session().Delete("name")
	})

	app.Get("/clear", func(ctx context.Context) {
		// removes all entries
		ctx.Session().Flush()
	})

	app.Get("/regenerate", func(ctx context.Context) {
		// removes all entries
		ctx.SessionRegenerateID()
	})

	app.Get("/destroy", func(ctx context.Context) {
		//destroy, removes the entire session data and cookie
		ctx.SessionDestroy()
	}) // Note about destroy:
	//
	// You can destroy a session outside of a handler too, using the:
	// mySessions.DestroyByID
	// mySessions.DestroyAll

	app.Run(siris.Addr(":8080"))
}
