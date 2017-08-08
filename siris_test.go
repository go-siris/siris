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
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"gopkg.in/gavv/httpexpect.v1"

	"github.com/go-siris/siris/cache"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/core/errors"
	"github.com/go-siris/siris/core/host"
	"github.com/go-siris/siris/core/nettools"
	"github.com/go-siris/siris/core/router"
	"github.com/go-siris/siris/view"
)

const (
	debug                  = false
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

	app := Default()
	app.Configure(WithJSONInteratorReplacement, EnableReuseport, EnableQUICSupport, WithTimeFormat(time.RFC3339), WithCharset("UTF-8"), WithRemoteAddrHeader("X-Real-Ip"), WithoutRemoteAddrHeader("X-No-Real-Ip"), WithOtherValue("AppName", "SIRIS"))
	app.AttachView(view.HTML("./", ".html").Binary(get_files, get_names))

	app.OnErrorCode(StatusPaymentRequired, func(ctx context.Context) {
		ctx.IsAjax()
		ctx.Application()
		ctx.ClientSupportsGzip()
		ctx.GetStatusCode()
		ctx.GetContentType()
		ctx.GetHeader("User-Agent")
		ctx.GetCookie("lang")
		ctx.SetCookieKV("lang", "value")
		ctx.SetMaxRequestBodySize(1234567890)
	})

	app.OnAnyErrorCode(func(ctx context.Context) {
		fmt.Printf("OnAnyErrorCode: %d %#v\n\n", ctx.GetStatusCode(), ctx.Err())
		errors.PrintAndReturnErrors(ctx.Err(), app.Logger().Infof)
		ctx.Next()
	})

	app.Favicon(staticDir)
	app.Get("/view", context.LimitRequestBodySize(1234567890), func(ctx context.Context) {
		ctx.View("view/parent.html")
	})
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

	app.Any("/any", func(ctx context.Context) {
		ctx.Text("/any")
	})

	app.Many("/many", []string{MethodGet, MethodHead, MethodPost}, func(ctx context.Context) {
		ctx.Text("/many")
	})

	// you can use the "string" type which is valid for a single path parameter that can be anything.
	app.Get("/username/{name}", func(ctx context.Context) {
		ctx.Writef("Hello %s", ctx.Params().Get("name"))
	}) // type is missing = {name:string}

	// Let's register our first macro attached to int macro type.
	// "min" = the function
	// "minValue" = the argument of the function
	// func(string) bool = the macro's path parameter evaluator, this executes in serve time when
	// a user requests a path which contains the :int macro type with the min(...) macro parameter function.
	app.Macros().Int.RegisterFunc("min", func(minValue int) func(string) bool {
		// do anything before serve here [...]
		// at this case we don't need to do anything
		return func(paramValue string) bool {
			n, err := strconv.Atoi(paramValue)
			if err != nil {
				return false
			}
			return n >= minValue
		}
	})

	// http://go-siris.com:8080/profile/id>=1
	// this will throw 404 even if it's found as route on : /profile/0, /profile/blabla, /profile/-1
	// macro parameter functions are optional of course.
	app.Get("/profile/{id:int min(1)}", func(ctx context.Context) {
		// second parameter is the error but it will always nil because we use macros,
		// the validaton already happened.
		id, _ := ctx.Params().GetInt("id")
		ctx.Writef("Hello id: %d", id)
	})

	// to change the error code per route's macro evaluator:
	app.Get("/profile/{id:int min(1)}/friends/{friendid:int min(1) else 504}", func(ctx context.Context) {
		id, _ := ctx.Params().GetInt("id")
		friendid, _ := ctx.Params().GetInt("friendid")
		ctx.Writef("Hello id: %d looking for friend id: %d", id, friendid)
	}) // this will throw e 504 error code instead of 404 if all route's macros not passed.

	// http://go-siris.com:8080/game/a-zA-Z/level/0-9
	// remember, alphabetical is lowercase or uppercase letters only.
	app.Get("/game/{name:alphabetical}/level/{level:int}", func(ctx context.Context) {
		ctx.Writef("name: %s | level: %s", ctx.Params().Get("name"), ctx.Params().Get("level"))
	})

	app.Get("/lowercase/static", func(ctx context.Context) {
		ctx.Writef("static and dynamic paths are not conflicted anymore!")
	})

	// let's use a trivial custom regexp that validates a single path parameter
	// which its value is only lowercase letters.

	// http://go-siris.com:8080/lowercase/anylowercase
	app.Get("/lowercase/{name:string regexp(^[a-z]+)}", func(ctx context.Context) {
		ctx.Writef("name should be only lowercase, otherwise this handler will never executed: %s", ctx.Params().Get("name"))
	})

	// http://go-siris.com:8080/single_file/app.js
	app.Get("/single_file/{myfile:file}", func(ctx context.Context) {
		ctx.Writef("file type validates if the parameter value has a form of a file name, got: %s", ctx.Params().Get("myfile"))
	})

	// http://go-siris.com:8080/myfiles/any/directory/here/
	// this is the only macro type that accepts any number of path segments.
	app.Get("/myfiles/{directory:path}", func(ctx context.Context) {
		ctx.Writef("path type accepts any number of path segments, path after /myfiles/ is: %s", ctx.Params().Get("directory"))
	}) // for wildcard path (any number of path segments) without validation you can use:

	app.Get("/cache", cache.WrapHandler(func(ctx context.Context) {
		ctx.Text("hi cache " + time.Now().String())
	}, 2*time.Minute))

	// file in staticDir blob.css
	app.StaticWeb("/static1/", staticDir)
	app.StaticServe(staticDir, "/static2/")
	app.StaticContent("/static3", "text/css", []byte(expectedFoundResponse))
	app.StaticEmbedded("/static4/", "/", get_files, get_names)

	app.Get("/get-siris", func(ctx context.Context) {
		ctx.Text("siris")
	})

	app.Get("/get-payment", func(ctx context.Context) {
		ctx.StatusCode(StatusPaymentRequired)
	})

	app.Get("/get-gzip", func(ctx context.Context) {
		ctx.Gzip(ctx.ClientSupportsGzip())
		ctx.Text("Gzip")
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

	app.Build()
	app.GetRoutes()
	app.GetRoute("HEAD/")

	app.None("/none-active", func(ctx context.Context) {
		ctx.Text("none-active route")
	})

	loggerP := serverErrLogger{app.Logger()}
	loggerP.Write([]byte("Logger: Start Servers"))

	app.ConfigureHost(configureHosts)
	go app.Run(Addr("go-siris.com:9080"), WithoutBanner, WithoutInterruptHandler)

	if runLetsEncrypt {
		go app.Run(AutoTLS("go-siris.com:443"), WithoutBanner, WithoutInterruptHandler)
	} else {
		go app.Run(TLS("go-siris.com:9443", sslCert, sslKey), WithoutBanner, WithoutInterruptHandler)
	}

	unicl, err := nettools.UNIX("/tmp/srv.sock", 0777) // see its code to see how you can manually create a new file listener, it's easy.
	if err != nil {
		panic(err)
	}

	go app.Run(Listener(unicl))

	loggerP.Write([]byte("Logger: Servers started"))

	time.Sleep(time.Duration(5 * time.Second))

	for _, s := range app.Hosts {
		s.RegisterOnErrorHook(func(err error) {
			fmt.Println("RegisterOnErrorHook: " + err.Error())
		})
	}

	time.Sleep(time.Duration(5 * time.Second))

	client := newTester(t, "http://go-siris.com:9080", nil)
	testClient(client)
	clientTls := newTester(t, "https://go-siris.com:9443", nil)
	testClient(clientTls)

	time.Sleep(time.Duration(5 * time.Second))

	app.Shutdown(stdContext.TODO())

	time.Sleep(time.Duration(5 * time.Second))
}

func configureHosts(su *host.Supervisor) {
	// here we have full access to the host that will be created
	// inside the `app.Run` or `app.NewHost` function .
	//
	// we're registering a shutdown "event" callback here:
	su.RegisterOnShutdownHook(func() {
		println("server is closed")
	})
	su.RegisterOnErrorHook(func(err error) {
		println("error:" + err.Error())
	})
	su.RegisterOnServeHook(func(th host.TaskHost) {
		println("Server Started")
		println(th.HostURL())
		println(th.Hostname())
	})
	// su.RegisterOnError
	// su.RegisterOnServe
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

var files2017 = map[string]string{
	"views/parent.html": `<!DOCTYPE html>
		<html>
			<head>
				<title>Iris with Pongo2 template engine</title>
					<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
					<meta http-equiv="Pragma" content="no-cache">
					<meta http-equiv="cache-control" content="no-cache">
			</head>
			<body>
				<h1>demo</h1>
			</body>
		</html>`,

	"css/css.css": "body{}",
}

func get_names() []string {
	var res []string

	for name := range files2017 {
		res = append(res, name)
	}

	return res
}

func get_files(name string) ([]byte, error) {
	content, exists := files2017[name]
	if !exists {
		return nil, os.ErrNotExist
	}

	return []byte(content), nil
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

	e.GET("/static4/").Expect().Status(StatusNotFound)
	bytes, _ := get_files("css/css.css")
	e.GET("/static4/css/css.css").Expect().Status(StatusOK).
		Body().Equal(string(bytes))
	bytes, _ = get_files("views/parent.html")
	e.GET("/static4/views/parent.html").Expect().Status(StatusOK).
		Body().Equal(string(bytes))

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

	e.GET("/username/Bob").Expect().Status(StatusOK).
		Body().Equal("Hello Bob")

	e.GET("/profile/1029").
		Expect().Status(StatusOK).
		Body().Equal("Hello id: 1029")
	e.GET("/profile/0").
		Expect().Status(StatusNotFound)

	e.GET("/profile/1029/friends/4321").
		Expect().Status(StatusOK).
		Body().Equal("Hello id: 1029 looking for friend id: 4321")
	e.GET("/profile/1029/friends/0").
		Expect().Status(StatusGatewayTimeout)

	e.GET("/game/Siris/level/302").
		Expect().Status(StatusOK).
		Body().Equal("name: Siris | level: 302")

	e.GET("/lowercase/static").
		Expect().Status(StatusOK).
		Body().Equal("static and dynamic paths are not conflicted anymore!")

	e.GET("/lowercase/staticlow").
		Expect().Status(StatusOK).
		Body().Equal("name should be only lowercase, otherwise this handler will never executed: staticlow")

	e.GET("/lowercase/killer").
		Expect().Status(StatusOK).
		Body().Equal("name should be only lowercase, otherwise this handler will never executed: killer")

	e.GET("/single_file/whatthehell.zip").
		Expect().Status(StatusOK).
		Body().Equal("file type validates if the parameter value has a form of a file name, got: whatthehell.zip")

	e.GET("/myfiles/whatthehell/get/me").
		Expect().Status(StatusOK).
		Body().Equal("path type accepts any number of path segments, path after /myfiles/ is: whatthehell/get/me")

	e.GET("/notfound").Expect().Status(StatusNotFound)
	e.GET("/favicon.ico").Expect().Status(StatusOK)

	e.GET("/get-payment").Expect().Status(StatusPaymentRequired)
	e.GET("/get-gzip").Expect().Status(StatusOK)
}
