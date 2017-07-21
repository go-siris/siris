package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"

	"github.com/go-siris/siris/sessions"
)

var (
	key = "my_sessionid"
)

func secret(ctx context.Context) {

	// Check if user is authenticated
	if auth, _ := ctx.Session().Get("authenticated"); !auth {
		ctx.StatusCode(siris.StatusForbidden)
		return
	}

	// Print secret message
	ctx.WriteString("The cake is a lie!")
}

func login(ctx context.Context) {
	session := ctx.Session()

	// Authentication goes here
	// ...

	// Set user as authenticated
	session.Set("authenticated", true)
}

func logout(ctx context.Context) {
	session := ctx.Session()

	// Revoke users authentication
	session.Set("authenticated", false)
}

func main() {
	app := siris.New()

	// Look https://github.com/go-siris/siris/tree/master/sessions/_examples for more features,
	// i.e encode/decode and lifetime.
	app.AttachSessionManager("memory", &sessions.ManagerConfig{
		CookieName:      key,
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     7200,
	})

	app.Get("/secret", secret)
	app.Get("/login", login)
	app.Get("/logout", logout)

	app.Run(siris.Addr(":8080"))
}
