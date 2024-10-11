package mnemo

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

type (
	// cache is a collection of data that can be cached and reduced.
	Cache[T any] struct {
		mu        sync.Mutex
		createdAt time.Time
		raw       *raw[T]
		reducer   *reducer[T]
	}
	// raw is a collection of cached data, it's history, and a feed of live updates
	// prior to reduction.
	raw[T any] struct {
		caches  map[CacheKey]*Item[T]
		history map[time.Time]map[CacheKey]Item[T]
		feed    chan map[time.Time]map[CacheKey]Item[T]
	}
	// reducer is a collection of reduced data, it's history, and a feed of live updates
	reducer[T any] struct {
		reduce  *func([]reducerCache[T]) []reducerCache[any]
		history map[time.Time][]reducerCache[any]
		feed    chan reducerFeed[any]
	}
	// Item holds cached data and the time it was cached.
	Item[T any] struct {
		CreatedAt time.Time `json:"created_at"`
		Data      *T        `json:"data"`
	}
	// cacheTimeoutConfig is a configuration for caching data with a timeout.
	cacheTimeoutConfig[T any] struct {
		data       *T
		key        any
		timeoutFun func(data *T)
		timeout    time.Duration
	}
	// CacheKey is a unique identifier for a cache.
	CacheKey any
	// ItemKey is a unique identifier for an item.
	ItemKey any
	// ReducerFunc takes a cache and returns a reduced version of it.
	//
	// It is run against the raw cache on every change and must return json serializable data.
	ReducerFunc[T any, U any] func(state T) (mutation U)
	// reducerCache wraps a reduced cache with it's key and creation time.
	reducerCache[U any] struct {
		Key       CacheKey  `json:"key"`
		CreatedAt time.Time `json:"created_at"`
		Data      U         `json:"data"`
	}
	// reducerFeed is sent to the reducer's feed channel on every change.
	reducerFeed[U any] struct {
		CreatedAt time.Time         `json:"created_at"`
		Cache     []reducerCache[U] `json:"cache"`
	}
	reducerHistory[T any] reducerFeed[T]
)

// newCache is an internal implementation of NewCache
func newCache[T any]() (data *Cache[T]) {
	c := &Cache[T]{
		createdAt: time.Now(),
		raw: &raw[T]{
			caches:  make(map[CacheKey]*Item[T]),
			history: make(map[time.Time]map[CacheKey]Item[T]),
			feed:    make(chan map[time.Time]map[CacheKey]Item[T], 1024),
		},
		reducer: &reducer[T]{
			history: make(map[time.Time][]reducerCache[any]),
			feed:    make(chan reducerFeed[any], 1024),
		},
	}
	return c
}

// monitorChanges monitors changes to the raw cache and caches the raw cache and it's reduction.
func (c *Cache[T]) monitorChanges(setup chan bool) {
	// cache initial state and confirm setup is complete
	pRaw := c.copyRaw()
	setup <- true

	prev := c.reduce(pRaw)
	t := time.Now()
	c.cacheRaw(t, pRaw)
	c.cacheReduction(t, prev)
	for {
		raw := c.copyRaw()
		current := c.reduce(raw)
		// TODO: Maybe be able to reduce this to a single comparison
		// by converting the reduced cache to a string
		if !reflect.DeepEqual(prev, current) {
			t := time.Now()
			c.cacheRaw(t, raw)
			c.cacheReduction(t, current)
			prev = current
		}
	}
}

// copyRaw copies the raw cache.
func (c *Cache[T]) copyRaw() map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	copy := make(map[CacheKey]Item[T])
	for k, v := range c.raw.caches {
		copy[k] = *v
	}
	return copy
}

