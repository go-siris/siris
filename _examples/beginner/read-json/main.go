package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

type Company struct {
	Name  string
	City  string
	Other string
}

func MyHandler(ctx context.Context) {
	c := &Company{}
	if err := ctx.ReadJSON(c); err != nil {
		ctx.StatusCode(siris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	ctx.Writef("Received: %#v\n", c)
}

func main() {
	app := siris.New()

	app.Post("/", MyHandler)

	// use Postman or whatever to do a POST request
	// to the http://localhost:8080 with RAW BODY:
	/*
		{
			"Name": "siris-Go",
			"City": "New York",
			"Other": "Something here"
		}
	*/
	// and Content-Type to application/json
	//
	// The response should be:
	// Received: &main.Company{Name:"siris-Go", City:"New York", Other:"Something here"}
	app.Run(siris.Addr(":8080"))
}
