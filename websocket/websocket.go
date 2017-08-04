// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"strings"

	"github.com/go-siris/siris"
)

// New returns a new websocket server policy adaptor.
func New(cfg Config) Server {
	return &server{
		config: cfg.Validate(),
		rooms:  make(map[string][]string, 0),
		onConnectionListeners: make([]ConnectionFunc, 0),
	}
}

func fixPath(s string) string {
	if s == "" {
		return ""
	}

	if s[0] != '/' {
		s = "/" + s
	}

	s = strings.Replace(s, "//", "/", -1)
	return s
}

// Attach adapts the websocket server to one or more Siris instances.
func (s *server) Attach(app *siris.Application) {
	wsPath := fixPath(s.config.Endpoint)
	if wsPath == "" {
		app.Logger().Warnf("websocket's configuration field 'Endpoint' cannot be empty, websocket server stops")
		return
	}

	wsClientSidePath := fixPath(s.config.ClientSourcePath)
	if wsClientSidePath == "" {
		app.Logger().Warnf("websocket's configuration field 'ClientSourcePath' cannot be empty, websocket server stops")
		return
	}

	// set the routing for client-side source (javascript) (optional)
	clientSideLookupName := "siris-websocket-client-side"
	wsHandler := s.Handler()
	app.Get(wsPath, wsHandler)
	// check if client side doesn't already exists
	if app.GetRoute(clientSideLookupName) == nil {
		// serve the client side on domain:port/siris-ws.js
		r := app.StaticContent(wsClientSidePath, "application/javascript", ClientSource)
		if r == nil {
			app.Logger().Warnf("websocket's route for javascript client-side library failed")
			return
		}
		r.Name = clientSideLookupName
	}
}
