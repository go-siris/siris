package siris

import (
	"github.com/go-siris/siris/configuration"
	"github.com/go-siris/siris/context"
)

// YAML load the YAML Configuration
var YAML = configuration.YAML

// TOML load the TOML Configuration
var TOML = configuration.TOML

var _ context.ConfigurationReadOnly = &configuration.Configuration{}

// Configurator is just an interface which accepts the framework instance.
//
// It can be used to register a custom configuration with `Configure` in order
// to modify the framework instance.
//
// Currently Configurator is being used to describe the configuration's fields values.
type Configurator func(*Application)

// variables for configurators don't need any receivers, functions
// for them that need (helps code editors to recognise as variables without parenthesis completion).

// EnableReuseport turns on the Reuseport feature.
var EnableReuseport = func(app *Application) {
	app.config.EnableReuseport = true
}

// EnableQUICSupport turns on the Reuseport feature.
var EnableQUICSupport = func(app *Application) {
	app.config.EnableQUICSupport = true
}

// WithoutBanner turns off the write banner on server startup.
var WithoutBanner = func(app *Application) {
	app.config.DisableBanner = true
}

// WithoutInterruptHandler disables the automatic graceful server shutdown
// when control/cmd+C pressed.
var WithoutInterruptHandler = func(app *Application) {
	app.config.DisableInterruptHandler = true
}

// WithoutPathCorrection disables the PathCorrection setting.
//
// See `Configuration`.
var WithoutPathCorrection = func(app *Application) {
	app.config.DisablePathCorrection = true
}

// WithoutBodyConsumptionOnUnmarshal disables BodyConsumptionOnUnmarshal setting.
//
// See `Configuration`.
var WithoutBodyConsumptionOnUnmarshal = func(app *Application) {
	app.config.DisableBodyConsumptionOnUnmarshal = true
}

// WithJSONInteratorReplacement enables JSONInteratorReplacement setting.
//
// See `Configuration`.
var WithJSONInteratorReplacement = func(app *Application) {
	app.config.JSONInteratorReplacement = true
}

// WithoutAutoFireStatusCode disables the AutoFireStatusCode setting.
//
// See `Configuration`.
var WithoutAutoFireStatusCode = func(app *Application) {
	app.config.DisableAutoFireStatusCode = true
}

// WithPathEscape enanbles the PathEscape setting.
//
// See `Configuration`.
var WithPathEscape = func(app *Application) {
	app.config.EnablePathEscape = true
}

// WithFireMethodNotAllowed enanbles the FireMethodNotAllowed setting.
//
// See `Configuration`.
var WithFireMethodNotAllowed = func(app *Application) {
	app.config.FireMethodNotAllowed = true
}

// WithTimeFormat sets the TimeFormat setting.
//
// See `Configuration`.
func WithTimeFormat(timeformat string) Configurator {
	return func(app *Application) {
		app.config.TimeFormat = timeformat
	}
}

// WithCharset sets the Charset setting.
//
// See `Configuration`.
func WithCharset(charset string) Configurator {
	return func(app *Application) {
		app.config.Charset = charset
	}
}

// WithRemoteAddrHeader enables or adds a new or existing request header name
// that can be used to validate the client's real IP.
//
// Existing values are:
// "X-Real-Ip":             false,
// "X-Forwarded-For":       false,
// "CF-Connecting-IP":      false
//
// Look `context.RemoteAddr()` for more.
func WithRemoteAddrHeader(headerName string) Configurator {
	return func(app *Application) {
		if app.config.RemoteAddrHeaders == nil {
			app.config.RemoteAddrHeaders = make(map[string]bool)
		}
		app.config.RemoteAddrHeaders[headerName] = true
	}
}

// WithoutRemoteAddrHeader disables an existing request header name
// that can be used to validate the client's real IP.
//
// Existing values are:
// "X-Real-Ip":             false,
// "X-Forwarded-For":       false,
// "CF-Connecting-IP": 	    false
//
// Look `context.RemoteAddr()` for more.
func WithoutRemoteAddrHeader(headerName string) Configurator {
	return func(app *Application) {
		if app.config.RemoteAddrHeaders == nil {
			app.config.RemoteAddrHeaders = make(map[string]bool)
		}
		app.config.RemoteAddrHeaders[headerName] = false
	}
}

// WithOtherValue adds a value based on a key to the Other setting.
//
// See `Configuration`.
func WithOtherValue(key string, val interface{}) Configurator {
	return func(app *Application) {
		if app.config.Other == nil {
			app.config.Other = make(map[string]interface{})
		}
		app.config.Other[key] = val
	}
}

// WithConfiguration sets the "c" values to the framework's configurations.
//
// Usage:
// app.Run(siris.Addr(":8080"), siris.WithConfiguration(siris.Configuration{/* fields here */ }))
// or
// siris.WithConfiguration(siris.YAML("./cfg/iris.yml"))
// or
// siris.WithConfiguration(siris.TOML("./cfg/iris.tml"))
func WithConfiguration(c configuration.Configuration) Configurator {
	return func(app *Application) {
		main := app.config

		if v := c.EnableQUICSupport; v {
			main.EnableQUICSupport = v
		}

		if v := c.EnableReuseport; v {
			main.EnableReuseport = v
		}

		if v := c.DisableBanner; v {
			main.DisableBanner = v
		}

		if v := c.DisableInterruptHandler; v {
			main.DisableInterruptHandler = v
		}

		if v := c.DisablePathCorrection; v {
			main.DisablePathCorrection = v
		}

		if v := c.EnablePathEscape; v {
			main.EnablePathEscape = v
		}

		if v := c.FireMethodNotAllowed; v {
			main.FireMethodNotAllowed = v
		}

		if v := c.DisableBodyConsumptionOnUnmarshal; v {
			main.DisableBodyConsumptionOnUnmarshal = v
		}

		if v := c.DisableAutoFireStatusCode; v {
			main.DisableAutoFireStatusCode = v
		}

		if v := c.TimeFormat; v != "" {
			main.TimeFormat = v
		}

		if v := c.Charset; v != "" {
			main.Charset = v
		}

		if v := c.TranslateFunctionContextKey; v != "" {
			main.TranslateFunctionContextKey = v
		}

		if v := c.TranslateLanguageContextKey; v != "" {
			main.TranslateLanguageContextKey = v
		}

		if v := c.ViewLayoutContextKey; v != "" {
			main.ViewLayoutContextKey = v
		}

		if v := c.ViewDataContextKey; v != "" {
			main.ViewDataContextKey = v
		}

		if v := c.RemoteAddrHeaders; len(v) > 0 {
			if main.RemoteAddrHeaders == nil {
				main.RemoteAddrHeaders = make(map[string]bool)
			}
			for key, value := range v {
				main.RemoteAddrHeaders[key] = value
			}
		}

		if v := c.Other; len(v) > 0 {
			if main.Other == nil {
				main.Other = make(map[string]interface{})
			}
			for key, value := range v {
				main.Other[key] = value
			}
		}
	}
}
