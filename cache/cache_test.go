// black-box testing
package cache_test

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-siris/siris/cache"
	"github.com/go-siris/siris/cache/client/rule"

	"github.com/go-siris/siris"
	"github.com/go-siris/siris/context"
	"github.com/go-siris/siris/core/errors"

	"github.com/go-siris/siris/httptest"
	"gopkg.in/gavv/httpexpect.v1"
)

var (
	cacheDuration   = 2 * time.Second
	expectedBodyStr = "Imagine it as a big message to achieve x20 response performance!"
	errTestFailed   = errors.New("expected the main handler to be executed %d times instead of %d")
)

func runTest(e *httpexpect.Expect, counterPtr *uint32, expectedBodyStr string, nocache string) error {
	e.GET("/").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	time.Sleep(cacheDuration / 5) // lets wait for a while, cache should be saved and ready
	e.GET("/").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	counter := atomic.LoadUint32(counterPtr)
	if counter > 1 {
		// n should be 1 because it doesn't changed after the first call
		return errTestFailed.Format(1, counter)
	}
	time.Sleep(cacheDuration)

	// cache should be cleared now
	e.GET("/").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	time.Sleep(cacheDuration / 5)
	// let's call again , the cache should be saved
	e.GET("/").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	counter = atomic.LoadUint32(counterPtr)
	if counter != 2 {
		return errTestFailed.Format(2, counter)
	}

	// we have cache response saved for the "/" path, we have some time more here, but here
	// we will make the requestS with some of the deniers options
	e.GET("/").WithHeader("max-age", "0").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	e.GET("/").WithHeader("Authorization", "basic or anything").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	counter = atomic.LoadUint32(counterPtr)
	if counter != 4 {
		return errTestFailed.Format(4, counter)
	}

	if nocache != "" {
		// test the NoCache, first sleep to pass the cache expiration,
		// second add to the cache with a valid request and response
		// third, do it with the "/nocache" path (static for now, pure test design) given by the consumer
		time.Sleep(cacheDuration)

		// cache should be cleared now, this should work because we are not in the "nocache" path
		e.GET("/").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr) // counter = 5
		time.Sleep(cacheDuration / 5)

		// let's call the "nocache", the expiration is not passed so but the "nocache"
		// route's path has the cache.NoCache so it should be not cached and the counter should be ++
		e.GET(nocache).Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr) // counter should be 6
		counter = atomic.LoadUint32(counterPtr)
		if counter != 6 { // 4 before, 5 with the first call to store the cache, and six with the no cache, again original handler executation
			return errTestFailed.Format(6, counter)
		}

		// let's call again the "/", the expiration is not passed so  it should be cached
		e.GET("/").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
		counter = atomic.LoadUint32(counterPtr)
		if counter != 6 {
			return errTestFailed.Format(6, counter)
		}

		// but now check for the No
	}

	return nil
}

func TestNoCache(t *testing.T) {
	app := siris.New()
	var n uint32

	app.Get("/", cache.WrapHandler(func(ctx context.Context) {
		atomic.AddUint32(&n, 1)
		ctx.Write([]byte(expectedBodyStr))
	}, cacheDuration))

	app.Get("/nocache", cache.WrapHandler(func(ctx context.Context) {
		cache.NoCache(ctx) // <----
		atomic.AddUint32(&n, 1)
		ctx.Write([]byte(expectedBodyStr))
	}, cacheDuration))

	e := httptest.New(t, app)
	if err := runTest(e, &n, expectedBodyStr, "/nocache"); err != nil {
		t.Fatalf(t.Name()+": %v", err)
	}

}

func TestCache(t *testing.T) {
	app := siris.New()
	var n uint32

	app.Use(cache.Handler(cacheDuration))

	app.Get("/", func(ctx context.Context) {
		atomic.AddUint32(&n, 1)
		ctx.Write([]byte(expectedBodyStr))
	})

	e := httptest.New(t, app)
	if err := runTest(e, &n, expectedBodyStr, ""); err != nil {
		t.Fatalf(t.Name()+": %v", err)
	}

}

func TestCacheHandlerParallel(t *testing.T) {
	t.Parallel()
	TestCache(t)
}

func TestCacheValidator(t *testing.T) {
	app := siris.New()
	var n uint32

	h := func(ctx context.Context) {
		atomic.AddUint32(&n, 1)
		ctx.Write([]byte(expectedBodyStr))
	}

	validCache := cache.Cache(h, cacheDuration)
	app.Get("/", validCache.ServeHTTP)

	managedCache := cache.Cache(h, cacheDuration)
	managedCache.AddRule(rule.Validator([]rule.PreValidator{
		func(ctx context.Context) bool {
			if ctx.Request().URL.Path == "/invalid" {
				return false // should always invalid for cache, don't bother to go to try to get or set cache
			}
			return true
		},
	}, nil))

	managedCache2 := cache.Cache(func(ctx context.Context) {
		atomic.AddUint32(&n, 1)
		ctx.Header("DONT", "DO not cache that response even if it was claimed")
		ctx.Write([]byte(expectedBodyStr))

	}, cacheDuration)
	managedCache2.AddRule(rule.Validator(nil,
		[]rule.PostValidator{
			func(ctx context.Context) bool {
				if ctx.ResponseWriter().Header().Get("DONT") != "" {
					return false // it's passed the Claim and now Valid checks if the response contains a header of "DONT"
				}
				return true
			},
		},
	))

	app.Get("/valid", validCache.ServeHTTP)

	app.Get("/invalid", managedCache.ServeHTTP)
	app.Get("/invalid2", managedCache2.ServeHTTP)

	e := httptest.New(t, app)

	// execute from cache the next time
	e.GET("/valid").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	time.Sleep(cacheDuration / 5) // lets wait for a while, cache should be saved and ready
	e.GET("/valid").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)
	counter := atomic.LoadUint32(&n)
	if counter > 1 {
		// n should be 1 because it doesn't changed after the first call
		t.Fatal(errTestFailed.Format(1, counter))
	}
	// don't execute from cache, execute the original, counter should ++ here
	e.GET("/invalid").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr)  // counter = 2
	e.GET("/invalid2").Expect().Status(http.StatusOK).Body().Equal(expectedBodyStr) // counter = 3

	counter = atomic.LoadUint32(&n)
	if counter != 3 {
		// n should be 1 because it doesn't changed after the first call
		t.Fatalf(t.Name()+": %v", errTestFailed.Format(3, counter))
	}
}
