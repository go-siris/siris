// Copyright 2017 Go-SIRIS Authors. All Rights Reserved.

// Package sessions provider
package sessions

import (
	"net/http"
)

type (

	// Sessions must be implemented within a session manager.
	//
	// A Sessions should be responsible to Start a sesion based
	// on raw http.ResponseWriter and http.Request, which should return
	// a compatible Session interface, type. If the external session manager
	// doesn't qualifies, then the user should code the rest of the functions with empty implementation.
	//
	// Sessions should be responsible to Destroy a session based
	// on the http.ResponseWriter and http.Request, this function should works individually.
	Sessions interface {
		// Start should start the session for the particular net/http request.
		SessionStart(http.ResponseWriter, *http.Request) (Store, error)

		// Destroy should kills the net/http session and remove the associated cookie.
		SessionDestroy(http.ResponseWriter, *http.Request)

		// Regenerate a new sessionId for security reasons.
		SessionRegenerateID(http.ResponseWriter, *http.Request) Store
	} // Sessions is being implemented by Manager
)

var _ Sessions = &Manager{}
