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
	mySessions := sessions.New(sessions.Config{
		// Cookie string, the session's client cookie name, for example: "mysessionid"
		//
		// Defaults to "sirissessionid"
		Cookie: "mysessionid",
		// it's time.Duration, from the time cookie is created, how long it can be alive?
		// 0 means no expire.
		// -1 means expire when browser closes
		// or set a value, like 2 hours:
		Expires: time.Hour * 2,
		// if you want to invalid cookies on different subdomains
		// of the same host, then enable it
		DisableSubdomainPersistence: false,
		// want to be crazy safe? Take a look at the "securecookie" example folder.
	})

	app.AttachSessionManager(mySessions) // Attach the session manager we just created.

	app.Get("/", func(ctx context.Context) {
		ctx.Writef("You should navigate to the /set, /get, /delete, /clear,/destroy instead")
	})
	app.Get("/set", func(ctx context.Context) {

		//set session values.

		ctx.Session().Set("name", "siris")

		//test if set here
		ctx.Writef("All ok session set to: %s", ctx.Session().GetString("name"))

		// Set will set the value as-it-is,
		// if it's a slice or map
		// you will be able to change it on .Get directly!
		// Keep note that I don't recommend saving big data neither slices or maps on a session
		// but if you really need it then use the `SetImmutable` instead of `Set`.
		// Use `SetImmutable` consistently, it's slower.
		// Read more about muttable and immutable go types: https://stackoverflow.com/a/8021081
	})

	app.Get("/get", func(ctx context.Context) {
		// get a specific value, as string, if no found returns just an empty string
		name := ctx.Session().GetString("name")

		ctx.Writef("The name on the /set was: %s", name)
	})

	app.Get("/delete", func(ctx context.Context) {
		// delete a specific key
		ctx.Session().Delete("name")
	})

	app.Get("/clear", func(ctx context.Context) {
		// removes all entries
		ctx.Session().Clear()
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

	// remember: slices and maps are muttable by-design
	// The `SetImmutable` makes sure that they will be stored and received
	// as immutable, so you can't change them directly by mistake.
	//
	// Use `SetImmutable` consistently, it's slower than `Set`.
	// Read more about muttable and immutable go types: https://stackoverflow.com/a/8021081
	app.Get("/set_immutable", func(ctx context.Context) {
		business := []businessModel{{Name: "Edward"}, {Name: "value 2"}}
		ctx.Session().SetImmutable("businessEdit", business)
		businessGet := ctx.Session().Get("businessEdit").([]businessModel)

		// try to change it, if we used `Set` instead of `SetImmutable` this
		// change will affect the underline array of the session's value "businessEdit", but now it will not.
		businessGet[0].Name = "Gabriel"

	})

	app.Get("/get_immutable", func(ctx context.Context) {
		valSlice := ctx.Session().Get("businessEdit")
		if valSlice == nil {
			ctx.HTML("please navigate to the <a href='/set_immutable'>/set_immutable</a> first")
			return
		}

		firstModel := valSlice.([]businessModel)[0]
		// businessGet[0].Name is equal to Edward initially
		if firstModel.Name != "Edward" {
			panic("Report this as a bug, immutable data cannot be changed from the caller without re-SetImmutable")
		}

		ctx.Writef("[]businessModel[0].Name remains: %s", firstModel.Name)

		// the name should remains "Edward"
	})

	app.Run(siris.Addr(":8080"))
}
