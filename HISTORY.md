# FAQ

### Looking for free support?

	http://support.go-siris.com
    https://gitter.im/gosiris/siris

### Looking for previous versions?

    https://github.com/go-siris/siris#version


### Should I upgrade my Iris?

Developers are not forced to upgrade if they don't really need it. Upgrade whenever you feel ready.
> Siris uses the [vendor directory](https://docs.google.com/document/d/1Bz5-UB7g2uPBdOx-rw5t9MxJwkfpx90cqG9AFL0JAYo) feature, so you get truly reproducible builds, as this method guards against upstream renames and deletes.

**How to upgrade**: Open your command-line and execute this command: `go get -u github.com/go-siris/siris`. 
For further installation support, please click [here](http://support.go-siris.com/d/16-how-to-install-iris-web-framework).


### About our new home page
    http://go-siris.com

Thanks to [Santosh Anand](https://github.com/santoshanand) the http://go-siris.com has been upgraded and it's really awesome!

[Santosh](https://github.com/santoshanand) is a freelancer, he has a great knowledge of nodejs and express js, Android, iOS, React Native, Vue.js etc, if you need a developer to find or create a solution for your problem or task, please contact with him.


The amount of the next two or three donations you'll send they will be immediately transferred to his own account balance, so be generous please!

# Sa, 24 June 2017 | v7.3.4

- bugfix in router for a bug that occurs under crazy conditions
- bugfix with vendors build
- bugfix reuseport not compatible with windows (yet - maybe ever)

# Th, 22 June 2017 | v7.3.0

### SIRIS the community fork of Iris

Since the announcement that Iris was aquired, we createt a new fork "SIRIS". All Iris 7.x apps will be compatible the only change you need to do is use `import iris "gopkg.in/go-siris/siris.v7"`

Feel free, to report issues or create pull-requests. We'll be happy to integrate the community to go forward with this web framework.

### Reuseport

[Reuseport feature](https://github.com/go-siris/siris#performance-optimization-tips-for-multi-core-systems)

To use the new reuseport feature enable it via:

```GO
app.Run(siris.Addr(":8080"), siris.EnableReuseport)
// or before run:
app.Configure(siris.EnableReuseport)
```



# Th, 15 June 2017 | v7.2.0

### Cache

Declare the `Siris.Cache alias` to the new, improved and most-suited for common usage, `cache.Handler function`.

`Siris.Cache` be used as middleware in the chain now, example [here](_examples/intermediate/cache-markdown/main.go). However [you can still use the cache as a wrapper](cache/cache_test.go) by importing the `github.com/go-siris/siris/cache` package. 


### File server

- **Fix** [that](https://github.com/iris-contrib/community-board/issues/12).

- `app.StaticHandler(requestPath string, systemPath string, showList bool, gzip bool)` -> `app.StaticHandler(systemPath,showList bool, gzip bool)`

- **New** feature for Single Page Applications, `app.SPA(assetHandler context.Handler)` implemented.

- **New** `app.StaticEmbeddedHandler(vdir string, assetFn func(name string) ([]byte, error), namesFn func() []string)` added in order to be able to pass that on `app.SPA(app.StaticEmbeddedHandler("./public", Asset, AssetNames))`.

- **Fix** `app.StaticEmbedded(requestPath string, vdir string, assetFn func(name string) ([]byte, error), namesFn func() []string)`.

Examples: 
- [Embedding Files Into Executable App](_examples/beginner/file-server/embedding-files-into-app)
- [Single Page Application](_examples/beginner/file-server/single-page-application)
- [Embedding Single Page Application](_examples/beginner/file-server/embedding-single-page-application)

> [app.StaticWeb](_examples/beginner/file-server/basic/main.go) doesn't works for root request path "/"  anymore, use the new `app.SPA` instead.   

### WWW subdomain entry

- [Example](_examples/intermediate/subdomains/www/main.go) added to copy all application's routes, including parties, to the `www.mydomain.com`


### Wrapping the Router

- [Example](_examples/beginner/routing/custom-wrapper/main.go) added to show you how you can use the `app.WrapRouter` 
to implement a similar to `app.SPA` functionality, don't panic, it's easier than it sounds.


### Testing

- `httptest.New(app *siris.Application, t *testing.T)` -> `httptest.New(t *testing.T, app *siris.Application)`.

- **New** `httptest.NewLocalListener() net.Listener` added.
- **New** `httptest.NewLocalTLSListener(tcpListener net.Listener) net.Listener` added.

Useful for testing tls-enabled servers: 

Proxies are trying to understand local addresses in order to allow `InsecureSkipVerify`.

-  `host.ProxyHandler(target *url.URL) *httputil.ReverseProxy`.
-  `host.NewProxy(hostAddr string, target *url.URL) *Supervisor`.
        
    Tests [here](core/host/proxy_test.go).

# Tu, 13 June 2017 | v7.1.1

Fix [that](https://github.com/iris-contrib/community-board/issues/11).

# Mo, 12 June 2017 | v7.1.0

Fix [that](https://github.com/iris-contrib/community-board/issues/10).


# Su, 11 June 2017 | v7.0.5

Siris now supports static paths and dynamic paths for the same path prefix with zero performance cost:

`app.Get("/profile/{id:int}", handler)` and `app.Get("/profile/create", createHandler)` are not in conflict anymore.


The rest of the special Siris' routing features, including static & wildcard subdomains are still work like a charm.

> This was one of the most popular community's feature requests. Click [here](https://github.com/go-siris/siris/blob/master/_examples/beginner/routing/overview/main.go) to see a trivial example.

# Sa, 10 June 2017 | v7.0.4

- Simplify and add a test for the [basicauth middleware](https://github.com/go-siris/siris/tree/master/middleware/basicauth), no need to be
stored inside the Context anymore, developers can get the validated user(username and password) via `context.Request().BasicAuth()`. `basicauth.Config.ContextKey` was removed, just remove that field from your configuration, it's useless now. 

# Sa, 10 June 2017 | v7.0.3

- New `context.Session().PeekFlash("key")` added, unlike `GetFlash` this will return the flash value but keep the message valid for the next requests too.
- Complete the [httptest example](https://github.com/iris-contrib/examples/tree/master/httptest).
- Fix the (marked as deprecated) `ListenLETSENCRYPT` function.
- Upgrade the [iris-contrib/middleware](https://github.com/iris-contrib/middleware) including JWT, CORS and Secure handlers.
- Add [OAuth2 example](https://github.com/iris-contrib/examples/tree/master/oauth2) -- showcases the third-party package [goth](https://github.com/markbates/goth) integration with siris.

### Community

 - Add github integration on https://kataras.rocket.chat/channel/iris , so users can login with their github accounts instead of creating new for the chat only.

# Th, 08 June 2017 | v7.0.2

- Able to set **immutable** data on sessions and context's storage. Aligned to fix an issue on slices and maps as reported [here](https://github.com/iris-contrib/community-board/issues/5).

# We, 07 June 2017 | v7.0.1

- Proof of concept of an internal release generator, navigate [here](https://github.com/iris-contrib/community-board/issues/2) to read more. 
- Remove tray icon "feature", click [here](https://github.com/iris-contrib/community-board/issues/1) to learn why.

# Sa, 03 June 2017 

After 2+ months of hard work and collaborations, Siris [version 7](https://github.com/go-siris/siris) was published earlier today.

If you're new to Siris you don't have to read all these, just navigate to the [updated examples](https://github.com/go-siris/siris/tree/master/_examples) and you should be fine:)

Note that this section will not
cover the internal changes, the difference is so big that anybody can see them with a glimpse, even the code structure itself.


## Changes from [v6](https://github.com/go-siris/siris/tree/v6)

The whole framework was re-written from zero but I tried to keep the most common public API that iris developers use.

Vendoring /w update 

The previous vendor action for v6 was done by-hand, now I'm using the [go dep](https://github.com/golang/dep) tool, I had to do
some small steps:

- remove files like testdata to reduce the folder size
- rollback some of the "golang/x/net/ipv4" and "ipv6" source files because they are downloaded to their latest versions
by go dep, but they had lines with the `typealias` feature, which is not ready by current golang version (it will be on August)
- fix "cannot use internal package" at golang/x/net/ipv4 and ipv6 packages
	- rename the interal folder to was-internal, everywhere and fix its references.
- fix "main redeclared in this block"
	- remove all examples folders.
- remove main.go files on jsondiff lib, used by gavv/httpexpect, produces errors on `test -v ./...` while jd and jp folders are not used at all.

The go dep tool does what is says, as expected, don't be afraid of it now.
I am totally recommending this tool for package authors, even if it's in its alpha state.
I remember when Siris was in its alpha state and it had 4k stars on its first weeks/or month and that helped me a lot to fix reported bugs by users and make the framework even better, so give love to go dep from today!

General

- Several enhancements for the typescript transpiler, view engine, websocket server and sessions manager
- All `Listen` methods replaced with a single `Run` method, see [here](https://github.com/go-siris/siris/tree/master/_examples/beginner/listening)
- Configuration, easier to modify the defaults, see [here](https://github.com/go-siris/siris/tree/master/_examples/beginner/cofiguration)
- `HandlerFunc` removed, just `Handler` of `func(context.Context)` where context.Context derives from `import "github.com/go-siris/siris/context"` (on August this import path will be optional)
    - Simplify API, i.e: instead of `Handle,HandleFunc,Use,UseFunc,Done,DoneFunc,UseGlobal,UseGlobalFunc` use `Handle,Use,Done,UseGlobal`.
- Response time decreased even more (9-35%, depends on the application)
- The `Adaptors` idea replaced with a more structural design pattern, but you have to apply these changes: 
    - `app.Adapt(view.HTML/Pug/Amber/Django/Handlebars...)` -> `app.AttachView(view.HTML/Pug/Amber/Django/Handlebars...)` 
    - `app.Adapt(sessions.New(...))` -> `app.AttachSessionManager(sessions.New(...))`
    - `app.Adapt(siris.LoggerPolicy(...))` -> `app.AttachLogger(io.Writer)`
    - `app.Adapt(siris.RenderPolicy(...))` -> removed and replaced with the ability to replace the whole context with a custom one or override some methods of it, see below.

Routing
- Remove of multiple routers, now we have the fresh Siris router which is based on top of the julien's [httprouter](https://github.com/julienschmidt/httprouter).
    > Update 11 June 2017: As of 7.0.5 this is changed, read [here](https://github.com/go-siris/siris/blob/master/HISTORY.md#su-11-june-2017--v705).
- Subdomains routing algorithm has been improved.
- Siris router is using a custom interpreter with parser and path evaluator to achieve the best expressiveness, with zero performance loss, you ever seen so far, i.e: 
    - `app.Get("/", "/users/{userid:int min(1)}", handler)`,
        - `{username:string}` or just `{username}`
        - `{asset:path}`,
        - `{firstname:alphabetical}`,
        - `{requestfile:file}` ,
        - `{mylowercaseParam regexp([a-z]+)}`.
        - The previous syntax of `:param` and `*param` still working as expected. Previous rules for paths confliction remain as they were.
            - Also, path parameter names should be only alphabetical now, numbers and symbols are not allowed (for your own good, I have seen a lot the last year...).

Click [here](https://github.com/go-siris/siris/tree/master/_examples/beginner/routing) for details.
> It was my first attempt/experience on the interpreters field, so be good with it :)

Context
- `siris.Context pointer` replaced with `context.Context interface` as we already mention
    - in order to be able to use a custom context and/or catch lifetime like `BeginRequest` and `EndRequest` from context itself, see below
- `context.JSON, context.JSONP, context.XML, context.Markdown, context.HTML` work faster
- `context.Render("filename.ext", bindingViewData{}, options) ` -> `context.View("filename.ext")`
    - `View` renders only templates, it will not try to search if you have a restful renderer adapted, because, now, you can do it via method overriding using a custom Context.
    - Able to set `context.ViewData` and `context.ViewLayout` via middleware when executing a template.
- `context.SetStatusCode(statusCode)` -> `context.StatusCode(statusCode)`
    - which is equivalent with the old `EmitError` too:
        - if status code >=400 given can automatically fire a custom http error handler if response wasn't written already.
    - `context.StatusCode()` -> `context.GetStatusCode()`
    - `app.OnError` -> `app.OnErrorCode`
    - Errors per party are removed by-default, you can just use one global error handler with logic like "if path starts with 'prefix' fire this error handler, else...". 
- Easy way to change Siris' default `Context` with a custom one, see [here](https://github.com/go-siris/siris/tree/master/_examples/intermediate/custom-context)
- `context.ResponseWriter().SetBeforeFlush(...)` works for Flush and HTTP/2 Push, respectfully
- Several improvements under the `Request transactions` 
- Remember that you had to set a status code on each of the render-relative methods? Now it's not required, it just renders
with the status code that user gave with `context.StatusCode` or with `200 OK`, i.e:
    -`context.JSON(siris.StatusOK, myJSON{})` -> `context.JSON(myJSON{})`.
    - Each one of the context's render methods has optional per-call settings,
    - **the new API is even more easier to read, understand and use.**

Server
- Able to set custom underline *http.Server(s) with new Host (aka Server Supervisor) feature 
    - `Done` and `Err` channels to catch shutdown or any errors on custom hosts,
    - Schedule custom tasks(with cancelation) when server is running, see [here](https://github.com/go-siris/siris/tree/master/_examples/intermediate/graceful-shutdown)
- Interrupt handler task for gracefully shutdown (when `CTRL/CMD+C`) are enabled by-default, you can disable its via configuration: `app.Run(siris.Addr(":8080"), siris.WithoutInterruptHandler)`

Future plans
- Future Go1.9's [ServeTLS](https://go-review.googlesource.com/c/38114/2/src/net/http/server.go) is ready when 1.9 released
- Future Go1.9's typealias feature is ready when 1.9 released, i.e `context.Context` -> `siris.Context` just one import path instead of todays' two.