// cacheRaw caches the raw cache.
func (c *Cache[T]) cacheRaw(t time.Time, copy map[CacheKey]Item[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.raw.history[t] = copy
	c.raw.feed <- c.raw.history
}

// newCacheReducer wraps a user defined reducer function with reducerCache meta data.
func newCacheReducer[T, U any](f ReducerFunc[T, U]) func([]reducerCache[T]) []reducerCache[U] {
	return func(rdt []reducerCache[T]) []reducerCache[U] {
		rdu := []reducerCache[U]{}
		for _, d := range rdt {
			rdu = append(rdu, reducerCache[U]{
				Key:       d.Key,
				CreatedAt: d.CreatedAt,
				Data:      f(d.Data),
			})
		}
		return rdu
	}
}

// reduce implements the user defined reducer function.
func (c *Cache[T]) reduce(copy map[CacheKey]Item[T]) []reducerCache[any] {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := []reducerCache[T]{}
	for key, item := range copy {
		data = append(data, reducerCache[T]{
			Key:       key,
			CreatedAt: item.CreatedAt,
			Data:      *item.Data,
		})
	}
	reduce := *c.reducer.reduce
	return reduce(data)
}

// cacheReduction caches the reduced cache.
func (c *Cache[T]) cacheReduction(t time.Time, r []reducerCache[any]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//sort by createdAt
	sort.Slice(r, func(i, j int) bool {
		return r[i].CreatedAt.Before(r[j].CreatedAt)
	})
	c.reducer.history[t] = r
	rf := reducerFeed[any]{CreatedAt: t, Cache: r}
	c.reducer.feed <- rf
}

// SetReducer sets the user defined reducer function and starts monitoring changes.
//
// Setting a reducer is mandatory for triggering change monitoring. DefaultReducer is available but
// merely returns the raw cache so it is not recommended if you are caching complex data types that cannot
// not be serialized to json.
func (c *Cache[T]) SetReducer(rf ReducerFunc[T, any]) {
	cr := newCacheReducer(rf)
	c.mu.Lock()
	c.reducer.reduce = &cr
	c.mu.Unlock()

	setup := make(chan bool)
	defer close(setup)
	go c.monitorChanges(setup)
	<-setup
}

// DefaultReducer is a reducer that returns the raw cache.
func (c *Cache[T]) DefaultReducer(state T) (mutation any) {
	return state
}

// RawFeed returns a channel of the raw cache updates.
func (c *Cache[T]) RawFeed() chan map[time.Time]map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.feed
}

// ReducerFeed returns a channel of the reduced cache updates.
func (c *Cache[T]) ReducerFeed() chan reducerFeed[any] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reducer.feed
}

// RawHistory returns the raw cache history.
func (c *Cache[T]) RawHistory() map[time.Time]map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.raw.history
}

// ReducerHistory returns the reduced cache history.
func (c *Cache[T]) ReducerHistory() []reducerHistory[any] {
	c.mu.Lock()
	defer c.mu.Unlock()
	rh := []reducerHistory[any]{}
	for time, cache := range c.reducer.history {
		rh = append(rh, reducerHistory[any]{
			CreatedAt: time,
			Cache:     cache,
		})
	}
	return rh
}

// Get returns a cache by key.
func (c *Cache[T]) Get(key CacheKey) (Item[T], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := c.raw.caches[key]
	if data == nil {
		return *new(Item[T]), false
	}
	return *data, true
}

// GetAll returns all caches.
func (c *Cache[T]) GetAll() map[CacheKey]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := make(map[CacheKey]Item[T])
	for key, item := range c.raw.caches {
		cache[key] = *item
	}
	return cache
}

// Cache caches data by key.
func (c *Cache[T]) Cache(key CacheKey, data *T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.raw.caches[key] != nil {
		return fmt.Errorf("duplicate cache key: %v", key)
	}
	item := &Item[T]{
		Data:      data,
		CreatedAt: time.Now(),
	}
	c.raw.caches[key] = item
	return nil
}

// TODO: Convert to option
func (c *Cache[T]) CacheWithTimeout(cfg cacheTimeoutConfig[T]) error {
	c.Cache(cfg.key, cfg.data)
	if !(cfg.timeout > time.Second*0) {
		return fmt.Errorf(
			"cache not set for timeout: %v; timeout must be greater than 0", cfg.timeout,
		)
	}

	go func() {
		timer := time.NewTimer(cfg.timeout)
		<-timer.C
		item, ok := c.Get(cfg.key)
		if !ok {
			logger.Fatalf("could not get cache with key %v", cfg.key)
		}
		err := c.Delete(cfg.key)
		if err != nil {
			NewError[Cache[T]](err.Error())
			return
		}
		cfg.timeoutFun(item.Data)
	}()
	return nil
}

// Update updates a cache with a new value. It returns false if the cache does not exist.
func (c *Cache[T]) Update(key CacheKey, update T) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	prev, ok := c.raw.caches[key]
	if !ok {
		return false
	}
	//TODO: ensure this is being updated in reducer
	c.raw.caches[key] = &Item[T]{Data: &update, CreatedAt: prev.CreatedAt}
	return true
}

// Delete deletes a cache by key.
func (c *Cache[T]) Delete(key interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.raw.caches[key] == nil {
		return fmt.Errorf("no cache with key: %v", key)
	}
	delete(c.raw.caches, key)
	return nil
}

// NewCacheTimeoutConfig creates a new cacheTimeoutConfig.
// TODO: Convert to option
func NewCacheTimeoutConfig[T any](
	data *T,
	key interface{},
	timeoutFun func(data *T),
	timeout time.Duration,
) cacheTimeoutConfig[T] {
	return cacheTimeoutConfig[T]{
		data:       data,
		key:        key,
		timeoutFun: timeoutFun,
		timeout:    timeout,
	}
}
