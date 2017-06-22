// black-box testing
package host_test

import (
	"net"
	"net/url"
	"testing"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/core/host"
	"github.com/go-siris/siris/httptest"
)

func TestProxy(t *testing.T) {
	expectedIndex := "ok /"
	expectedAbout := "ok /about"
	unexpectedRoute := "unexpected"

	// proxySrv := siris.New()
	u, err := url.Parse("https://localhost:4444")
	if err != nil {
		t.Fatalf("%v while parsing url", err)
	}

	// p := host.ProxyHandler(u)
	// transport := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// p.Transport = transport
	// proxySrv.Downgrade(p.ServeHTTP)
	// go proxySrv.Run(siris.Addr(":80"), siris.WithoutBanner, siris.WithoutInterruptHandler)

	go host.NewProxy("localhost:2017", u).ListenAndServe() // should be localhost/127.0.0.1:80 but travis throws permission denied.

	app := siris.New()
	app.Get("/", func(ctx context.Context) {
		ctx.WriteString(expectedIndex)
	})

	app.Get("/about", func(ctx context.Context) {
		ctx.WriteString(expectedAbout)
	})

	app.OnErrorCode(siris.StatusNotFound, func(ctx context.Context) {
		ctx.WriteString(unexpectedRoute)
	})

	l, err := net.Listen("tcp", "localhost:4444") // should be localhost/127.0.0.1:443 but travis throws permission denied.
	if err != nil {
		t.Fatalf("%v while creating tcp4 listener for new tls local test listener", err)
	}
	// main server
	go app.Run(siris.Listener(httptest.NewLocalTLSListener(l)), siris.WithoutBanner)

	e := httptest.NewInsecure(t, httptest.URL("http://localhost:2017"))
	e.GET("/").Expect().Status(siris.StatusOK).Body().Equal(expectedIndex)
	e.GET("/about").Expect().Status(siris.StatusOK).Body().Equal(expectedAbout)
	e.GET("/notfound").Expect().Status(siris.StatusNotFound).Body().Equal(unexpectedRoute)
}
