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
	"github.com/go-siris/siris/core/errors"
	"github.com/go-siris/siris/core/router"
)

const (
	debug                  = true
	expectedFoundResponse  = "body{font-size: 30px}\n"
	expectedFoundResponse2 = "1234test4321"
)

var (
	rwmu           = sync.RWMutex{}
	args           []string
	runLetsEncrypt bool
	sslCert        string
	sslKey         string
	staticDir      string
)

func TestSiris(t *testing.T) {
	initial()

	app := New()

	app.OnAnyErrorCode(func(ctx context.Context) {
		fmt.Printf("OnAnyErrorCode: %d %#v\n\n", ctx.GetStatusCode(), ctx.Err())
		errors.PrintAndReturnErrors(ctx.Err(), app.Logger().Infof)
		ctx.Next()
	})

	app.Favicon(staticDir)

	app.Get("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Post("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Delete("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Put("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Head("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Options("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Patch("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Connect("/", func(ctx context.Context) {
		ctx.Text("hi")
	})
	app.Trace("/", func(ctx context.Context) {
		ctx.Text("hi")
	})

	party := app.Party("/party")
	{
		party.Get("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Post("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Delete("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Put("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Head("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Options("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Patch("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Connect("/", func(ctx context.Context) {
			ctx.Text("party")
		})
		party.Trace("/", func(ctx context.Context) {
			ctx.Text("party")
		})
	}

	app.PartyFunc("/partyfunc", func(u router.Party) {
		u.Get("/", func(ctx context.Context) {
			ctx.Text("partyfunc")
		})
		u.Post("/", func(ctx context.Context) {
			ctx.Text("partyfunc")
		})
		u.Delete("/", func(ctx context.Context) {
			ctx.Text("partyfunc")
		})
	})

	app.Get("/cache", cache.WrapHandler(func(ctx context.Context) {
		ctx.Text("hi cache " + time.Now().String())
	}, 2*time.Minute))

	// file in staticDir blob.css
	app.StaticWeb("/static1/", staticDir)
	app.StaticServe(staticDir, "/static2/")
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

	go app.Run(Addr("127.0.0.1:9080"), WithoutBanner, WithoutInterruptHandler)

	if runLetsEncrypt {
		go app.Run(AutoTLS("127.0.0.1:443"), WithoutBanner, WithoutInterruptHandler)
	} else {
		go app.Run(TLS("127.0.0.1:9443", sslCert, sslKey), WithoutBanner, WithoutInterruptHandler)
	}
	defer app.Shutdown(stdContext.TODO())

	time.Sleep(time.Duration(5 * time.Second))

	for _, s := range app.Hosts {
		s.RegisterOnErrorHook(func(err error) {
			fmt.Println("RegisterOnErrorHook: " + err.Error())
		})
	}
	time.Sleep(time.Duration(5 * time.Second))

	client := newTester(t, "http://127.0.0.1:9080", nil)
	testClient(client)
	clientTls := newTester(t, "https://127.0.0.1:9443", nil)
	testClient(clientTls)

	time.Sleep(time.Duration(5 * time.Second))
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

func testClient(e *httpexpect.Expect) {
	e.GET("/static1/").Expect().Status(StatusNotFound)
	e.GET("/static1/test.css").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse)
	e.GET("/static1/demo/test.txt").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse2)

	e.GET("/static2/").Expect().Status(StatusNotFound)
	e.GET("/static2/test.css").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse)
	e.GET("/static2/demo/test.txt").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse2)

	e.GET("/static3").Expect().Status(StatusOK).
		Body().Equal(expectedFoundResponse)

	e.GET("/get-siris").Expect().Status(StatusOK)

	e.GET("/cache").Expect().Status(StatusOK)
	time.Sleep(time.Duration(1 * time.Second))
	e.GET("/cache").Expect().Status(StatusOK)
	time.Sleep(time.Duration(1 * time.Second))

	e.GET("/").Expect().Status(StatusOK)
	e.POST("/").Expect().Status(StatusOK)
	e.PUT("/").Expect().Status(StatusOK)
	e.DELETE("/").Expect().Status(StatusOK)
	e.OPTIONS("/").Expect().Status(StatusOK)
	e.HEAD("/").Expect().Status(StatusOK)
	e.PATCH("/").Expect().Status(StatusOK)

	e.GET("/party/").Expect().Status(StatusOK)
	e.POST("/party/").Expect().Status(StatusOK)
	e.PUT("/party/").Expect().Status(StatusOK)
	e.DELETE("/party/").Expect().Status(StatusOK)
	e.OPTIONS("/party/").Expect().Status(StatusOK)
	e.HEAD("/party/").Expect().Status(StatusOK)
	e.PATCH("/party/").Expect().Status(StatusOK)

	e.GET("/notfound").Expect().Status(StatusNotFound)
	e.GET("/favicon.ico").Expect().Status(StatusOK)
}
