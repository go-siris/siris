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
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	obadger "github.com/dgraph-io/badger"
	"github.com/go-siris/siris/sessions"
	"github.com/go-siris/siris/sessions/utils"
)

var badgerpder = &Provider{}

// SessionStore redis session store
type SessionStore struct {
	kv          *obadger.KV
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
	tostore := map[interface{}]interface{}{"Expire": time.Now().Add(time.Duration(bs.maxlifetime) * time.Second), "Values": bs.values}
	b, err := utils.EncodeGob(tostore)
	if err != nil {
		return
	}
	bs.kv.Set([]byte(bs.sid), b)
	bs.kv.Close()
}

// Provider redis session provider
type Provider struct {
	maxlifetime int64
	savePath    string
	kv          *obadger.KV
}

// SessionInit init badger session
// configString like json string with options
// https://godoc.org/github.com/dgraph-io/badger#Options
func (bp *Provider) SessionInit(maxlifetime int64, configString string) error {
	bp.maxlifetime = maxlifetime
	opt := obadger.DefaultOptions
	dir, _ := ioutil.TempDir("", "badger_tmp")
	opt.Dir = dir
	opt.ValueDir = dir
	kv, err := obadger.NewKV(&opt)
	bp.kv = kv
	return err
}

// SessionRead read redis session by sid
func (bp *Provider) SessionRead(sid string) (sessions.Store, error) {
	var kv map[interface{}]interface{}
	var err error

	var item obadger.KVItem
	if err = bp.kv.Get([]byte(sid), &item); err != nil {
		return nil, err
	}
	kvs := item.Value()
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
		kv["Expire"] = time.Now().Add(time.Duration(bp.maxlifetime) * time.Second)
		kv["Values"] = make(map[interface{}]interface{})
	} else if kv, err = utils.DecodeGob(kvs); err != nil {
		return nil, err
	}
	values := kv["Values"].(map[interface{}]interface{})

	bs := &SessionStore{kv: bp.kv, sid: sid, values: values, maxlifetime: bp.maxlifetime}
	return bs, nil
}

// SessionExist check redis session exist by sid
func (bp *Provider) SessionExist(sid string) bool {
	if existed, err := bp.kv.Exists([]byte(sid)); err != nil || !existed {
		return false
	}
	return true
}

// SessionRegenerate generate new sid for badger session
func (bp *Provider) SessionRegenerate(oldsid, sid string) (sessions.Store, error) {
	var kv map[interface{}]interface{}
	var kv2 map[interface{}]interface{}
	var err error

	if existed, _ := bp.kv.Exists([]byte(sid)); !existed {
		// oldsid doesn't exists, set the new sid directly
		// ignore error here, since if it return error
		// the existed value will be 0
		kv = make(map[interface{}]interface{})
		kv["Expire"] = time.Now().Add(time.Duration(bp.maxlifetime) * time.Second)
		kv["Values"] = make(map[interface{}]interface{})
		b, err2 := utils.EncodeGob(kv)
		if err2 != nil {
			return nil, err2
		}
		bp.kv.Set([]byte(sid), b)
	} else {
		var item obadger.KVItem
		if err = bp.kv.Get([]byte(oldsid), &item); err != nil {
			return nil, err
		}
		kvs := item.Value()
		if len(kvs) == 0 {
			kv = make(map[interface{}]interface{})
			kv["Expire"] = time.Now().Add(time.Duration(bp.maxlifetime) * time.Second)
			kv["Values"] = make(map[interface{}]interface{})
		} else if kv, err = utils.DecodeGob(kvs); err != nil {
			return nil, err
		}
		values := kv["Values"].(map[interface{}]interface{})

		kv2["Expire"] = time.Now().Add(time.Duration(bp.maxlifetime) * time.Second)
		kv2["Values"] = values
		b, err := utils.EncodeGob(kv2)
		if err != nil {
			return nil, err
		}
		bp.kv.Set([]byte(sid), b)
		bp.kv.Delete([]byte(oldsid))
	}
	return bp.SessionRead(sid)
}

// SessionDestroy delete redis session by id
func (bp *Provider) SessionDestroy(sid string) error {
	bp.kv.Delete([]byte(sid))
	return nil
}

// SessionGC Impelment method, no used.
func (bp *Provider) SessionGC() {
	now := time.Now()
	opt := obadger.DefaultIteratorOptions
	itr := bp.kv.NewIterator(opt)
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		key := item.Key()
		val := item.Value() // This could block while value is fetched from value log.
		// For key only iteration, set opt.FetchValues to false, and don't call
		// item.Value().
		fmt.Printf("%#v -> %#v", key, val)

		values, _ := utils.DecodeGob(val)
		expires, ok := values["Expire"].(*time.Time)
		if ok {
			if b := expires.After(now); !b {
				bp.kv.Delete(key)
			}
		}

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
	opt := obadger.DefaultIteratorOptions
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
