// Package mnemo provides a simple and robust way to generate and manage caches for any data type.
package mnemo

import "sync"

type (
	// Mnemo is the main struct for the Mnemo package.
	Mnemo struct {
		mu     sync.Mutex
		server *Server
		logger Logger
		stores map[StoreKey]bool
	}
	Opt[T any] func(t *T)
)

// New returns a new Mnemo instance.
func New(opts ...Opt[Mnemo]) *Mnemo {
	m := &Mnemo{
		logger: logger,
		stores: make(map[StoreKey]bool),
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// WithServer creates a server for the Mnemo instance.
func (m *Mnemo) WithServer(key string, opts ...Opt[Server]) *Mnemo {
	srv, err := NewServer(key, opts...)
	if err != nil {
		NewError[Server](err.Error()).Log()
	}
	srv.withNemo(m)
	m.server = srv
	return m
}

// Server returns the Mnemo instance's server or panics if the server is nil.
func (m *Mnemo) Server() *Server {
	if m.server == nil {
		NewError[Mnemo]("server is nil").WithLogLevel(Panic).Log()
	}
	return m.server
}

// WithStore adds one or more stores to the Mnemo instance.
func (m *Mnemo) WithStores(keys ...StoreKey) *Mnemo {
	for _, k := range keys {
		m.stores[k] = true
	}
	return m
}

// UseStore returns a store from the Mnemo instance.
//
// Returns an error if the store does not exist or does not belong to the Mnemo instance.
func (m *Mnemo) UseStore(key StoreKey) (*Store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// if not a store of mnemo instance, find it in the store manager
	// and add it to the mnemo instance
	if _, ok := m.stores[key]; !ok {
		s, err := UseStore(key)
		if err != nil {
			return nil, err
		}
		if s.mnemo != m {
			return nil, NewError[Store]("store does not belong to mnemo instance")
		}
		m.stores[key] = true
		return s, nil
	}
	s, err := UseStore(key)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Mnemo) StoreKeys() map[StoreKey]bool {
	return m.stores
}

func (m *Mnemo) DetachStore(key StoreKey) {
	delete(m.stores, key)
}
