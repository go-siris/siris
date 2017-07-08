// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"time"

	"github.com/satori/go.uuid"
)

const (
	// DefaultCookieName the secret cookie's name for sessions
	DefaultCookieName = "sirissessionid"
	// DefaultCookieLength is the default Session Manager's CookieLength, which is 32
	DefaultCookieLength = 32
)

type (
	// Config is the configuration for sessions. Please review it well before using sessions.
	Config struct {
		// Cookie string, the session's client cookie name, for example: "mysessionid"
		//
		// Defaults to "sirissessionid"
		Cookie string

		// CookieSecureTLS set to true if server is running over TLS
		// and you need the session's cookie "Secure" field to be set true.
		//
		// Note: The user should fill the Decode configuation field in order for this to work.
		// Recommendation: You don't need this to be set to true, just fill the Encode and Decode fields
		// with a third-party library like secure cookie, example is provided at the _examples folder.
		//
		// Defaults to false
		CookieSecureTLS bool

		// Encode the cookie value if not nil.
		// Should accept as first argument the cookie name (config.Name)
		//         as second argument the server's generated session id.
		// Should return the new session id, if error the session id set to empty which is invalid.
		//
		// Note: Errors are not printed, so you have to know what you're doing,
		// and remember: if you use AES it only supports key sizes of 16, 24 or 32 bytes.
		// You either need to provide exactly that amount or you derive the key from what you type in.
		//
		// Defaults to nil
		Encode func(cookieName string, value interface{}) (string, error)
		// Decode the cookie value if not nil.
		// Should accept as first argument the cookie name (config.Name)
		//               as second second accepts the client's cookie value (the encoded session id).
		// Should return an error if decode operation failed.
		//
		// Note: Errors are not printed, so you have to know what you're doing,
		// and remember: if you use AES it only supports key sizes of 16, 24 or 32 bytes.
		// You either need to provide exactly that amount or you derive the key from what you type in.
		//
		// Defaults to nil
		Decode func(cookieName string, cookieValue string, v interface{}) error

		// Expires the duration of which the cookie must expires (created_time.Add(Expires)).
		// If you want to delete the cookie when the browser closes, set it to -1.
		//
		// 0 means no expire, (24 years)
		// -1 means when browser closes
		// > 0 is the time.Duration which the session cookies should expire.
		//
		// Defaults to infinitive/unlimited life duration(0)
		Expires time.Duration

		// SessionIDGenerator should returns a random session id.
		// By default we will use a uuid impl package to generate
		// that, but developers can change that with simple assignment.
		SessionIDGenerator func() string

		// DisableSubdomainPersistence set it to true in order dissallow your q subdomains to have access to the session cookie
		//
		// Defaults to false
		DisableSubdomainPersistence bool
	}
)

// Validate corrects missing fields configuration fields and returns the right configuration
func (c Config) Validate() Config {

	if c.Cookie == "" {
		c.Cookie = DefaultCookieName
	}

	if c.SessionIDGenerator == nil {
		c.SessionIDGenerator = func() string {
			id := uuid.NewV4()
			return id.String()
		}
	}

	return c
}
