# ![Logo](go-siris.jpg)

A fast, cross-platform and efficient web framework with robust set of well-designed features, written entirely in Go.

[![Build status](https://api.travis-ci.org/go-siris/siris.svg?branch=master&style=flat-square)](https://travis-ci.org/go-siris/siris)
[![Report card](https://img.shields.io/badge/report%20card%20-a%2B-F44336.svg?style=flat-square)](https://goreportcard.com/report/github.com/go-siris/siris)
[![Support forum](https://img.shields.io/badge/support-page-ec2eb4.svg?style=flat-square)](http://support.go-siris.com)
[![Examples](https://img.shields.io/badge/howto-examples-3362c2.svg?style=flat-square)](https://github.com/go-siris/siris/tree/master/_examples#table-of-contents)
[![Godocs](https://img.shields.io/badge/7.3.4-%20documentation-5272B4.svg?style=flat-square)](https://godoc.org/github.com/go-siris/siris)
[![Chat](https://img.shields.io/badge/community-%20chat-00BCD4.svg?style=flat-square)](https://gitter.im/gosiris/siris)

Build your own web applications and portable APIs with the highest performance and countless potentials.

If you're coming from [Node.js](https://nodejs.org) world, this is the [expressjs](https://github.com/expressjs/express)++ equivalent for the [Go Programming Language](https://golang.org).


# Table of contents

* [Installation](#installation)
* [Feature overview](#feature-overview)
* [Documentation](#documentation)
    * [Examples](https://github.com/go-siris/siris/tree/master/_examples)
* [Support](#support)
* [Third-party middleware list](#third-party-middleware)    
* [Testing](#testing)
* [Philosophy](#philosophy)
* [People](#people)
* [Versioning](#version)
    * [When should I upgrade?](#should-i-upgrade-my-siris)
    * [Where can I find older versions?](#where-can-i-find-older-versions)

# Installation

The only requirement is the [Go Programming Language](https://golang.org/dl/), at least version 1.8

```sh
$ go get -u github.com/go-siris/siris
```

> Siris uses the [vendor directory](https://docs.google.com/document/d/1Bz5-UB7g2uPBdOx-rw5t9MxJwkfpx90cqG9AFL0JAYo) feature, so you get truly reproducible builds, as this method guards against upstream renames and deletes.

```go
package main

import (
    "github.com/go-siris/siris"
    "github.com/go-siris/siris/context"
    "github.com/go-siris/siris/view"
)

// User is just a bindable object structure.
type User struct {
    Username  string `json:"username"`
    Firstname string `json:"firstname"`
    Lastname  string `json:"lastname"`
    City      string `json:"city"`
    Age       int    `json:"age"`
}

func main() {
    app := siris.New()

    // Define templates using the std html/template engine.
    // Parse and load all files inside "./views" folder with ".html" file extension.
    // Reload the templates on each request (development mode).
    app.AttachView(view.HTML("./views", ".html").Reload(true))

    // Regster custom handler for specific http errors.
    app.OnErrorCode(siris.StatusInternalServerError, func(ctx context.Context) {
    	// .Values are used to communicate between handlers, middleware.
    	errMessage := ctx.Values().GetString("error")
    	if errMessage != "" {
    		ctx.Writef("Internal server error: %s", errMessage)
    		return
    	}

    	ctx.Writef("(Unexpected) internal server error")
    })

    app.Use(func(ctx context.Context) {
    	ctx.Application().Log("Begin request for path: %s", ctx.Path())
    	ctx.Next()
    })

    // app.Done(func(ctx context.Context) {})

    // Method POST: http://localhost:8080/decode
    app.Post("/decode", func(ctx context.Context) {
    	var user User
    	ctx.ReadJSON(&user)
    	ctx.Writef("%s %s is %d years old and comes from %s", user.Firstname, user.Lastname, user.Age, user.City)
    })

    // Method GET: http://localhost:8080/encode
    app.Get("/encode", func(ctx context.Context) {
    	doe := User{
    		Username:  "Johndoe",
    		Firstname: "John",
    		Lastname:  "Doe",
    		City:      "Neither FBI knows!!!",
    		Age:       25,
    	}

    	ctx.JSON(doe)
    })

    // Method GET: http://localhost:8080/profile/anytypeofstring
    app.Get("/profile/{username:string}", profileByUsername)

    usersRoutes := app.Party("/users", logThisMiddleware)
    {
    	// Method GET: http://localhost:8080/users/42
    	usersRoutes.Get("/{id:int min(1)}", getUserByID)
    	// Method POST: http://localhost:8080/users/create
    	usersRoutes.Post("/create", createUser)
    }

    // Listen for incoming HTTP/1.x & HTTP/2 clients on localhost port 8080.
    app.Run(siris.Addr(":8080"), siris.WithCharset("UTF-8"))
}

func logThisMiddleware(ctx context.Context) {
    ctx.Application().Log("Path: %s | IP: %s", ctx.Path(), ctx.RemoteAddr())

    // .Next is required to move forward to the chain of handlers,
    // if missing then it stops the execution at this handler.
    ctx.Next()
}

func profileByUsername(ctx context.Context) {
    // .Params are used to get dynamic path parameters.
    username := ctx.Params().Get("username")
    ctx.ViewData("Username", username)
    // renders "./views/users/profile.html"
    // with {{ .Username }} equals to the username dynamic path parameter.
    ctx.View("users/profile.html")
}

func getUserByID(ctx context.Context) {
    userID := ctx.Params().Get("id") // Or convert directly using: .Values().GetInt/GetInt64 etc...
    // your own db fetch here instead of user :=...
    user := User{Username: "username" + userID}

    ctx.XML(user)
}

func createUser(ctx context.Context) {
    var user User
    err := ctx.ReadForm(&user)
    if err != nil {
    	ctx.Values().Set("error", "creating user, read and parse form failed. "+err.Error())
    	ctx.StatusCode(siris.StatusInternalServerError)
    	return
    }
    // renders "./views/users/create_verification.html"
    // with {{ . }} equals to the User object, i.e {{ .Username }} , {{ .Firstname}} etc...
    ctx.ViewData("", user)
    ctx.View("users/create_verification.html")
}
```

# Feature Overview

| Feature | More Information |
| --------|------------------|
| high performance | [benchmark](https://github.com/smallnest/go-web-framework-benchmark) |
| Authentication | [basicauth](_examples/beginner/basicauth) [oauth2](_examples/intermediate/oauth2) TODO JWT|
| Cache | [cache-markdown](_examples/intermediate/cache-markdown)  |
| Certificates | [letsencrypt](/_examples/beginner/listening/listen-letsencrypt/main.go) [custom](/_examples/beginner/listening/listen-tls/main.go) |
| Compression | gzip and deflate |
| Decode Json, Forms | [json](_examples/beginner/read-json) [form](_examples/beginner/read-form)|
| Encode Json, JsonP, XML, Forms, Markdown | [json](_examples/beginner/write-json) [cache-markdown](_examples/intermediate/cache-markdown) |
| Error handling | custom [error handler] [panic handler]|
| http2 push | TODO add Documentation |
| Limits | RequestBody TODO add docu|
| Localization | [i18n](_examples/intermediate/i18n) |
| Logger Engines | [file-logger](_examples/beginner/file-logger) [request-logger](_examples/beginner/request-logger) |
| Routing | [routing](_examples/beginner/routing)  |
| Sessions | [sessions](_examples/intermediate/sessions) |
| Static Files | [file-server](_examples/beginner/file-server) |
| Subdomains and Grouping | [subdomains](_examples/intermediate/subdomains) |
| Tempalte HTML, django, pug(jade), handlebars, amber | [views](_examples/intermediate/view) |
| Tooling | Typescript integration + Web IDE |
| Websockets | [websockets](_examples/intermediate/websockets) |

# Documentation

Small but practical [examples](https://github.com/go-siris/siris/tree/master/_examples#table-of-contents) --they cover each feature.

Wanna create your own fast URL Shortener Service Using Siris? --click [here](https://medium.com/@kataras/a-url-shortener-service-using-go-iris-and-bolt-4182f0b00ae7) to learn how.

[Godocs](https://godoc.org/github.com/go-siris/siris) --for deep understanding.

[Hints](HINTS.md) --some performance hints and tooling.

# Support

- [Chat](https://gitter.im/gosiris/siris) Ask the Community for help
- [Post](https://github.com/go-siris/siris/issues) a feature request or report a bug, will help to make the framework even better.
- :star: and watch [the project](https://github.com/go-siris/siris/stargazers), will notify you about updates.
- :earth_americas: publish [an article](https://medium.com/) or share a [tweet](https://twitter.com/) about siris.

# Third Party Middleware

See [THIRD-PARTY-MIDDLEWARE.md](https://github.com/go-siris/siris/blob/master/THIRD-PARTY-MIDDLEWARE.md)

# Testing

The `httptest` package is your way for end-to-end HTTP testing, it uses the httpexpect library created by our friend, [gavv](https://github.com/gavv).

A simple test is located to [./_examples/intermediate/httptest/main_test.go](https://github.com/go-siris/siris/blob/master/_examples/intermediate/httptest/main_test.go)

# Philosophy

The Siris philosophy is to provide robust tooling for HTTP, making it a great solution for single page applications, web sites, hybrids, or public HTTP APIs.

Siris does not force you to use any specific ORM or template engine. With support for the most popular template engines, you can quickly craft your perfect application.

# People

The original author of Iris is [@kataras](https://github.com/kataras). Before we forked to Siris.

However the real Success of Siris belongs to you with your bug reports and feature requests that made this Framework so Unique.

# Version

Current: **7.3.4**

Each new release is pushed to the master. It stays there until the next version. When a next version is released then the previous version goes to its own branch with `gopkg.in` as its import path (and its own vendor folder), in order to keep it working "for-ever".

Community members can request additional features or report a bug fix for a specific siris version.


### Should I upgrade my Siris?

Developers are not forced to use the latest Siris version, they can use any version in production, they can update at any time they want.

Testers should upgrade immediately, if you're willing to use Siris in production you can wait a little more longer, transaction should be as safe as possible.

### Where can I find older versions?

Each Siris version is independent. Only bug fixes, Router's API and experience are kept.

Previous versions can be found at [releases page](https://github.com/go-siris/siris/releases).

Previous Iris versions are archived at [iris archive](https://github.com/go-siris/iris)


# License

Unless otherwise noted, the source files are distributed
under the BSD-3-Clause License found in the [LICENSE file](LICENSE).

Note that some third-party packages that you use with Siris may requires
different license agreements.
