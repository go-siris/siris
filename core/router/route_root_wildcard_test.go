// black-box testing
//
// see _examples/routing/main_test.go for the most common router tests that you may want to see,
// this is a test for the new feature that I just coded: wildcard "/{p:path}" on root without conflicts

package router_test

import (
	"net/http"
	"testing"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/httptest"
)

const (
	same_as_request_path                         = "same"
	from_status_code                             = "from"
	staticPathPrefixBody                         = "from the static path: "
	prefix_static_path_following_by_request_path = "prefix_same"
)

type testRouteRequest struct {
	method             string
	subdomain          string
	path               string
	expectedStatusCode int
	expectedBody       string
}

type testRoute struct {
	method   string
	path     string
	handler  context.Handler
	requests []testRouteRequest
}

var h = func(ctx context.Context) {
	ctx.WriteString(ctx.Path())
}

var h2 = func(ctx context.Context) {
	ctx.StatusCode(siris.StatusForbidden) // ! 200 but send the body as expected,
	// we need that kind of behavior to determinate which handler is executed for routes that
	// both having wildcard path but first one is registered on root level.
	ctx.WriteString(ctx.Path())
}

func h3(ctx context.Context) {
	ctx.Writef(staticPathPrefixBody + ctx.Path())
}

func TestRouterWildcardDifferentPrefixPath(t *testing.T) {
	var tt = []testRoute{
		{"GET", "/s/{p:path}", h, []testRouteRequest{
			{"GET", "", "/s/that/is/wildcard", siris.StatusOK, same_as_request_path},
			{"GET", "", "/s/ok", siris.StatusOK, same_as_request_path},
		}},
		{"GET", "/som/{p:path}", h, []testRouteRequest{
			{"GET", "", "/som/that/is/wildcard", siris.StatusOK, same_as_request_path},
			{"GET", "", "/som/ok", siris.StatusOK, same_as_request_path},
		}},
		{"GET", "/some/{p:path}", h, []testRouteRequest{
			{"GET", "", "/some/that/is/wildcard", siris.StatusOK, same_as_request_path},
			{"GET", "", "/some1/that/is/wildcard", siris.StatusNotFound, from_status_code},
		}},
	}

	testTheRoutes(t, tt, false)
}

func TestRouterWildcardAndStatic(t *testing.T) {
	var tt = []testRoute{
		{"GET", "/some/{p:path}", h2, []testRouteRequest{
			{"GET", "", "/some/that/is/wildcard", siris.StatusForbidden, same_as_request_path},
			{"GET", "", "/some/did", siris.StatusForbidden, same_as_request_path},
			{"GET", "", "/some1/that/is/wildcard", siris.StatusNotFound, from_status_code},
		}},
		{"GET", "/some/static", h, []testRouteRequest{
			{"GET", "", "/some/static", siris.StatusOK, same_as_request_path},
		}},
	}

	testTheRoutes(t, tt, false)
}

func TestRouterWildcardRootMany(t *testing.T) {
	var tt = []testRoute{
		// all routes will be handlded by "h" because we added wildcard to root,
		// this feature is very important and can remove noumerous of previous hacks on our apps.
		{"GET", "/{p:path}", h, []testRouteRequest{
			{"GET", "", "/this/is/wildcard/on/root", siris.StatusOK, same_as_request_path},
		}}, // mormally, order matters, root should be registered at last
		// but we change the front level order algorithm to put last these automatically
		// see handler.go
		{"GET", "/some/{p:path}", h2, []testRouteRequest{
			{"GET", "", "/some/that/is/wildcard", siris.StatusForbidden, same_as_request_path},
			{"GET", "", "/some/did", siris.StatusForbidden, same_as_request_path},
		}},
		{"GET", "/some/static", h, []testRouteRequest{
			{"GET", "", "/some/static", siris.StatusOK, same_as_request_path},
		}},
		{"GET", "/some1", h, []testRouteRequest{
			{"GET", "", "/some1", siris.StatusOK, same_as_request_path},
			// this will show up because of the first wildcard, as we wanted to do.
			{"GET", "", "/some1/that/is/wildcard", siris.StatusOK, same_as_request_path},
		}},
	}

	testTheRoutes(t, tt, false)
}

func TestRouterWildcardRootManyAndRootStatic(t *testing.T) {
	var tt = []testRoute{
		// all routes will be handlded by "h" because we added wildcard to root,
		// this feature is very important and can remove noumerous of previous hacks on our apps.
		{"GET", "/{p:path}", h, []testRouteRequest{
			{"GET", "", "/other2almost/some", siris.StatusOK, same_as_request_path},
			// it's a request to /other , not other/something, therefore the root wildcard is the handler
			{"GET", "", "/other", siris.StatusOK, same_as_request_path},
		}},
		{"GET", "/", h, []testRouteRequest{
			{"GET", "", "/", siris.StatusOK, same_as_request_path},
		}},
		{"GET", "/other/{paramother:path}", h2, []testRouteRequest{
			{"GET", "", "/other/wildcard", siris.StatusForbidden, same_as_request_path},
			{"GET", "", "/other/wildcard/here", siris.StatusForbidden, same_as_request_path},
		}},
		{"GET", "/other2/{paramothersecond:path}", h2, []testRouteRequest{
			{"GET", "", "/other2/wildcard", siris.StatusForbidden, same_as_request_path},
			{"GET", "", "/other2/more/than/one/path/parts", siris.StatusForbidden, same_as_request_path},
		}},
		{"GET", "/other2/static", h3, []testRouteRequest{
			{"GET", "", "/other2/static", siris.StatusOK, prefix_static_path_following_by_request_path},
			{"GET", "", "/other2/staticed", siris.StatusForbidden, same_as_request_path},
		}},
	}

	testTheRoutes(t, tt, false)
}

func testTheRoutes(t *testing.T, tests []testRoute, debug bool) {
	// build the api
	app := siris.New()
	for _, tt := range tests {
		app.Handle(tt.method, tt.path, tt.handler)
	}

	// setup the test suite
	e := httptest.New(t, app, httptest.Debug(debug))

	// run the tests
	for _, tt := range tests {
		for _, req := range tt.requests {
			method := req.method
			if method == "" {
				method = tt.method
			}
			ex := e.Request(method, req.path)
			if req.subdomain != "" {
				ex.WithURL("http://" + req.subdomain + ".localhost:8080")
			}

			expectedBody := req.expectedBody
			if req.expectedBody == same_as_request_path {
				expectedBody = req.path
			}
			if req.expectedBody == from_status_code {
				expectedBody = http.StatusText(req.expectedStatusCode)
			}
			if req.expectedBody == prefix_static_path_following_by_request_path {
				expectedBody = staticPathPrefixBody + req.path
			}

			ex.Expect().Status(req.expectedStatusCode).Body().Equal(expectedBody)
		}
	}
}
