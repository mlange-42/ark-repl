package repl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type cmd struct {
	Sub subCmd
}

type subCmd struct {
	SubSub subSubCmd
}

type subSubCmd struct {
	Arg1 bool
	Arg2 int
	Arg3 float64
	Arg4 string
}

func TestParser(t *testing.T) {
	allCommands := map[string]any{
		"cmd": cmd{},
	}

	cmdString := "cmd sub subsub arg1 arg2=1 arg3=2.0 arg4=test"
	out, err := parseInput(cmdString, allCommands)

	assert.Nil(t, err)
	assert.Equal(t, "repl.subSubCmd{Arg1:true, Arg2:1, Arg3:2, Arg4:\"test\"}", fmt.Sprintf("%#v", out))
}
