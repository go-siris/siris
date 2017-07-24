// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package siris

import (
	// std packages
	stdContext "context"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	//logger
	//"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	// context for the handlers
	"github.com/go-siris/siris/context"
	// core packages, needed to build the application
	"github.com/go-siris/siris/core/errors"
	"github.com/go-siris/siris/core/host"
	"github.com/go-siris/siris/core/nettools"
	"github.com/go-siris/siris/core/router"
	// sessions and view
	"github.com/go-siris/siris/sessions"
	"github.com/go-siris/siris/view"
	// middleware used in Default method
	requestLogger "github.com/go-siris/siris/middleware/logger"
	"github.com/go-siris/siris/middleware/recover"
)

const (
	banner = `
  _________.___ __________ .___   _________
 /   _____/|   |\______   \|   | /   _____/
 \_____  \ |   | |       _/|   | \_____  \
 /        \|   | |    |   \|   | /        \
/_______  /|___| |____|_  /|___|/_______  /
        \/              \/              \/
         the fastest webframework
`

	// Version is the current version number of the Siris Web framework.
	//
	// Look https://github.com/go-siris/siris#where-can-i-find-older-versions for older versions.
	Version = "7.3.4"
)

const (
	// MethodNone is a Virtual method
	// to store the "offline" routes.
	//
	// Conversion for router.MethodNone.
	MethodNone = router.MethodNone
	// NoLayout to disable layout for a particular template file
	// Conversion for view.NoLayout.
	NoLayout = view.NoLayout
)

var (
	// HTML view engine.
	// Conversion for the view.HTML.
	HTML = view.HTML
	// Django view engine.
	// Conversion for the view.Django.
	Django = view.Django
	// Handlebars view engine.
	// Conversion for the view.Handlebars.
	Handlebars = view.Handlebars
	// Pug view engine.
	// Conversion for the view.Pug.
	Pug = view.Pug
	// Amber view engine.
	// Conversion for the view.Amber.
	Amber = view.Amber
)

var (
	// LimitRequestBodySize is a middleware which sets a request body size limit
	// for all next handlers in the chain.
	LimitRequestBodySize = context.LimitRequestBodySize
)

// serverErrLogger allows us to use the zap.Logger as our http.Server ErrorLog
type serverErrLogger struct {
	log *zap.SugaredLogger
}

// Implement Write to log server errors using the zap logger
func (s serverErrLogger) Write(b []byte) (int, error) {
	s.log.Debug(string(b))
	return 0, nil
}

// Application is responsible to manage the state of the application.
// It contains and handles all the necessary parts to create a fast web server.
type Application struct {
	// routing embedded | exposing APIBuilder's and Router's public API.
	*router.APIBuilder
	*router.Router
	ContextPool *context.Pool

	// config contains the configuration fields
	// all fields defaults to something that is working, developers don't have to set it.
	config *Configuration

	// the logrus logger instance, defaults to "Info" level messages (all except "Debug")
	logger *zap.SugaredLogger

	// view engine
	view view.View

	// sessions messages
	sessions *sessions.Manager

	// used for build
	once sync.Once

	mu sync.Mutex

	// Hosts contains a list of all servers (Host Supervisors) that this app is running on.
	//
	// Hosts may be empty only if application ran(`app.Run`) with `siris.Raw` option runner,
	// otherwise it contains a single host (`app.Hosts[0]`).
	//
	// Additional Host Supervisors can be added to that list by calling the `app.NewHost` manually.
	//
	// Hosts field is available after `Run` or `NewHost`.
	Hosts []*host.Supervisor
}

// New creates and returns a fresh empty Siris *Application instance.
func New() *Application {
	config := DefaultConfiguration()
	logger, _ := zap.NewDevelopmentConfig().Build()

	app := &Application{
		config:     &config,
		logger:     logger.Sugar(),
		APIBuilder: router.NewAPIBuilder(),
		Router:     router.NewRouter(),
	}

	app.ContextPool = context.New(func() context.Context {
		return context.NewContext(app)
	})

	return app
}

