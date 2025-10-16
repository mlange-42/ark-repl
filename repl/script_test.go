package repl

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	world := ecs.NewWorld()
	out := strings.Builder{}
	repl := NewRepl(&world, Callbacks{})

	cmd := runScript{
		Script: `fmt.Fprint(out, "TEST")`,
	}

	cmd.Execute(repl, &out)

	assert.Equal(t, "TEST", out.String())
}
