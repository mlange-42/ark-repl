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
func (c cmd) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprint(out, "Help text.")
}

type subCmd struct {
	SubSub subSubCmd
}

func (c subCmd) Execute(repl *Repl, out *strings.Builder) {}
func (c subCmd) Help(repl *Repl, out *strings.Builder)    {}

type subSubCmd struct {
	Arg1 bool    `help:"help text"`
	Arg2 int     `default:"5"`
	Arg3 float64 `default:"3.5"`
	Arg4 string  `default:"abc"`
}

func (c subSubCmd) Execute(repl *Repl, out *strings.Builder) {}
func (c subSubCmd) Help(repl *Repl, out *strings.Builder)    {}

func TestParser(t *testing.T) {
	allCommands := map[string]Command{
		"help": help{},
		"cmd":  cmd{},
	}

	cmdString := "cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, help, err := parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, help)
	assert.Equal(t, "repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:\"test\"}", fmt.Sprintf("%#v", out))

	cmdString = "cmd sub subsub"
	out, help, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, help)
	assert.Equal(t, "repl.subSubCmd{Arg1:false, Arg2:5, Arg3:3.5, Arg4:\"abc\"}", fmt.Sprintf("%#v", out))

	cmdString = "help cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, help, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.True(t, help)
	assert.Equal(t, "repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:\"test\"}", fmt.Sprintf("%#v", out))

	cmdString = "help"
	out, help, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, help)
	assert.Equal(t, "repl.help{}", fmt.Sprintf("%#v", out))
}

func TestParserListEntities(t *testing.T) {
	cmdString := "list entities with=Position,Velocity"
	out, help, err := parseInput(cmdString, defaultCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, help)
	assert.Equal(t, "repl.listEntities{N:25, With:[]string{\"Position\", \"Velocity\"}, Without:[]string(nil), Exclusive:false}", fmt.Sprintf("%#v", out))
}

func TestExtractHelp(t *testing.T) {
	out := strings.Builder{}
	repl := NewRepl(nil, Callbacks{})

	extractHelp(repl, cmd{}, &out)
	assert.Equal(t, `Help text.
Commands:
  sub
`, out.String())

	out = strings.Builder{}
	extractHelp(repl, subSubCmd{}, &out)
	assert.Equal(t, `
Options:
  arg1          bool     help text 
  arg2          int      Default: 5
  arg3          float    Default: 3.5
  arg4          string   Default: abc
`, out.String())
}
