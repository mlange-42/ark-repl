package repl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type cmd struct {
	Sub subCmd
}

func (c cmd) Execute(repl *Repl, out *strings.Builder) {}
func (c cmd) Help(repl *Repl, out *strings.Builder)    {}

type subCmd struct {
	SubSub subSubCmd
}

func (c subCmd) Execute(repl *Repl, out *strings.Builder) {}
func (c subCmd) Help(repl *Repl, out *strings.Builder)    {}

type subSubCmd struct {
	Arg1 bool
	Arg2 int
	Arg3 float64
	Arg4 string
}

func (c subSubCmd) Execute(repl *Repl, out *strings.Builder) {}
func (c subSubCmd) Help(repl *Repl, out *strings.Builder)    {}

func TestParser(t *testing.T) {
	allCommands := map[string]command{
		"help": hlp{},
		"cmd":  cmd{},
	}

	cmdString := "cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, help, err := parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.False(t, help)
	assert.Equal(t, "repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:\"test\"}", fmt.Sprintf("%#v", out))

	cmdString = "help cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, help, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.True(t, help)
	assert.Equal(t, "repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:\"test\"}", fmt.Sprintf("%#v", out))
}
