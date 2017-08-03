// +build testsiris

package siris

import (
	stdContext "context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"gopkg.in/gavv/httpexpect.v1"

	"github.com/go-siris/siris/cache"
	"github.com/go-siris/siris/context"
)

const (
	debug                 = true
	expectedFoundResponse = "body{font-size: 30px}\n"
)

var (
	rwmu           = sync.RWMutex{}
	args           []string
	runLetsEncrypt bool
	sslCert        string
	sslKey         string
	staticDir      string
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
	initial()

	clientTls := newTester(t, "https://pp.derhost.com", nil)
	client := newTester(t, "http://pp.derhost.com", nil)

	app := New()
	app.Get("/", func(ctx context.Context) {
		ctx.Text("hi")
	})

	app.Get("/cache", cache.WrapHandler(func(ctx context.Context) {
		ctx.Text("hi")
	}, 2*time.Minute))

	app.StaticWeb("/static1", staticDir)
	app.StaticServe(staticDir, "/static2")
	app.StaticContent("/static3", "text/css", []byte(expectedFoundResponse))

	app.Get("/get-siris", func(ctx context.Context) {
		ctx.Text("siris")
	})

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			// kill -SIGINT XXXX or Ctrl+c
			os.Interrupt,
			syscall.SIGINT, // register that too, it should be ok
			// os.Kill  is equivalent with the syscall.Kill
			os.Kill,
			syscall.SIGKILL, // register that too, it should be ok
			// kill -SIGTERM XXXX
			syscall.SIGTERM,
		)
		select {
		case <-ch:
			println("shutdown...")

			timeout := 5 * time.Second
			ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
			defer cancel()
			app.Shutdown(ctx)
		}
	}()

	go app.Run(Addr("pp.derhost.com:80"), WithoutBanner, WithoutInterruptHandler)
	if runLetsEncrypt {
		go app.Run(AutoTLS("pp.derhost.com:443"), WithoutBanner, WithoutInterruptHandler)
	} else {
		go app.Run(TLS("pp.derhost.com:443", sslCert, sslKey), WithoutBanner, WithoutInterruptHandler)
	}
	defer app.Shutdown(stdContext.TODO())

	time.Sleep(time.Duration(10 * time.Second))

	testClient(client)
	testClient(clientTls)
}

func initial() {
	for _, arg := range os.Args {
		switch {
		case arg == "--letsencrypt":
			runLetsEncrypt = true
		default:
			args = append(args, arg)
		}
	}
	os.Args = args

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	sslCert = path.Join(dir, "coverage-tests", "fixtures", "server.crt")
	sslKey = path.Join(dir, "coverage-tests", "fixtures", "server.key")
	staticDir = path.Join(dir, "coverage-tests", "fixtures", "static")

	fmt.Println(staticDir)
}

func testClient(e *httpexpect.Expect) {
	e.GET("/static1/").Expect().Status(StatusNotFound)
	e.GET("/static1/test.css").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/static2/").Expect().Status(StatusNotFound)
	e.GET("/static2/test.css").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/static3").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/get-siris").Expect().Status(StatusOK)

	e.GET("/cache").Expect().Status(StatusOK)
	e.GET("/cache").Expect().Status(StatusOK)

	e.GET("/notfound").Expect().Status(StatusNotFound)
}
