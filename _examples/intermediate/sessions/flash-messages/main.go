package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/sessions"
)

func main() {
	app := siris.New()
	// output startup banner and error logs on os.Stdout

	app.AttachSessionManager("memory", &sessions.ManagerConfig{
		CookieName:      "go-session-id",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     7200,
	})

	app.Get("/set", func(ctx context.Context) {
		ctx.Session().Set("name", "siris")
		ctx.Writef("Message set, is available for the next request")
	})

	app.Get("/get", func(ctx context.Context) {
		name := ctx.Session().Get("name")
		if name == "" {
			ctx.Writef("Empty name!!")
			return
		}
		defer ctx.Session().Delete("name")
		ctx.Writef("Hello %s", name)
	})

	app.Get("/test", func(ctx context.Context) {
		name := ctx.Session().Get("name")
		if name == "" {
			ctx.Writef("Empty name!!")
			return
		}
		defer ctx.Session().Delete("name")

		ctx.Writef("Ok you are comming from /set ,the value of the name is %s", name)
		ctx.Writef(", and again from the same context: %s", name)

	})

	app.Run(siris.Addr(":8080"))
}
