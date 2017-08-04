package router

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-siris/siris/core/nettools"
)

const (
	// ParamStart the character in string representation where the underline router starts its dynamic named parameter.
	ParamStart = ":"
	// WildcardParamStart the character in string representation where the underline router starts its dynamic wildcard
	// path parameter.
	WildcardParamStart = "*"
)

// Param receives a parameter name prefixed with the ParamStart symbol.
func Param(name string) string {
	return prefix(name, ParamStart)
}

// WildcardParam receives a parameter name prefixed with the WildcardParamStart symbol.
func WildcardParam(name string) string {
	if len(name) == 0 {
		return ""
	}
	return prefix(name, WildcardParamStart)
}

func prefix(s string, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		return prefix + s
	}

	return s
}

func suffix(s string, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		return s + suffix
	}
	return s
}

func joinPath(path1 string, path2 string) string {
	return path.Join(path1, path2)
}

// cleanPath applies the following rules
// iteratively until no further processing can be done:
//
//	1. Replace multiple slashes with a single slash.
//	3. Eliminate each inner .. path name element (the parent directory)
//	   along with the non-.. element that precedes it.
//	4. Eliminate .. elements that begin a rooted path:
//	   that is, replace "/.." by "/" at the beginning of a path.
//
// The returned path ends in a slash only if it is the root "/".
func cleanPath(s string) string {
	if s == "" || s == "." {
		return "/"
	}

	// remove suffix "/"
	if lidx := len(s) - 1; s[lidx] == '/' {
		s = s[:lidx]
	}

	// prefix with "/"
	s = prefix(s, "/")

	// remove the os specific dir sep
	s = strings.Replace(s, "\\", "/", -1)

	// use std path to clean the path
	s = path.Clean(s)

	return s
}

const (
	// SubdomainWildcardIndicator where a registered path starts with '*.'.
	// if subdomain == "*." then its wildcard.
	//
	// used internally by router and api builder.
	SubdomainWildcardIndicator = "*."

	// SubdomainWildcardPrefix where a registered path starts with "*./",
	// then this route should accept any subdomain.
	SubdomainWildcardPrefix = SubdomainWildcardIndicator + "/"
	// SubdomainPrefix where './' exists in a registered path then it contains subdomain
	//
	// used on api builder.
	SubdomainPrefix = "./" // i.e subdomain./ -> Subdomain: subdomain. Path: /
)

func hasSubdomain(s string) bool {
	if s == "" {
		return false
	}

	// subdomain./path
	// .*/path
	//
	// remember: path always starts with "/"
	// if not start with "/" then it should be something else,
	// we don't assume anything else but subdomain.
	slashIdx := strings.IndexByte(s, '/')
	return slashIdx > 0 || s[0] == SubdomainPrefix[0] || (len(s) >= 2 && s[0:2] == SubdomainWildcardIndicator)
}

// splitSubdomainAndPath checks if the path has subdomain and if it's
// it splits the subdomain and path and returns them, otherwise it returns
// an empty subdomain and the clean path.
//
// First return value is the subdomain, second is the path.
func splitSubdomainAndPath(fullUnparsedPath string) (subdomain string, path string) {
	s := fullUnparsedPath
	if s == "" || s == "/" {
		return "", "/"
	}

	slashIdx := strings.IndexByte(s, '/')
	if slashIdx == 0 {
		// no subdomain
		return "", cleanPath(s)
	}

	return s[0:slashIdx], cleanPath(s[slashIdx:]) // return subdomain without slash, path with slash
}

// RoutePathReverserOption option signature for the RoutePathReverser.
type RoutePathReverserOption func(*RoutePathReverser)

// WithScheme is an option for the RoutepathReverser,
// it sets the optional field "vscheme",
// v for virtual.
// if vscheme is empty then it will try to resolve it from
// the RoutePathReverser's vhost field.
//
// See WithHost or WithServer to enable the URL feature.
func WithScheme(scheme string) RoutePathReverserOption {
	return func(ps *RoutePathReverser) {
		ps.vscheme = scheme
	}
}