// Default returns a new Application instance.
// Unlike `New` this method prepares some things for you.
// std html templates from the "./templates" directory,
// session manager is attached with a default expiration of 7 days,
// recovery and (request) logger handlers(middleware) are being registered.
func Default() *Application {
	app := New()

	_ = app.AttachView(view.HTML("./templates", ".html"))
	app.AttachSessionManager("memory", &sessions.ManagerConfig{
		CookieName:      "go-session-id",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     7200,
	})

	app.Use(recover.New())
	app.Use(requestLogger.New())

	return app
}

// Configure can called when modifications to the framework instance needed.
// It accepts the framework instance
// and returns an error which if it's not nil it's printed to the logger.
// See configuration.go for more.
//
// Returns itself in order to be used like app:= New().Configure(...)
func (app *Application) Configure(configurators ...Configurator) *Application {
	for _, cfg := range configurators {
		cfg(app)
	}

	return app
}

// These are the different logging levels. You can set the logging level to log
// on the application 's instance of logger, obtained with `app.Logger()`.
//
// These are conversions from logrus.
const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zap.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zap.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zap.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zap.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zap.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zap.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zap.FatalLevel
)

// Logger returns the zap logger instance(pointer) that is being used inside the "app".
func (app *Application) Logger() *zap.SugaredLogger {
	return app.logger
}

// Build sets up, once, the framework.
// It builds the default router with its default macros
// and the template functions that are very-closed to siris.
func (app *Application) Build() (err error) {
	app.once.Do(func() {
		// view engine
		// here is where we declare the closed-relative framework functions.
		// Each engine has their defaults, i.e yield,render,render_r,partial, params...
		rv := router.NewRoutePathReverser(app.APIBuilder)
		app.view.AddFunc("urlpath", rv.Path)
		// app.view.AddFunc("url", rv.URL)
		err = app.view.Load()
		if err != nil {
			return // if view engine loading failed then don't continue
		}

		if !app.Router.Downgraded() {
			// router
			// create the request handler, the default routing handler
			var routerHandler = router.NewDefaultHandler()

			err = app.Router.BuildRouter(app.ContextPool, routerHandler, app.APIBuilder)
			// re-build of the router from outside can be done with;
			// app.RefreshRouter()
		}

	})

	return
}

// NewHost accepts a standar *http.Server object,
// completes the necessary missing parts of that "srv"
// and returns a new, ready-to-use, host (supervisor).
func (app *Application) NewHost(srv *http.Server) *host.Supervisor {
	app.mu.Lock()
	defer app.mu.Unlock()

	// set the server's handler to the framework's router
	if srv.Handler == nil {
		srv.Handler = app.Router
	}

	// check if different ErrorLog provided, if not bind it with the framework's logger
	if srv.ErrorLog == nil {
		srv.ErrorLog = log.New(serverErrLogger{app.Logger()}, "[HTTP Server] ", 0)
	}

	if srv.Addr == "" {
		srv.Addr = ":8080"
	}

	// create the new host supervisor
	// bind the constructed server and return it
	su := host.New(srv, app.config.EnableReuseport)

	if app.config.vhost == "" { // vhost now is useful for router subdomain on wildcard subdomains,
		// in order to correct decide what to do on:
		// mydomain.com -> invalid
		// localhost -> invalid
		// sub.mydomain.com -> valid
		// sub.localhost -> valid
		// we need the host (without port if 80 or 443) in order to validate these, so:
		app.config.vhost = nettools.ResolveVHost(srv.Addr)
	}

	if !app.config.DisableBanner {
		// show the banner and the available keys to exit from app.
		su.RegisterOnServeHook(host.WriteStartupLogOnServe(app.Logger(), banner+"V"+Version))
	}

	// the below schedules some tasks that will run among the server

	if !app.config.DisableInterruptHandler {
		// when CTRL+C/CMD+C pressed.
		shutdownTimeout := 5 * time.Second
		host.RegisterOnInterruptHook(host.ShutdownOnInterrupt(su, shutdownTimeout))
	}

	app.Hosts = append(app.Hosts, su)

	return su
}

