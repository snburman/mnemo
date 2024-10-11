package mnemo

import (
	"fmt"
	"testing"
)

func Example() {
	// Create some keys for stores, caches, and commands
	var (
		ExampleStoreKey StoreKey   = "example_store"
		ExampleCacheKey CacheKey   = "example_cache"
		ExampleCmdKey   CommandKey = "example_cmd"
	)

	m := New()
	// Add optional configuration for one server per instance
	// Instances do not require a server to access stores, caches, or commands
	// However, a server is required to access the Mnemo API with a client
	m.WithServer("example", WithPattern("/example"), WithPort(8080))

	// Create a new store
	store, _ := NewStore(ExampleStoreKey)

	// Add commands to the store
	cmd := store.Commands()
	cmd.Assign(map[CommandKey]func(){
		ExampleCmdKey: func() {
			fmt.Println("I'm a command!")
		},
	})

	// Create a type to cache
	type Message struct {
		Msg string
	}

	// Create a new cache for messages
	cache, _ := NewCache[Message](ExampleStoreKey, ExampleCacheKey)

	// Cache a message
	cache.Cache(ExampleCacheKey, &Message{Msg: "Hello, Mnemo!"})

	// Set a reducer for the cache
	// A reducer is a function that takes the current state of the cache and returns a mutation
	// The mutation is then cached and sent to the reducer's feed channel
	cache.SetReducer(func(state Message) (mutation any) {
		return state.Msg + " With a reducer!"
	})

	// The integrity of the original cache is maintained and stored by it's creation time
	_ = cache.RawHistory()

	// Add store(s) by key
	// This can be done at any time, however, it is recommended to do so before attempting to access the store
	// Unique stores may only be added to one instance of Mnemo, although you may have multiple instances
	m.WithStores(ExampleStoreKey)

	// Access stores or caches by key from anywhere in the application
	myStore, _ := UseStore(ExampleStoreKey)

	// Access and execute commands from anywhere
	myCmds := myStore.Commands()
	myCmds.Execute(ExampleCmdKey)

	// List commands if needed
	_ = myCmds.List()

	// Access caches from anywhere
	myCache, _ := UseCache[Message](ExampleStoreKey, ExampleCacheKey)

	// Access data from the cache
	item, _ := myCache.Get(ExampleCacheKey)
	fmt.Println(item.Data.Msg)

	// Update the cache
	myCache.Update(ExampleCacheKey, Message{Msg: "Hello, Mnemo! With an update!"})

	// Server is a wrapper around http.Server and is non-blocking
	m.Server().ListenAndServe()

	// Output:
	// I'm a command!
	// Hello, Mnemo! With a reducer!
}

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("Expected a non-nil Mnemo instance, but got nil")
	}
	// Add additional assertions for the Mnemo instance if needed
}

func TestMnemo_WithServer(t *testing.T) {
	m := New()
	m.WithServer("test", WithPattern("/test"), WithPort(8080))
	if m.server == nil {
		t.Error("Expected a non-nil Server instance, but got nil")
	}

	if m.server.cfg.Pattern != "/test" {
		t.Errorf("Expected server pattern to be '/test', but got %s", m.server.cfg.Pattern)
	}

	if m.server.cfg.Port != 8080 {
		t.Errorf("Expected server port to be 8080, but got %d", m.server.cfg.Port)
	}
}