// WithHost enables the RoutePathReverser's URL feature.
// Both "WithHost" and "WithScheme" can be different from
// the real server's listening address, i.e nginx in front.
func WithHost(host string) RoutePathReverserOption {
	return func(ps *RoutePathReverser) {
		ps.vhost = host
		if ps.vscheme == "" {
			ps.vscheme = nettools.ResolveSchemeFromVHost(host)
		}
	}
}

// WithServer enables the RoutePathReverser's URL feature.
// It receives an *http.Server and tries to resolve
// a scheme and a host to be used in the URL function.
func WithServer(srv *http.Server) RoutePathReverserOption {
	return func(ps *RoutePathReverser) {
		ps.vhost = nettools.ResolveVHost(srv.Addr)
		ps.vscheme = nettools.ResolveSchemeFromServer(srv)
	}
}

// RoutePathReverser contains methods that helps to reverse a
// (dynamic) path from a specific route,
// route name is required because a route may being registered
// on more than one http method.
type RoutePathReverser struct {
	provider RoutesProvider
	// both vhost and vscheme are being used, optionally, for the URL feature.
	vhost   string
	vscheme string
}

// NewRoutePathReverser returns a new path reverser based on
// a routes provider, needed to get a route based on its name.
// Options is required for the URL function.
// See WithScheme and WithHost or WithServer.
func NewRoutePathReverser(apiRoutesProvider RoutesProvider, options ...RoutePathReverserOption) *RoutePathReverser {
	ps := &RoutePathReverser{
		provider: apiRoutesProvider,
	}
	for _, o := range options {
		o(ps)
	}
	return ps
}

// Path  returns a route path based on a route name and any dynamic named parameter's values-only.
func (ps *RoutePathReverser) Path(routeName string, paramValues ...interface{}) string {
	r := ps.provider.GetRoute(routeName)
	if r == nil {
		return ""
	}

	if len(paramValues) == 0 {
		return r.Path
	}

	return r.ResolvePath(toStringSlice(paramValues)...)
}

func toStringSlice(args []interface{}) []string {
	var argsString []string
	if len(args) > 0 {
		argsString = make([]string, len(args), len(args))
	}

	for i, v := range args {
		if s, ok := v.(string); ok {
			argsString[i] = s
		} else if num, ok := v.(int); ok {
			argsString[i] = strconv.Itoa(num)
		} else if b, ok := v.(bool); ok {
			argsString[i] = strconv.FormatBool(b)
		} else if arr, ok := v.([]string); ok {
			if len(arr) > 0 {
				argsString[i] = arr[0]
				argsString = append(argsString, arr[1:]...)
			}
		}
	}
	return argsString
}

// Remove the URL for now, it complicates things for the whole framework without a specific benefits,
// developers can just concat the subdomain, (host can be auto-retrieve by browser using the Path).

// URL same as Path but returns the full uri, i.e https://mysubdomain.mydomain.com/hello/iris
func (ps *RoutePathReverser) URL(routeName string, paramValues ...interface{}) (url string) {
	if ps.vhost == "" || ps.vscheme == "" {
		return "not supported"
	}

	r := ps.provider.GetRoute(routeName)
	if r == nil {
		return
	}

	if len(paramValues) == 0 {
		return r.Path
	}

	args := toStringSlice(paramValues)

	host := ps.vhost
	scheme := ps.vscheme
	// if it's dynamic subdomain then the first argument is the subdomain part
	// for this part we are responsible not the custom routers
	if r.Subdomain == SubdomainWildcardIndicator {
		subdomain := args[0]
		host = subdomain + "." + host
		args = args[1:] // remove the subdomain part for the arguments,

	}

	if parsedPath := r.ResolvePath(args...); parsedPath != "" {
		url = scheme + "://" + host + parsedPath
	}

	return
}
