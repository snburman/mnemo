package mnemo

import "testing"

func TestNewCommands(t *testing.T) {
	c := NewCommands()
	if c.list == nil {
		t.Error("expected list to be initialized")
	}
}

func TestAssign(t *testing.T) {
	c := NewCommands()
	cmds := map[CommandKey]func(){
		"test": func() {},
	}
	c.Assign(cmds)
	if len(c.list) != len(cmds) {
		t.Error("expected commands to be assigned")
	}
}

func TestExecute(t *testing.T) {
	c := NewCommands()
	cmds := map[CommandKey]func(){
		"test": func() {},
	}
	c.Assign(cmds)
	err := c.Execute("test")
	if err != nil {
		t.Error("expected command to execute")
	}
	err = c.Execute("invalid")
	if err == nil {
		t.Error("expected command to not execute")
	}
}
