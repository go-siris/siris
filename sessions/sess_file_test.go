// Copyright 2017 Go-SIRIS Author. All Rights Reserved.

package sessions

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/gavv/httpexpect.v1"
)

func TestSessionFile(t *testing.T) {
	config := `{"cookieName":"gosessionid","gclifetime":600,"cookieLifeTime":1800,"domain":"example.com","EnableSidInURLQuery":true,"enableSetCookie":true,"EnableSidInHTTPHeader":true,"SessionNameInHTTPHeader":"Gosessionid","providerConfig":"./tmp"}`
	conf := new(ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode error", err)
	}
	globalSessions, _ := NewManager("file", conf)
	go globalSessions.GC()

	// Routes:
	//  GET /session           			get first sessionId
	//  GET /session/same-session-id   	get same sessionId again
	mux := http.NewServeMux()

	var sessionId1 string
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		sess, err := globalSessions.SessionStart(w, r)
		if err != nil {
			t.Fatal("start error,", err)
		}
		sessionId1 = sess.SessionID()
		sess.SessionRelease(w)
	})

	var sessionId2 string
	mux.HandleFunc("/session/same-session-id", func(w http.ResponseWriter, r *http.Request) {
		sess, err := globalSessions.SessionStart(w, r)
		if err != nil {
			t.Fatal("start error,", err)
		}
		sessionId2 = sess.SessionID()
		sess.SessionRelease(w)
	})

	var sessionId3 string
	mux.HandleFunc("/session/destroy", func(w http.ResponseWriter, r *http.Request) {
		sess, err := globalSessions.SessionStart(w, r)
		if err != nil {
			t.Fatal("start error,", err)
		}
		sessionId3 = sess.SessionID()
		sess.SessionRelease(w)
		globalSessions.SessionDestroy(w, r)
	})

	var sessionId4 string
	mux.HandleFunc("/session/getnew", func(w http.ResponseWriter, r *http.Request) {
		sess, err := globalSessions.SessionStart(w, r)
		if err != nil {
			t.Fatal("start error,", err)
		}
		sessionId4 = sess.SessionID()
		sess.SessionRelease(w)
	})

	var sessionId5 string
	mux.HandleFunc("/session/regenerate", func(w http.ResponseWriter, r *http.Request) {
		sess, err := globalSessions.SessionStart(w, r)
		if err != nil {
			t.Fatal("start error,", err)
		}
		sess = globalSessions.SessionRegenerateID(w, r)
		sessionId5 = sess.SessionID()
		sess.SessionRelease(w)
	})

	var sessionId6 string
	mux.HandleFunc("/session/functions-test", func(w http.ResponseWriter, r *http.Request) {
		sess, err := globalSessions.SessionStart(w, r)
		if err != nil {
			t.Fatal("start error,", err)
		}
		err = sess.Set("username", "astaxie")
		if err != nil {
			t.Fatal("set error,", err)
		}
		err = sess.Set("username3", "astaxie3")
		if err != nil {
			t.Fatal("set error,", err)
		}
		if username := sess.Get("username"); username != "astaxie" {
			t.Fatal("get username error")
		}
		if username2 := sess.Get("username2"); username2 != nil {
			t.Fatal("get username2 error")
		}
		err = sess.Delete("username3")
		if err != nil {
			t.Fatal("delete error,", err)
		}
		if username3 := sess.Get("username3"); username3 != nil {
			t.Fatal("get !username3 error")
		}
		err = sess.Flush()
		if err != nil {
			t.Fatal("flush error,", err)
		}
		if username4 := sess.Get("username"); username4 != nil {
			t.Fatal("flush not work as excepted")
		}
		sessionId6 = sess.SessionID()
		sess.SessionRelease(w)
	})

	// invoke http.Handler directly using httpexpect.Binder
	var handler http.Handler = mux

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://example.com",
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
			Timeout:   time.Second * 30,
		},
		// use verbose logging
		Printers: []httpexpect.Printer{
			httpexpect.NewCurlPrinter(t),
			httpexpect.NewDebugPrinter(t, false),
		},
	})

	t.Log("/session: Test Start")
	// get first session cookie
	e.GET("/session").
		Expect().
		Status(http.StatusOK)
	t.Log("/session: Test Done\n")

	t.Log("/session/same-session-id: Test Start")
	// check cookies
	e.GET("/session/same-session-id").
		WithCookie("gosessionid", sessionId1).
		Expect().
		Status(http.StatusOK)
	if sessionId1 != sessionId2 {
		t.Fatalf("sessionIds not Match: %s != %s", sessionId1, sessionId2)
	}
	t.Log("/session/same-session-id: Test Done\n")

	t.Log("/session/destroy: Test Start")
	// check cookies destroyed
	resp := e.GET("/session/destroy").
		WithCookie("gosessionid", sessionId2).
		Expect().
		Status(http.StatusOK)
	cookie := resp.Cookie("gosessionid")
	cookie.Domain().Equal("example.com")
	cookie.Path().Equal("/")
	cookie.Expires().Equal(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
	if sessionId1 != sessionId2 {
		t.Fatalf("sessionIds not Match: %s != %s", sessionId1, sessionId2)
	}
	t.Log("/session/destroy: Test Done\n")

	t.Log("/session/getnew: Test Start")
	// get new session cookie
	resp = e.GET("/session/getnew").
		WithCookie("gosessionid", sessionId3).
		Expect().
		Status(http.StatusOK)
	cookie = resp.Cookie("gosessionid")
	cookie.Domain().Equal("example.com")
	cookie.Path().Equal("/")
	if sessionId4 == sessionId3 {
		t.Fatalf("sessionIds3/4 Match but not allowed: %s == %s", sessionId3, sessionId4)
	}
	t.Log("/session/getnew: Test Done\n")

	t.Log("/session/regenerate: Test Start")
	// get regenerated session cookie
	resp = e.GET("/session/regenerate").
		WithCookie("gosessionid", sessionId4).
		Expect().
		Status(http.StatusOK)
	cookie = resp.Cookie("gosessionid")
	cookie.Domain().Equal("example.com")
	cookie.Path().Equal("/")
	if sessionId4 == sessionId5 {
		t.Fatalf("sessionIds4/5 Match but not allowed: %s != %s", sessionId4, sessionId5)
	}
	t.Log("/session/regenerate: Test Done\n")

	t.Log("/session/functions-test: Test Start")
	// get regenerated session cookie
	resp = e.GET("/session/functions-test").
		WithCookie("gosessionid", sessionId4).
		Expect().
		Status(http.StatusOK)
	cookie = resp.Cookie("gosessionid")
	cookie.Domain().Equal("example.com")
	cookie.Path().Equal("/")
	if sessionId5 == sessionId6 {
		t.Fatalf("sessionIds5/6 Match but not allowed: %s != %s", sessionId5, sessionId6)
	}
	t.Log("/session/functions-test: Test Done\n")

	t.Log("globalSessions: Test Start")
	err := errors.New("err test")
	globalSessions.SetSecure(true)
	if activeSessions := globalSessions.GetActiveSession(); activeSessions == 0 {
		t.Fatal("GetActiveSession returns 0")
	}
	if _, err = globalSessions.GetSessionStore(sessionId1); err != nil {
		t.Fatalf("GetSessionStore returns err, %s", err)
	}
	newseddion, _ := globalSessions.sessionID()
	_, err = globalSessions.provider.SessionRegenerate("notfound1234", newseddion)
	if err != nil {
		t.Fatalf("SessionRegenerate returns err, %s", err)
	}
	t.Log("globalSessions: Test Done\n")
}

func TestSessionFilePanic_Register(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	Register("file", nil)
}

func TestSessionFilePanic_RegisterDup(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	Register("file", filepder)
}

func TestSessionFilePanic_Manager(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()

	config := `{"cookieName":"gosessionid","gclifetime":600,"domain":"example.com","enableSetCookie":true,"EnableSidInHTTPHeader":true,"SessionNameInHTTPHeader":"","providerConfig":"./tmp"}`
	conf := new(ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode error", err)
	}
	_, _ = NewManager("file", conf)
}

func TestSessionFilePanic_Manager2(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()

	config := `{"cookieName":"gosessionid","gclifetime":600,"domain":"example.com","enableSetCookie":true,"EnableSidInHTTPHeader":true,"SessionNameInHTTPHeader":"gosessionid","providerConfig":"./tmp"}`
	conf := new(ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode error", err)
	}
	_, _ = NewManager("file", conf)
}
