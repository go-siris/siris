// black-box testing
package router_test

import (
	"testing"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/httptest"
)

func TestStatic(t *testing.T) {
	app := siris.New()

	expectedFoundResponse := "body{font-size: 30px}\n"

	app.StaticWeb("/static1", "../../coverage-tests/fixtures/static")
	app.StaticServe("../../coverage-tests/fixtures/static", "/static2")
	app.StaticContent("/static3", "text/css", []byte(expectedFoundResponse))

	e := httptest.New(t, app)

	e.GET("/static1/").Expect().Status(siris.StatusNotFound)
	e.GET("/static1/test.css").Expect().Status(siris.StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/static2/").Expect().Status(siris.StatusNotFound)
	e.GET("/static2/test.css").Expect().Status(siris.StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/static3").Expect().Status(siris.StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/notfound").Expect().Status(siris.StatusNotFound)

}
