package main

import (
	"net/http"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/core/handlerconv"
)

func main() {
	app := siris.New()
	irisMiddleware := handlerconv.FromStdWithNext(negronilikeTestMiddleware)
	app.Use(sirisMiddleware)

	// Method GET: http://localhost:8080/
	app.Get("/", func(ctx context.Context) {
		ctx.HTML("<h1> Home </h1>")
		// this will print an error,
		// this route's handler will never be executed because the middleware's criteria not passed.
	})

	// Method GET: http://localhost:8080/ok
	app.Get("/ok", func(ctx context.Context) {
		ctx.Writef("Hello world!")
		// this will print "OK. Hello world!".
	})

	// http://localhost:8080
	// http://localhost:8080/ok
	app.Run(siris.Addr(":8080"))
}

func negronilikeTestMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Path == "/ok" && r.Method == "GET" {
		w.Write([]byte("OK. "))
		next(w, r) // go to the next route's handler
		return
	}
	// else print an error and do not forward to the route's handler.
	w.WriteHeader(siris.StatusBadRequest)
	w.Write([]byte("Bad request"))
}
