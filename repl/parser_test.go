package repl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"github.com/stretchr/testify/assert"
)

type cmd struct {
	Sub subCmd
}

func (c cmd) Execute(world *ecs.World, out *strings.Builder) {}
func (c cmd) Help(out *strings.Builder) {
	fmt.Fprint(out, "Help text.")
}

type subCmd struct {
	SubSub subSubCmd
}

func (c subCmd) Execute(world *ecs.World, out *strings.Builder) {}
func (c subCmd) Help(out *strings.Builder) {
	fmt.Fprint(out, "Help text.")
}

type subSubCmd struct {
	Arg1 bool    `help:"help text"`
	Arg2 int     `default:"5"`
	Arg3 float64 `default:"3.5"`
	Arg4 string  `default:"abc"`
}

func (c subSubCmd) Execute(world *ecs.World, out *strings.Builder) {}
func (c subSubCmd) Help(out *strings.Builder) {
	fmt.Fprint(out, "Help text.")
}

func TestParser(t *testing.T) {
	allCommands := map[string]commandEntry{
		"help": {help{}, true},
		"cmd":  {cmd{}, true},
	}

	cmdString := "cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, isHelp, err := parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, isHelp)
	assert.Equal(t, `repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:"test"}`, fmt.Sprintf("%#v", out))

	cmdString = "cmd sub subsub"
	out, isHelp, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, isHelp)
	assert.Equal(t, `repl.subSubCmd{Arg1:false, Arg2:5, Arg3:3.5, Arg4:"abc"}`, fmt.Sprintf("%#v", out))

	cmdString = "help cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, isHelp, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.True(t, isHelp)
	assert.Equal(t, `repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:"test"}`, fmt.Sprintf("%#v", out))

	cmdString = "help"
	out, isHelp, err = parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, isHelp)

	_, ok := out.(help)
	assert.True(t, ok)
}

func TestParserListEntities(t *testing.T) {
	cmdString := "query comps=Position with=Velocity"
	out, help, err := parseInput(cmdString, defaultCommands(nil))

	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.False(t, help)
	assert.Equal(t, `repl.query{N:25, Page:0, Comps:[]string{"Position"}, With:[]string{"Velocity"}, Without:[]string(nil), Exclusive:false, Full:false}`, fmt.Sprintf("%#v", out))
}

func TestExtractHelp(t *testing.T) {
	out := strings.Builder{}

	err := extractHelp(cmd{}, &out)
	assert.Nil(t, err)
	assert.Equal(t, `Help text.
Commands:
  sub          Help text.
`, out.String())

	out = strings.Builder{}
	err = extractHelp(subSubCmd{}, &out)
	assert.Nil(t, err)
	assert.Equal(t, `Help text.
Options:
  arg1          bool     help text 
  arg2          int      Default: 5
  arg3          float    Default: 3.5
  arg4          string   Default: abc
`, out.String())
}
