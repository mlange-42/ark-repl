package repl

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func TestScript(t *testing.T) {
	out := strings.Builder{}

	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	i.Use(map[string]map[string]reflect.Value{
		"script/script": {
			"Out": reflect.ValueOf(&out),
		},
	})
	i.ImportUsed()
	script := `
        //package main
        import (
		    "fmt"
			. "script"
        )

		type test[T any] struct {
		    V T
		}

        func main() {
			t := test[string]{V: "abc"}
            fmt.Fprintln(Out, "TEST")
        }
    `
	_, err := i.Eval(script)
	if err != nil {
		panic(err)
	}
	assert.Nil(t, err)

	fmt.Println(out.String())
}
