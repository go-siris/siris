// Copyright 2017 Go-SIRIS Author. All Rights Reserved.

package coverageTests

// Use as Example
/*

import (
	"net/http"
	"testing"
	"time"

	//"github.com/stretchr/testify/assert"
	"gopkg.in/gavv/httpexpect.v1"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
)


type Query struct {
	Search     string
	Page       int
	PageSize   int
	Categories []string `qs:"category"`
}

func createApp_Query() *siris.Application {
	app := siris.Default()

	app.Get("/read", func(c context.Context) {
		var q Query
		err := c.ReadQuery(&q)
		if err != nil {
			c.StatusCode(siris.StatusNotFound)
			return
		}

		queryString, err2 := c.CreateQuery(q)
		if err2 != nil {
			c.StatusCode(siris.StatusNotFound)
			return
		}

		c.StatusCode(siris.StatusOK)
		c.Text(queryString)
	})

	app.Get("/write", func(c context.Context) {
		queryString, err := c.CreateQuery(&Query{
			Search:     "my search",
			Page:       2,
			PageSize:   50,
			Categories: []string{"c1", "c2"},
		})
		if err != nil {
			c.StatusCode(siris.StatusNotFound)
			return
		}

		c.StatusCode(siris.StatusOK)
		c.Text(queryString)
		return
	})

	app.Build()

	return app
}

func createClient_Query(t *testing.T) *httpexpect.Expect {
	handler := createApp_Query()

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

func TestSiris_Query(t *testing.T) {
	e := createClient_Query(t)

	e.GET("/write").
		Expect().
		Status(http.StatusOK).Body().Equal("category=c1&category=c2&page=2&page_size=50&search=my+search")

	e.GET("/read").
		WithQueryString("category=c3&category=c4&page=5&page_size=500&search=my+siris").
		Expect().
		Status(http.StatusOK).Body().Equal("category=c3&category=c4&page=5&page_size=500&search=my+siris")
}
*/
