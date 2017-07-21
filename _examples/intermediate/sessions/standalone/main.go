package main

import (
	"time"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/sessions"
)

type businessModel struct {
	Name string
}

func main() {
	app := siris.New()
	app.AttachSessionManager("memory", &sessions.ManagerConfig{
		CookieName:      "go-session-id",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     7200,
	})

	app.Get("/", func(ctx context.Context) {
		ctx.Writef("You should navigate to the /set, /get, /delete, /clear, /destroy instead")
	})
	app.Get("/set", func(ctx context.Context) {

		//set session values.

		ctx.Session().Set("name", "siris")

		//test if set here
		ctx.Writef("All ok session set to: %s", ctx.Session().Get("name"))

		// Set will set the value as-it-is,
		// if it's a slice or map
		// you will be able to change it on .Get directly!
		// Read more about muttable and immutable go types: https://stackoverflow.com/a/8021081
	})

	app.Get("/get", func(ctx context.Context) {
		// get a specific value, as string, if no found returns just an empty string
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

	app.Get("/destroy", func(ctx context.Context) {

		//destroy, removes the entire session data and cookie
		ctx.SessionDestroy()
	})
	// Note about Destroy:
	//
	// You can destroy a session outside of a handler too, using the:
	// mySessions.DestroyByID
	// mySessions.DestroyAll

	app.Run(siris.Addr(":8080"))
}
