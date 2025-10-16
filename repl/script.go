package repl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

type runScript struct {
	Script string
}

func (c runScript) Execute(repl *Repl, out *strings.Builder) {
	template := `package main

import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/mlange-42/ark/ecs"
)

//export RunScript
func RunScript(worldPtr uintptr, outPtr uintptr) {
	world := (*ecs.World)(unsafe.Pointer(worldPtr))
	out := (*strings.Builder)(unsafe.Pointer(outPtr))

	$$CODE$$

	_ = world
	_ = out
}

func main() {} // Required for c-shared build
`

	script := strings.Replace(template, "$$CODE$$", c.Script, 1)

	tmpScript, err := os.CreateTemp("", "*.go")
	if err != nil {
		fmt.Fprintf(out, "Error creating temporary script file: %s\n", err.Error())
		return
	}
	if _, err := tmpScript.WriteString(script); err != nil {
		fmt.Fprintf(out, "Error writing temporary script file: %s\n", err.Error())
		return
	}
	tmpScript.Close()

	tmpDll, err := os.CreateTemp("", "*.so")
	if err != nil {
		fmt.Fprintf(out, "Error creating temporary library file: %s\n", err.Error())
		return
	}
	tmpDll.Close()

	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", tmpDll.Name(), tmpScript.Name())
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "building script failed: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	dll := syscall.NewLazyDLL(tmpDll.Name())
	run := dll.NewProc("RunScript")
	_, _, err = run.Call(
		uintptr(unsafe.Pointer(repl.World())),
		uintptr(unsafe.Pointer(out)),
	)
	if err != nil && err.Error() != "The operation completed successfully." {
		panic(err)
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
