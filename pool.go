package mnemo

import (
	"fmt"
	"sync"
)

// Pool is a collection of Conns
type Pool struct {
	mu    sync.Mutex
	conns map[interface{}]*Conn
}

// NewPool creates a new pool of Conns
func NewPool() *Pool {
	return &Pool{
		conns: make(map[interface{}]*Conn),
	}
}

// Conns returns a map of the pool's Conns
func (p *Pool) Conns() map[interface{}]*Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conns
}

// AddConn adds a Conn to the pool
func (p *Pool) AddConn(c *Conn) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.conns[c.Key]
	if ok {
		return fmt.Errorf("connection with key %v already exists", c.Key)
	}
	p.conns[c.Key] = c
	c.Pool = p
	return nil
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.conns {
		c.Close()
	}
}

// RemoveConn removes a Conn from the pool
func (p *Pool) removeConnection(c *Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.conns, c)
}
