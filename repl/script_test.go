package repl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"github.com/stretchr/testify/assert"
)

func TestScriptPrint(t *testing.T) {
	world := ecs.NewWorld()
	out := strings.Builder{}
	repl := NewRepl(&world, Callbacks{})

	cmd := runScript{
		Script: `fmt.Fprint(out, "TEST")`,
	}

	cmd.Execute(repl.world, &out)

	assert.Equal(t, "TEST", out.String())
}

func TestScriptQuery(t *testing.T) {
	world := ecs.NewWorld()
	out := strings.Builder{}
	repl := NewRepl(&world, Callbacks{})

	cmd := runScript{
		Script: `
	_ = ecs.NewFilter1[examples.Position](world)
		`,
	}

	cmd.Execute(repl.world, &out)

	fmt.Println(out.String())
	//assert.Equal(t, "TEST", out.String())
}
