// Copyright 2017 Go-SIRIS Author. All Rights Reserved.

package coverageTests

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	//"github.com/stretchr/testify/assert"
	"gopkg.in/gavv/httpexpect.v1"

	"github.com/go-siris/middleware-basicauth"
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)

func createApp() *siris.Application {
	app := siris.Default()

	app.Get("/things", func(c context.Context) {
		c.JSON([]map[string]interface{}{
			{
				"name":        "foo",
				"description": "foo thing",
			},
			{
				"name":        "bar",
				"description": "bar thing",
			},
		})
	})

	app.Post("/redirect", func(c context.Context) {
		c.Redirect("/things", siris.StatusFound)
	})

	app.Post("/params/:x/:y", func(c context.Context) {
		c.JSON(map[string]interface{}{
			"x":  c.Params().Get("x"),
			"y":  c.Params().Get("y"),
			"q":  c.URLParam("q"),
			"p1": c.FormValue("p1"),
			"p2": c.FormValue("p2"),
		})
	})

	authConfig := basicauth.Config{
		Users:   map[string]string{"siris": "framework", "ford": "betelgeuse7"},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(30) * time.Minute,
	}

	authentication := basicauth.New(authConfig)

	app.Get("/auth", authentication, func(c context.Context) {
		c.Writef("authenticated!")
	})

	app.Post("/session/set", func(c context.Context) {
		var sess map[string]interface{}

		if err := c.ReadJSON(&sess); err != nil {
			panic(err.Error())
		}

		c.Session().Set("name", sess["name"])
	})

	app.Get("/session/get", func(c context.Context) {
		name := c.Session().Get("name")

		c.JSON(map[string]interface{}{
			"name": name,
		})
	})

	app.Get("/stream", func(c context.Context) {
		c.StreamWriter(func(w io.Writer) bool {
			for i := 0; i < 10; i++ {
				fmt.Fprintf(w, "%d", i)
			}
			// return true to continue, return false to stop and flush
			return false
		})
		// if we had to write here then the StreamWriter callback should
		// return true
	})

	app.Post("/stream", func(c context.Context) {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			c.StatusCode(siris.StatusBadRequest)
			return
		}
		c.Write(body)
	})

	sub := app.Party("subdomain.")

	sub.Post("/set", func(c context.Context) {
		c.Session().Set("message", "hello from subdomain")
	})

	sub.Get("/get", func(c context.Context) {
		c.Text(c.Session().Get("message").(string))
	})

	app.Build()

	return app
}

func createClient(t *testing.T) *httpexpect.Expect {
	handler := createApp()

	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
			Timeout:   time.Second * 30,
		},
		// use verbose logging
		//Printers: []httpexpect.Printer{
		//	httpexpect.NewCurlPrinter(t),
		// httpexpect.NewDebugPrinter(t, false),
		//},
	})
}

func TestSiris_NewApp(t *testing.T) {
	e := createClient(t)

	schema := `{
		"type": "array",
		"items": {
			"type": "object",
			"properties": {
				"name":        {"type": "string"},
				"description": {"type": "string"}
			},
			"required": ["name", "description"]
		}
	}`

	things := e.GET("/things").
		Expect().
		Status(http.StatusOK).JSON()

	things.Schema(schema)

	names := things.Path("$[*].name").Array()

	names.Elements("foo", "bar")

	for n, desc := range things.Path("$..description").Array().Iter() {
		m := desc.String().Match("(.+) (.+)")

		m.Index(1).Equal(names.Element(n).String().Raw())
		m.Index(2).Equal("thing")
	}
}

func TestSiris_Redirect(t *testing.T) {
	e := createClient(t)

	things := e.POST("/redirect").
		Expect().
		Status(http.StatusOK).JSON().Array()

	things.Length().Equal(2)

	things.Element(0).Object().ValueEqual("name", "foo")
	things.Element(1).Object().ValueEqual("name", "bar")
}

func TestSiris_Params(t *testing.T) {
	e := createClient(t)

	type Form struct {
		P1 string `form:"p1"`
		P2 string `form:"p2"`
	}

	// POST /params/xxx/yyy?q=qqq
	// Form: p1=P1&p2=P2

	r := e.POST("/params/{x}/{y}", "xxx", "yyy").
		WithQuery("q", "qqq").WithForm(Form{P1: "P1", P2: "P2"}).
		Expect().
		Status(http.StatusOK).JSON().Object()

	r.Value("x").Equal("xxx")
	r.Value("y").Equal("yyy")
	r.Value("q").Equal("qqq")

	r.ValueEqual("p1", "P1")
	r.ValueEqual("p2", "P2")
}

func TestSiris_Auth(t *testing.T) {
	e := createClient(t)

	e.GET("/auth").
		Expect().
		Status(http.StatusUnauthorized)

	e.GET("/auth").WithBasicAuth("ford", "<bad password>").
		Expect().
		Status(http.StatusUnauthorized)

	e.GET("/auth").WithBasicAuth("ford", "betelgeuse7").
		Expect().
		Status(http.StatusOK).Body().Equal("authenticated!")
}

func TestSiris_Session(t *testing.T) {
	e := createClient(t)

	e.POST("/session/set").WithJSON(map[string]string{"name": "test"}).
		Expect().
		Status(http.StatusOK).Cookies().NotEmpty()

	r := e.GET("/session/get").
		Expect().
		Status(http.StatusOK).JSON().Object()

	r.Equal(map[string]string{
		"name": "test",
	})
}

func TestSiris_Stream(t *testing.T) {
	e := createClient(t)

	e.GET("/stream").
		Expect().
		Status(http.StatusOK).
		TransferEncoding("chunked"). // ensure server sent chunks
		Body().Equal("0123456789")

	// send chunks to server
	e.POST("/stream").WithChunked(strings.NewReader("<long text>")).
		Expect().
		Status(http.StatusOK).Body().Equal("<long text>")
}

func TestSiris_Subdomain(t *testing.T) {
	e := createClient(t)

	sub := e.Builder(func(req *httpexpect.Request) {
		req.WithURL("http://subdomain.127.0.0.1")
	})

	sub.POST("/set").
		Expect().
		Status(http.StatusOK)

	sub.GET("/get").
		Expect().
		Status(http.StatusOK).
		Body().Equal("hello from subdomain")
}
