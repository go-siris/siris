package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"

	"github.com/go-siris/siris/sessions"
	_ "github.com/go-siris/siris/sessions/redis"
)

func main() {

	// the rest of the code stays the same.
	app := siris.New()
	// Attach the session manager we just created
	app.AttachSessionManager("redis", &sessions.ManagerConfig{
		CookieName:      "go-session-id",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     7200,
		ProviderConfig:  "127.0.0.1:7070,100,,10",
	})

	app.Get("/", func(ctx context.Context) {
		ctx.Writef("You should navigate to the /set, /get, /delete, /clear,/destroy instead")
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

	app.Get("/destroy", func(ctx context.Context) {
		//destroy, removes the entire session data and cookie
		ctx.SessionDestroy()
	})

	app.Run(siris.Addr(":8080"))
}
