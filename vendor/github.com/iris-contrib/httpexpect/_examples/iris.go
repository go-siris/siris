package examples

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/middleware/basicauth"
	"github.com/kataras/iris/sessions"
)

// IrisHandler tests iris v6's handler
func IrisHandler() http.Handler {
	api := iris.New()
	api.AttachSessionManager(sessions.New(sessions.Config{Cookie: "irissessionid"}))

	api.Get("/things", func(ctx context.Context) {
		ctx.JSON([]interface{}{
			context.Map{
				"name":        "foo",
				"description": "foo thing",
			},
			context.Map{
				"name":        "bar",
				"description": "bar thing",
			},
		})
	})

	api.Post("/redirect", func(ctx context.Context) {
		ctx.Redirect("/things", iris.StatusFound)
	})

	api.Post("/params/{x}/{y}", func(ctx context.Context) {
		ctx.JSON(context.Map{
			"x":  ctx.Params().Get("x"),
			"y":  ctx.Params().Get("y"),
			"q":  ctx.URLParam("q"),
			"p1": ctx.FormValue("p1"),
			"p2": ctx.FormValue("p2"),
		})
	})

	auth := basicauth.Default(map[string]string{
		"ford": "betelgeuse7",
	})

	api.Get("/auth", auth, func(ctx context.Context) {
		ctx.Writef("authenticated!")
	})

	api.Post("/session/set", func(ctx context.Context) {
		sess := context.Map{}

		if err := ctx.ReadJSON(&sess); err != nil {
			panic(err.Error())
		}

		ctx.Session().Set("name", sess["name"])
	})

	api.Get("/session/get", func(ctx context.Context) {
		name := ctx.Session().GetString("name")

		ctx.JSON(context.Map{
			"name": name,
		})
	})

	api.Get("/stream", func(ctx context.Context) {
		ctx.StreamWriter(func(w io.Writer) bool {
			for i := 0; i < 10; i++ {
				fmt.Fprintf(w, "%d", i)
			}
			// return true to continue, return false to stop and flush
			return false
		})
		// if we had to write here then the StreamWriter callback should
		// return true
	})

	api.Post("/stream", func(ctx context.Context) {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.StopExecution()
			return
		}
		ctx.Write(body)
	})

	sub := api.Party("subdomain.")

	sub.Post("/set", func(ctx context.Context) {
		ctx.Session().Set("message", "hello from subdomain")
	})

	sub.Get("/get", func(ctx context.Context) {
		ctx.Writef(ctx.Session().GetString("message"))
	})

	api.Build()
	return api
}
