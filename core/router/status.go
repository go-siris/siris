// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http" // just for status codes
	"sync"

	"github.com/go-siris/siris/context"
)

// ErrorCodeHandler is the entry
// of the list of all http error code handlers.
type ErrorCodeHandler struct {
	StatusCode int
	Handler    context.Handler
	mu         sync.Mutex
}

// Fire executes the specific an error http error status.
// it's being wrapped to make sure that the handler
// will render correctly.
func (ch *ErrorCodeHandler) Fire(ctx context.Context) {
	// if we can reset the body
	if w, ok := ctx.IsRecording(); ok {
		// reset if previous content and it's recorder
		w.Reset()
	} else if w, ok := ctx.ResponseWriter().(*context.GzipResponseWriter); ok {
		// reset and disable the gzip in order to be an expected form of http error result
		w.ResetBody()
		w.Disable()
	} else {
		// if we can't reset the body and the body has been filled
		// which means that the status code already sent,
		// then do not fire this custom error code.
		if ctx.ResponseWriter().Written() != -1 {
			return
		}
	}
	// ctx.StopExecution() // not uncomment this, is here to remember why to.
	// note for me: I don't stopping the execution of the other handlers
	// because may the user want to add a fallback error code
	// i.e
	// users := app.Party("/users")
	// users.Done(func(ctx context.Context){ if ctx.StatusCode() == 400 { /*  custom error code for /users */ }})
	ch.Handler(ctx)
}

func (ch *ErrorCodeHandler) updateHandler(h context.Handler) {
	ch.mu.Lock()
	ch.Handler = h
	ch.mu.Unlock()
}

// ErrorCodeHandlers contains the http error code handlers.
// User of this struct can register, get
// a status code handler based on a status code or
// fire based on a receiver context.
type ErrorCodeHandlers struct {
	handlers []*ErrorCodeHandler
}

func defaultErrorCodeHandlers() *ErrorCodeHandlers {
	chs := new(ErrorCodeHandlers)
	// register some common error handlers.
	// Note that they can be registered on-fly but
	// we don't want to reduce the performance even
	// on the first failed request.
	for _, statusCode := range []int{
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusInternalServerError} {
		chs.Register(statusCode, statusText(statusCode))
	}

	return chs
}

func statusText(statusCode int) context.Handler {
	return func(ctx context.Context) {
		if _, err := ctx.WriteString(http.StatusText(statusCode)); err != nil {
			// ctx.Application().Logger().Info("(status code: %d) %s",
			// 	err.Error(), statusCode)
		}
	}
}

// Get returns an http error handler based on the "statusCode".
// If not found it returns nil.
func (s *ErrorCodeHandlers) Get(statusCode int) *ErrorCodeHandler {
	for i, n := 0, len(s.handlers); i < n; i++ {
		if h := s.handlers[i]; h.StatusCode == statusCode {
			return h
		}
	}
	return nil
}

// Register registers an error http status code
// based on the "statusCode" >= 400.
// The handler is being wrapepd by a generic
// handler which will try to reset
// the body if recorder was enabled
// and/or disable the gzip if gzip response recorder
// was active.
func (s *ErrorCodeHandlers) Register(statusCode int, handler context.Handler) *ErrorCodeHandler {
	if statusCode < 400 {
		return nil
	}

	h := s.Get(statusCode)
	if h == nil {
		ch := &ErrorCodeHandler{
			StatusCode: statusCode,
			Handler:    handler,
		}
		s.handlers = append(s.handlers, ch)
		// create new and add it
		return ch
	}
	// otherwise update the handler
	h.updateHandler(handler)
	return h
}

// Fire executes an error http status code handler
// based on the context's status code.
//
// If a handler is not already registered,
// then it creates & registers a new trivial handler on the-fly.
func (s *ErrorCodeHandlers) Fire(ctx context.Context) {
	statusCode := ctx.GetStatusCode()
	if statusCode < 400 {
		return
	}
	ch := s.Get(statusCode)
	if ch == nil {
		ch = s.Register(statusCode, statusText(statusCode))
	}

	ch.Fire(ctx)
}
