package repl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type runScript struct {
	Script string
}

func (c runScript) Execute(repl *Repl, out *strings.Builder) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	i.Use(map[string]map[string]reflect.Value{
		"main": {
			"world": reflect.ValueOf(repl.world),
			"out":   reflect.ValueOf(out),
		},
	})
	template := `
        package main
        import (
		    "fmt"
        )

        func main() {
            $$CODE$$
        }
    `

	script := strings.Replace(template, "$$CODE$$", c.Script, 1)

	_, err := i.Eval(script)
	if err != nil {
		fmt.Fprintf(out, "Error evaluating script: %s\n", err.Error())
	}
}

func (c runScript) Help(repl *Repl, out *strings.Builder) {}

func parseScript(script string) (Command, bool) {
	if !(strings.HasPrefix(script, "$\n") && strings.HasSuffix(script, "\n$")) {
		return nil, false
	}

	script = strings.TrimPrefix(script, "$\n")
	script = strings.TrimSuffix(script, "\n$")

	return runScript{Script: script}, true
}