// RegisterOnInterruptHook registers a global function to call when CTRL+C/CMD+C pressed or a unix kill command received.
//
// A shortcut for the `host#RegisterOnInterrupt`.
var RegisterOnInterruptHook = host.RegisterOnInterruptHook

// Shutdown gracefully terminates all the application's server hosts.
// Returns an error on the first failure, otherwise nil.
func (app *Application) Shutdown(ctx stdContext.Context) error {
	for _, su := range app.Hosts {
		if err := su.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Runner is just an interface which accepts the framework instance
// and returns an error.
//
// It can be used to register a custom runner with `Run` in order
// to set the framework's server listen action.
//
// Currently Runner is being used to declare the built'n server listeners.
//
// See `Run` for more.
type Runner func(*Application) error

// Listener can be used as an argument for the `Run` method.
// It can start a server with a custom net.Listener via server's `Serve`.
//
// See `Run` for more.
func Listener(l net.Listener) Runner {
	return func(app *Application) error {
		app.config.vhost = nettools.ResolveVHost(l.Addr().String())
		return app.NewHost(new(http.Server)).
			Serve(l)
	}
}

// Server can be used as an argument for the `Run` method.
// It can start a server with a *http.Server.
//
// See `Run` for more.
func Server(srv *http.Server) Runner {
	return func(app *Application) error {
		return app.NewHost(srv).
			ListenAndServe()
	}
}

// Addr can be used as an argument for the `Run` method.
// It accepts a host address which is used to build a server
// and a listener which listens on that host and port.
//
// Addr should have the form of [host]:port, i.e localhost:8080 or :8080.
//
// See `Run` for more.
func Addr(addr string) Runner {
	return func(app *Application) error {
		return app.NewHost(&http.Server{Addr: addr}).
			ListenAndServe()
	}
}

// TLS can be used as an argument for the `Run` method.
// It will start the Application's secure server.
//
// Use it like you used to use the http.ListenAndServeTLS function.
//
// Addr should have the form of [host]:port, i.e localhost:443 or :443.
// CertFile & KeyFile should be filenames with their extensions.
//
// See `Run` for more.
func TLS(addr string, certFile, keyFile string) Runner {
	return func(app *Application) error {
		return app.NewHost(&http.Server{Addr: addr}).
			ListenAndServeTLS(certFile, keyFile)
	}
}

// AutoTLS can be used as an argument for the `Run` method.
// It will start the Application's secure server using
// certifications created on the fly by the "autocert" golang/x package,
// so localhost may not be working, use it at "production" machine.
//
// Addr should have the form of [host]:port, i.e mydomain.com:443.
//
// See `Run` for more.
func AutoTLS(addr string) Runner {
	return func(app *Application) error {
		return app.NewHost(&http.Server{Addr: addr}).
			ListenAndServeAutoTLS()
	}
}

// Raw can be used as an argument for the `Run` method.
// It accepts any (listen) function that returns an error,
// this function should be block and return an error
// only when the server exited or a fatal error caused.
//
// With this option you're not limited to the servers
// that Siris can run by-default.
//
// See `Run` for more.
func Raw(f func() error) Runner {
	return func(*Application) error {
		return f()
	}
}

// ErrServerClosed is returned by the Server's Serve, ServeTLS, ListenAndServe,
// and ListenAndServeTLS methods after a call to Shutdown or Close.
//
// Conversion for the http.ErrServerClosed.
var ErrServerClosed = http.ErrServerClosed

// Run builds the framework and starts the desired `Runner` with or without configuration edits.
//
// Run should be called only once per Application instance, it blocks like http.Server.
//
// If more than one server needed to run on the same siris instance
// then create a new host and run it manually by `go NewHost(*http.Server).Serve/ListenAndServe` etc...
// or use an already created host:
// h := NewHost(*http.Server)
// Run(Raw(h.ListenAndServe), WithoutBanner, WithCharset("UTF-8"))
//
// The Application can go online with any type of server or siris's host with the help of
// the following runners:
// `Listener`, `Server`, `Addr`, `TLS`, `AutoTLS` and `Raw`.
func (app *Application) Run(serve Runner, withOrWithout ...Configurator) error {
	// first Build because it doesn't need anything from configuration,
	//  this give the user the chance to modify the router inside a configurator as well.
	if err := app.Build(); err != nil {
		return errors.PrintAndReturnErrors(err, app.logger.Errorf)
	}

	app.Configure(withOrWithout...)

	// this will block until an error(unless supervisor's DeferFlow called from a Task).
	err := serve(app)
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		app.Logger().Error(err)
	}
	return err
}

// AttachView should be used to register view engines mapping to a root directory
// and the template file(s) extension.
// Returns an error on failure, otherwise nil.
func (app *Application) AttachView(viewEngine view.Engine) error {
	return app.view.Register(viewEngine)
}

// View executes and writes the result of a template file to the writer.
//
// First parameter is the writer to write the parsed template.
// Second parameter is the relative, to templates directory, template filename, including extension.
// Third parameter is the layout, can be empty string.
// Forth parameter is the bindable data to the template, can be nil.
//
// Use context.View to render templates to the client instead.
// Returns an error on failure, otherwise nil.
func (app *Application) View(writer io.Writer, filename string, layout string, bindingData interface{}) error {
	if app.view.Len() == 0 {
		err := errors.New("view engine is missing, use `AttachView`")
		app.Logger().Error(err)
		return err
	}
	err := app.view.ExecuteWriter(writer, filename, layout, bindingData)
	if err != nil {
		app.Logger().Error(err)
	}
	return err
}

// AttachSessionManager registers a session manager to the framework which is used for flash messages too.
//
// See context.Session too.
func (app *Application) AttachSessionManager(provider string, cfg *sessions.ManagerConfig) {
	manager, err := sessions.NewManager(provider, cfg)
	if err != nil {
		return
	}
	app.sessions = manager
	go app.sessions.GC()
	app.Done(func(ctx context.Context) {
		ctx.Session().SessionRelease(ctx.ResponseWriter())
	})
}

// SessionManager returns the session manager which contain a Start and Destroy methods
// used inside the context.Session().
//
// It's ready to use after the RegisterSessions.
func (app *Application) SessionManager() (*sessions.Manager, error) {
	var sessions *sessions.Manager
	if app.sessions == nil {
		return sessions, errors.New("session manager is missing")
	}
	return app.sessions, nil
}

// SPA  accepts an "assetHandler" which can be the result of an
// app.StaticHandler or app.StaticEmbeddedHandler.
// It wraps the router and checks:
// if it;s an asset, if the request contains "." (this behavior can be changed via /core/router.NewSPABuilder),
// if the request is index, redirects back to the "/" in order to let the root handler to be executed,
// if it's not an asset then it executes the router, so the rest of registered routes can be
// executed without conflicts with the file server ("assetHandler").
//
// Use that instead of `StaticWeb` for root "/" file server.
//
// Example: https://github.com/go-siris/siris/tree/master/_examples/beginner/file-server/single-page-application
func (app *Application) SPA(assetHandler context.Handler) {
	s := router.NewSPABuilder(assetHandler)
	wrapper := s.BuildWrapper(app.ContextPool)
	app.Router.WrapRouter(wrapper)
}

// ConfigurationReadOnly returns a structure which doesn't allow writing.
func (app *Application) ConfigurationReadOnly() context.ConfigurationReadOnly {
	return app.config
}
