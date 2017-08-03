// black-box testing
package coverageTests

import (
	stdContext "context"
	"crypto/tls"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"

	"gopkg.in/gavv/httpexpect.v1"
)

const (
	debug = false
)

var (
	rwmu = sync.RWMutex{}
)

func newTester(t *testing.T, baseURL string, handler http.Handler) *httpexpect.Expect {

	var transporter http.RoundTripper

	if strings.HasPrefix(baseURL, "http") { // means we are testing real serve time
		transporter = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else { // means we are testing the handler itself
		transporter = httpexpect.NewBinder(handler)
	}

	testConfiguration := httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: transporter,
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
	}

	if debug {
		testConfiguration.Printers = []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		}
	}

	return httpexpect.WithConfig(testConfiguration)
}

func TestSiris(t *testing.T) {
	app := siris.New()
	e := newTester(t, "https://127.0.0.1:7444", nil)

	expectedFoundResponse := "body{font-size: 30px}\n"

	app.StaticWeb("/static1", "./fixtures/static")
	app.StaticServe("./fixtures/static", "/static2")
	app.StaticContent("/static3", "text/css", []byte(expectedFoundResponse))

	app.Get("/get-siris", func(ctx context.Context) {
		ctx.Text("siris")
	})

	go app.Run(siris.TLS("127.0.0.1:7444", "./fixtures/server.crt", "./fixtures/server.key"))
	defer app.Shutdown(stdContext.TODO())

	time.Sleep(time.Duration(10 * time.Second))

	e.GET("/static1/").Expect().Status(siris.StatusNotFound)
	e.GET("/static1/test.css").Expect().Status(siris.StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/static2/").Expect().Status(siris.StatusNotFound)
	e.GET("/static2/test.css").Expect().Status(siris.StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/static3").Expect().Status(siris.StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/get-siris").Expect().Status(siris.StatusOK)

	e.GET("/notfound").Expect().Status(siris.StatusNotFound)

}
