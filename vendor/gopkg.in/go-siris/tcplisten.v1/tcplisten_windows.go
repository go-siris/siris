// +build windows

// Package tcplisten provides customizable TCP net.Listener with various
// performance-related options:
//
//   - SO_REUSEPORT. This option allows linear scaling server performance
//     on multi-CPU servers.
//     See https://www.nginx.com/blog/socket-sharding-nginx-release-1-9-1/ for details.
//
//   - TCP_DEFER_ACCEPT. This option expects the server reads from the accepted
//     connection before writing to them.
//
//   - TCP_FASTOPEN. See https://lwn.net/Articles/508865/ for details.
//
// The package is derived from https://github.com/kavu/go_reuseport .
package tcplisten

import (
	"net"
)

// Config provides options to enable on the returned listener.
type Config struct {
	// ReusePort enables SO_REUSEPORT.
	ReusePort bool

	// DeferAccept enables TCP_DEFER_ACCEPT.
	DeferAccept bool

	// FastOpen enables TCP_FASTOPEN.
	FastOpen bool

	// Backlog is the maximum number of pending TCP connections the listener
	// may queue before passing them to Accept.
	// See man 2 listen for details.
	//
	// By default system-level backlog value is used.
	Backlog int
}

// NewListener returns TCP listener with options set in the Config.
//
// The function may be called many times for creating distinct listeners
// with the given config.
//
// Only tcp4 and tcp6 networks are supported.
func (cfg *Config) NewListener(network, addr string) (net.Listener, error) {
	return net.Listen(network, addr)
}
