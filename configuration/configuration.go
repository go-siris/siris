// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"

	"github.com/go-siris/siris/core/errors"
)

var errConfigurationDecode = errors.New("error while trying to decode configuration")

// YAML reads Configuration from a configuration.yml file.
//
// Accepts the absolute path of the configuration.yml.
// An error will be shown to the user via panic with the error message.
// Error may occur when the configuration.yml doesn't exists or is not formatted correctly.
//
// Usage:
// app := siris.Run(siris.Addr(":8080"), siris.WithConfiguration(siris.YAML("myconfig.yml")))
func YAML(filename string) Configuration {
	c := DefaultConfiguration()

	// get the abs
	// which will try to find the 'filename' from current workind dir too.
	yamlAbsPath, err := filepath.Abs(filename)
	if err != nil {
		panic(errConfigurationDecode.AppendErr(err))
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		panic(errConfigurationDecode.AppendErr(err))
	}

	// put the file's contents as yaml to the default configuration(c)
	if err := yaml.Unmarshal(data, &c); err != nil {
		panic(errConfigurationDecode.AppendErr(err))
	}

	return c
}

// TOML reads Configuration from a toml-compatible document file.
// Read more about toml's implementation at:
// https://github.com/toml-lang/toml
//
//
// Accepts the absolute path of the configuration file.
// An error will be shown to the user via panic with the error message.
// Error may occur when the file doesn't exists or is not formatted correctly.
//
// Usage:
// app := siris.Run(siris.Addr(":8080"), siris.WithConfiguration(siris.YAML("myconfig.tml")))
func TOML(filename string) Configuration {
	c := DefaultConfiguration()

	// get the abs
	// which will try to find the 'filename' from current workind dir too.
	tomlAbsPath, err := filepath.Abs(filename)
	if err != nil {
		panic(errConfigurationDecode.AppendErr(err))
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(tomlAbsPath)
	if err != nil {
		panic(errConfigurationDecode.AppendErr(err))
	}

	// put the file's contents as toml to the default configuration(c)
	if _, err := toml.Decode(string(data), &c); err != nil {
		panic(errConfigurationDecode.AppendErr(err))
	}
	// Author's notes:
	// The toml's 'usual thing' for key naming is: the_config_key instead of TheConfigKey
	// but I am always prefer to use the specific programming language's syntax
	// and the original configuration name fields for external configuration files
	// so we do 'toml: "TheConfigKeySameAsTheConfigField" instead.
	return c
}

// Configuration the whole configuration for an Siris instance
// these can be passed via options also, look at the top of this file(configuration.go).
// Configuration is a valid OptionSetter.
type Configuration struct {
	// vhost is private and set only with .Run method, it cannot be changed after the first set.
	// It can be retrieved by the context if needed (i.e router for subdomains)
	vhost    string
	vhostSet bool

	// EnableReuseport if set to true then it turns on the Reuseport feature.
	//
	// Defaults to false.
	EnableReuseport bool `yaml:"EnableReuseport" toml:"EnableReuseport"`

	// EnableQUICSupport if set to true then it turns on the QUIC protocol feature.
	// Only with a TLS Server
	//
	// Defaults to false.
	EnableQUICSupport bool `yaml:"EnableQUICSupport" toml:"EnableQUICSupport"`

	// DisableBanner if set to true then it turns off the write banner on server startup.
	//
	// Defaults to false.
	DisableBanner bool `yaml:"DisableBanner" toml:"DisableBanner"`

	// DisableInterruptHandler if set to true then it disables the automatic graceful server shutdown
	// when control/cmd+C pressed.
	// Turn this to true if you're planning to handle this by your own via a custom host.Task.
	//
	// Defaults to false.
	DisableInterruptHandler bool `yaml:"DisableInterruptHandler" toml:"DisableInterruptHandler"`

	// DisablePathCorrection corrects and redirects the requested path to the registered path
	// for example, if /home/ path is requested but no handler for this Route found,
	// then the Router checks if /home handler exists, if yes,
	// (permant)redirects the client to the correct path /home
	//
	// Defaults to false.
	DisablePathCorrection bool `yaml:"DisablePathCorrection" toml:"DisablePathCorrection"`

	// EnablePathEscape when is true then its escapes the path, the named parameters (if any).
	// Change to false it if you want something like this https://github.com/go-siris/siris/issues/135 to work
	//
	// When do you need to Disable(false) it:
	// accepts parameters with slash '/'
	// Request: http://localhost:8080/details/Project%2FDelta
	// ctx.Param("project") returns the raw named parameter: Project%2FDelta
	// which you can escape it manually with net/url:
	// projectName, _ := url.QueryUnescape(c.Param("project").
	//
	// Defaults to false.
	EnablePathEscape bool `yaml:"EnablePathEscape" toml:"EnablePathEscape"`

	// FireMethodNotAllowed if it's true router checks for StatusMethodNotAllowed(405) and
	//  fires the 405 error instead of 404
	// Defaults to false.
	FireMethodNotAllowed bool `yaml:"FireMethodNotAllowed" toml:"FireMethodNotAllowed"`

	// DisableBodyConsumptionOnUnmarshal manages the reading behavior of the context's body readers/binders.
	// If set to true then it
	// disables the body consumption by the `context.UnmarshalBody/ReadJSON/ReadXML`.
	//
	// By-default io.ReadAll` is used to read the body from the `context.Request.Body which is an `io.ReadCloser`,
	// if this field set to true then a new buffer will be created to read from and the request body.
	// The body will not be changed and existing data before the
	// context.UnmarshalBody/ReadJSON/ReadXML will be not consumed.
	DisableBodyConsumptionOnUnmarshal bool `yaml:"DisableBodyConsumptionOnUnmarshal" toml:"DisableBodyConsumptionOnUnmarshal"`

	// JSONInteratorReplacement manages the reading behavior of the context's body readers/binders.
	// If set to true, it replaces the "encoding / json" libery with jsoniter (JSON-Interator) for
	// `context.UnmarshalBody/ReadJSON/JSON/JSONP`.
	// If returns true then the JSON body is Marshal and Unmarshal by JSON-Interator a dropin replacment
	// for encoding/json.
	JSONInteratorReplacement bool `yaml:"JSONInteratorReplacement" toml:"JSONInteratorReplacement"`

	// DisableAutoFireStatusCode if true then it turns off the http error status code handler automatic execution
	// from "context.StatusCode(>=400)" and instead app should manually call the "context.FireStatusCode(>=400)".
	//
	// By-default a custom http error handler will be fired when "context.StatusCode(code)" called,
	// code should be >=400 in order to be received as an "http error handler".
	//
	// Developer may want this option to set as true in order to manually call the
	// error handlers when needed via "context.FireStatusCode(>=400)".
	// HTTP Custom error handlers are being registered via "framework.OnStatusCode(code, handler)".
	//
	// Defaults to false.
	DisableAutoFireStatusCode bool `yaml:"DisableAutoFireStatusCode" toml:"DisableAutoFireStatusCode"`

	// TimeFormat time format for any kind of datetime parsing
	// Defaults to  "Mon, 02 Jan 2006 15:04:05 GMT".
	TimeFormat string `yaml:"TimeFormat" toml:"TimeFormat"`

	// Charset character encoding for various rendering
	// used for templates and the rest of the responses
	// Defaults to "UTF-8".
	Charset string `yaml:"Charset" toml:"Charset"`

	//  +----------------------------------------------------+
	//  | Context's keys for values used on various featuers |
	//  +----------------------------------------------------+

	// Context values' keys for various features.
	//
	// TranslateLanguageContextKey & TranslateFunctionContextKey are used by i18n handlers/middleware
	// currently we have only one: https://github.com/go-siris/siris/tree/master/middleware/i18n.
	//
	// Defaults to "siris.translate" and "siris.language"
	TranslateFunctionContextKey string `yaml:"TranslateFunctionContextKey" toml:"TranslateFunctionContextKey"`

	// TranslateLanguageContextKey used for i18n.
	//
	// Defaults to "siris.language"
	TranslateLanguageContextKey string `yaml:"TranslateLanguageContextKey" toml:"TranslateLanguageContextKey"`

	// GetViewLayoutContextKey is the key of the context's user values' key
	// which is being used to set the template
	// layout from a middleware or the main handler.
	// Overrides the parent's or the configuration's.
	//
	// Defaults to "siris.ViewLayout"
	ViewLayoutContextKey string `yaml:"ViewLayoutContextKey" toml:"ViewLayoutContextKey"`
	// GetViewDataContextKey is the key of the context's user values' key
	// which is being used to set the template
	// binding data from a middleware or the main handler.
	//
	// Defaults to "siris.viewData"
	ViewDataContextKey string `yaml:"ViewDataContextKey" toml:"ViewDataContextKey"`

	// RemoteAddrHeaders returns the allowed request headers names
	// that can be valid to parse the client's IP based on.
	//
	// Defaults to:
	// "X-Real-Ip":             false,
	// "X-Forwarded-For":       false,
	// "CF-Connecting-IP":      false
	//
	// Look `context.RemoteAddr()` for more.
	RemoteAddrHeaders map[string]bool `yaml:"RemoteAddrHeaders" toml:"RemoteAddrHeaders"`

	// Other are the custom, dynamic options, can be empty.
	// This field used only by you to set any app's options you want
	// or by custom adaptors, it's a way to simple communicate between your adaptors (if any)
	// Defaults to a non-nil empty map
	Other map[string]interface{} `yaml:"Other" toml:"Other"`
}

// GetVHost returns the non-exported vhost config field.
//
// If original addr ended with :443 or :80, it will return the host without the port.
// If original addr was :https or :http, it will return localhost.
// If original addr was 0.0.0.0, it will return localhost.
func (c Configuration) GetVHost() string {
	return c.vhost
}
func (c Configuration) SetVHost(vhost string) {
	if !c.vhostSet && c.vhost != "" {
		c.vhost = vhost
		c.vhostSet = true
	}
	return
}

// GetDisablePathCorrection returns the configuration.DisablePathCorrection,
// DisablePathCorrection corrects and redirects the requested path to the registered path
// for example, if /home/ path is requested but no handler for this Route found,
// then the Router checks if /home handler exists, if yes,
// (permant)redirects the client to the correct path /home.
func (c Configuration) GetDisablePathCorrection() bool {
	return c.DisablePathCorrection
}

// GetEnablePathEscape is the configuration.EnablePathEscape,
// returns true when its escapes the path, the named parameters (if any).
func (c Configuration) GetEnablePathEscape() bool {
	return c.EnablePathEscape
}

// GetEnableReuseport is the configuration.EnableReuseport,
// returns true when its use the feature of the Reusepost listener.
func (c Configuration) GetEnableReuseport() bool {
	return c.EnableReuseport
}

// GetEnableQUICSupport is the configuration.EnableQUICSupport,
// returns true when its use the feature of the QUIC protocol for TLS Server.
func (c Configuration) GetEnableQUICSupport() bool {
	return c.EnableQUICSupport
}

// GetFireMethodNotAllowed returns the configuration.FireMethodNotAllowed.
func (c Configuration) GetFireMethodNotAllowed() bool {
	return c.FireMethodNotAllowed
}

// GetDisableBodyConsumptionOnUnmarshal returns the configuration.DisableBodyConsumptionOnUnmarshal,
// manages the reading behavior of the context's body readers/binders.
// If returns true then the body consumption by the `context.UnmarshalBody/ReadJSON/ReadXML`
// is disabled.
//
// By-default io.ReadAll` is used to read the body from the `context.Request.Body which is an `io.ReadCloser`,
// if this field set to true then a new buffer will be created to read from and the request body.
// The body will not be changed and existing data before the
// context.UnmarshalBody/ReadJSON/ReadXML will be not consumed.
func (c Configuration) GetDisableBodyConsumptionOnUnmarshal() bool {
	return c.DisableBodyConsumptionOnUnmarshal
}

// GetJSONInteratorReplacement returns the configuration.JSONInteratorReplacement,
// manages the reading behavior of the context's body readers/binders.
//
// If returns true then the JSON body is Marshal and Unmarshal by JSON-Interator a dropin replacment
// for encoding/json.
func (c Configuration) GetJSONInteratorReplacement() bool {
	return c.JSONInteratorReplacement
}

// GetDisableAutoFireStatusCode returns the configuration.DisableAutoFireStatusCode.
// Returns true when the http error status code handler automatic execution turned off.
func (c Configuration) GetDisableAutoFireStatusCode() bool {
	return c.DisableAutoFireStatusCode
}

// GetTimeFormat returns the configuration.TimeFormat,
// format for any kind of datetime parsing.
func (c Configuration) GetTimeFormat() string {
	return c.TimeFormat
}

// GetCharset returns the configuration.Charset,
// the character encoding for various rendering
// used for templates and the rest of the responses.
func (c Configuration) GetCharset() string {
	return c.Charset
}

// GetTranslateFunctionContextKey returns the configuration's TranslateFunctionContextKey value,
// used for i18n.
func (c Configuration) GetTranslateFunctionContextKey() string {
	return c.TranslateFunctionContextKey
}

// GetTranslateLanguageContextKey returns the configuration's TranslateLanguageContextKey value,
// used for i18n.
func (c Configuration) GetTranslateLanguageContextKey() string {
	return c.TranslateLanguageContextKey
}

// GetViewLayoutContextKey returns the key of the context's user values' key
// which is being used to set the template
// layout from a middleware or the main handler.
// Overrides the parent's or the configuration's.
func (c Configuration) GetViewLayoutContextKey() string {
	return c.ViewLayoutContextKey
}

// GetViewDataContextKey returns the key of the context's user values' key
// which is being used to set the template
// binding data from a middleware or the main handler.
func (c Configuration) GetViewDataContextKey() string {
	return c.ViewDataContextKey
}

// GetRemoteAddrHeaders returns the allowed request headers names
// that can be valid to parse the client's IP based on.
//
// Defaults to:
// "X-Real-Ip":             false,
// "X-Forwarded-For":       false,
// "CF-Connecting-IP":      false
//
// Look `context.RemoteAddr()` for more.
func (c Configuration) GetRemoteAddrHeaders() map[string]bool {
	return c.RemoteAddrHeaders
}

// GetOther returns the configuration.Other map.
func (c Configuration) GetOther() map[string]interface{} {
	return c.Other
}

// DefaultConfiguration returns the default configuration for an Siris station, fills the main Configuration
func DefaultConfiguration() Configuration {
	return Configuration{
		EnableReuseport:                   false,
		EnableQUICSupport:                 false,
		DisableBanner:                     false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		TimeFormat:                        "Mon, Jan 02 2006 15:04:05 GMT",
		Charset:                           "UTF-8",
		TranslateFunctionContextKey:       "siris.translate",
		TranslateLanguageContextKey:       "siris.language",
		ViewLayoutContextKey:              "siris.viewLayout",
		ViewDataContextKey:                "siris.viewData",
		RemoteAddrHeaders: map[string]bool{
			"X-Real-Ip":        false,
			"X-Forwarded-For":  false,
			"CF-Connecting-IP": false,
		},
		Other: make(map[string]interface{}),
	}
}
