package qs_test

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/pasztorpisti/qs"
)

// This example shows how to speed up marshaling by implementing a cached
// ValuesMarshalerFactory. The cache stores the ValuesMarshaler objects
// created by the original factory so marshaling the same type multiple
// times requires creating the ValuesMarshaler object only once.
//
// Note that you can implement caching for 4 different factories:
// MarshalOptions.ValuesMarshalerFactory, MarshalOptions.MarshalerFactory,
// UnmarshalOptions.ValuesUnmarshalerFactory and UnmarshalOptions.UnmarshalerFactory.
// In this example we cache only the ValuesMarshalerFactory.
func Example_caching() {
	cachedFactory := &valuesMarshalerFactory{
		factory: qs.DefaultValuesMarshalerFactory,
		cache:   map[reflect.Type]qs.ValuesMarshaler{},
	}
	marshaler := qs.NewMarshaler(&qs.MarshalOptions{
		ValuesMarshalerFactory: cachedFactory,
	})

	type Query struct {
		Search   string
		Page     int
		PageSize int
	}

	// this will be a cache miss
	queryStr, err := marshaler.Marshal(&Query{})
	fmt.Println("Marshal-Result:", queryStr, err)

	// this will be a cache hit
	queryStr, err = marshaler.Marshal(&Query{
		Search:   "my search",
		Page:     2,
		PageSize: 50,
	})
	fmt.Println("Marshal-Result:", queryStr, err)

	var q Query
	err = qs.Unmarshal(&q, queryStr)
	fmt.Println("Unmarshal-Result:", q, err)

	// Output:
	// cache miss: qs_test.Query
	// Marshal-Result: page=0&page_size=0&search= <nil>
	// cache hit: qs_test.Query
	// Marshal-Result: page=2&page_size=50&search=my+search <nil>
	// Unmarshal-Result: {my search 2 50} <nil>
}

// valuesMarshalerFactory implements the qs.ValuesMarshalerFactory interface
// and serves as a caching proxy to another ValuesMarshalerFactory.
type valuesMarshalerFactory struct {
	factory qs.ValuesMarshalerFactory
	cache   map[reflect.Type]qs.ValuesMarshaler
	mu      sync.RWMutex
}

func (p *valuesMarshalerFactory) ValuesMarshaler(t reflect.Type,
	opts *qs.MarshalOptions) (qs.ValuesMarshaler, error) {
	// Note: If you use the cache only from one goroutine then you can
	// optimise this by removing the mutex from this implementation.
	p.mu.RLock()
	vm, ok := p.cache[t]
	p.mu.RUnlock()
	if ok {
		// Of course this logging wouldn't be concurrency safe,
		// we put it here only for the sake of the example output check.
		fmt.Printf("cache hit: %v\n", t)
		return vm, nil
	}
	fmt.Printf("cache miss: %v\n", t)
	vm, err := p.factory.ValuesMarshaler(t, opts)
	if err == nil {
		p.mu.Lock()
		p.cache[t] = vm
		p.mu.Unlock()
	}
	return vm, err
}
