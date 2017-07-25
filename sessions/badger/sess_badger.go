// Copyright 2017 SIRIS Authors. All Rights Reserved.

// Package badger for session provider
//
// depend on github.com/dgraph-io/badger
//
// go install github.com/dgraph-io/badger
//
// Usage:
// import(
//   _ "github.com/go-siris/siris/sessions/badger"
//   "github.com/go-siris/siris/sessions"
// )
//
//	func init() {
//		globalSessions, _ = sessions.NewManager("badger", {"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"127.0.0.1:7070"}`)
//		go globalSessions.GC()
//	}
//
// more docs: http://beego.me/docs/module/sessions.md
package badger

import (
	"net/http"
	"strconv"
	"strings"
	"io/ioutil"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/go-siris/siris/sessions"
)

var badgerpder = &Provider{}

// SessionStore redis session store
type SessionStore struct {
	p           *badger.KV
	sid         string
	lock        sync.RWMutex
	values      map[interface{}]interface{}
	maxlifetime int64
}

// Set value in redis session
func (bs *SessionStore) Set(key, value interface{}) error {
	bs.lock.Lock()
	bs.values[key] = value
	bs.lock.Unlock()
	return nil
}

// Get value in redis session
func (bs *SessionStore) Get(key interface{}) interface{} {
	bs.lock.RLock()
	if v, ok := bs.values[key]; ok {
		bs.lock.RUnlock()
		return v
	}
	bs.lock.RUnlock()
	return nil
}

// Delete value in redis session
func (bs *SessionStore) Delete(key interface{}) error {
	bs.lock.Lock()
	delete(bs.values, key)
	bs.lock.Unlock()
	return nil
}

// Flush clear all values in redis session
func (bs *SessionStore) Flush() error {
	bs.lock.Lock()
	bs.values = make(map[interface{}]interface{})
	bs.lock.Unlock()
	return nil
}

// SessionID get redis session id
func (bs *SessionStore) SessionID() string {
	return bs.sid
}

// SessionRelease save session values to redis
func (bs *SessionStore) SessionRelease(w http.ResponseWriter) {
	b, err := sessions.EncodeGob(bs.values)
	if err != nil {
		return
	}
	c := bs.p.Get()
	defer c.Close()
	c.Do("SETEX", bs.sid, bs.maxlifetime, string(b))
}

// Provider redis session provider
type Provider struct {
	maxlifetime int64
	savePath    string
	kv					*badger.KV
}

// SessionInit init badger session
// configString like json string with options
// https://godoc.org/github.com/dgraph-io/badger#Options
func (bp *Provider) SessionInit(maxlifetime int64, configString string) error {
	bp.maxlifetime = maxlifetime
	opt := badger.DefaultOptions
	dir, _ := ioutil.TempDir("/tmp", "badger")
	opt.Dir = dir
	opt.ValueDir = dir
	bp.kv, err := badger.NewKV(&opt)
	return err
}

// SessionRead read redis session by sid
func (bp *Provider) SessionRead(sid string) (sessions.Store, error) {
	c := rp.poollist.Get()
	defer c.Close()

	var kv map[interface{}]interface{}

	kvs, err := redis.String(c.Do("GET", sid))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		if kv, err = sessions.DecodeGob([]byte(kvs)); err != nil {
			return nil, err
		}
	}

	bs := &SessionStore{p: rp.poollist, sid: sid, values: kv, maxlifetime: rp.maxlifetime}
	return bs, nil
}

// SessionExist check redis session exist by sid
func (bp *Provider) SessionExist(sid string) bool {
	c := rp.poollist.Get()
	defer c.Close()

	if existed, err := redis.Int(c.Do("EXISTS", sid)); err != nil || existed == 0 {
		return false
	}
	return true
}

// SessionRegenerate generate new sid for redis session
func (bp *Provider) SessionRegenerate(oldsid, sid string) (sessions.Store, error) {
	c := rp.poollist.Get()
	defer c.Close()

	if existed, _ := redis.Int(c.Do("EXISTS", oldsid)); existed == 0 {
		// oldsid doesn't exists, set the new sid directly
		// ignore error here, since if it return error
		// the existed value will be 0
		c.Do("SET", sid, "", "EX", rp.maxlifetime)
	} else {
		c.Do("RENAME", oldsid, sid)
		c.Do("EXPIRE", sid, rp.maxlifetime)
	}
	return rp.SessionRead(sid)
}

// SessionDestroy delete redis session by id
func (bp *Provider) SessionDestroy(sid string) error {
	c := rp.poollist.Get()
	defer c.Close()

	c.Do("DEL", sid)
	return nil
}

// SessionGC Impelment method, no used.
func (bp *Provider) SessionGC() {
	opt := badger.DefaultIteratorOptions
	itr := bp.kv.NewIterator(opt)
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		key := item.Key()
		val := item.Value() // This could block while value is fetched from value log.
		// For key only iteration, set opt.FetchValues to false, and don't call
		// item.Value().

		// TODO: Delete keys that not used maxlifetime
		// value touch < time - maxlifetime

		// Remember that both key, val would become invalid in the next iteration of the loop.
		// So, if you need access to them outside, copy them or parse them.
	}
	itr.Close()
}

// SessionAll return all activeSession
func (bp *Provider) SessionAll() int {
	count := 0
	opt := badger.DefaultIteratorOptions
	opt.FetchValues = false
	itr := bp.kv.NewIterator(opt)
	for itr.Rewind(); itr.Valid(); itr.Next() {
		count++
	}
	itr.Close()
	return count
}

func init() {
	sessions.Register("badger", badgerpder)
}
