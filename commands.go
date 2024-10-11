package mnemo

import (
	"fmt"
	"sync"
)

type (
	// CommandKey is a unique identifier for a command.
	CommandKey string
	// Commands is a collection of commands.
	Commands struct {
		mu   sync.Mutex
		list map[CommandKey]func()
	}
)

// NewCommands creates a new collection of commands.
func NewCommands() Commands {
	return Commands{
		list: make(map[CommandKey]func()),
	}
}

// Assign assigns a map of commands to the collection.
func (c *Commands) Assign(cmds map[CommandKey]func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range cmds {
		c.list[k] = v
	}
}

// Execute executes a command and returns an error if the command does not exist.
func (c *Commands) Execute(key CommandKey) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, ok := c.list[key]
	if !ok {
		return fmt.Errorf("no command with key %v", key)
	}
	cmd := f
	cmd()
	return nil
}

func (c *Commands) List() map[CommandKey]func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list
}
