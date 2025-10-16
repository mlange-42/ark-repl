package repl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/mlange-42/ark/ecs"
)

type runScript struct {
	Script string
}

func (c runScript) Execute(world *ecs.World, out *strings.Builder) {
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
	fmt.Fprint(out, "")

	$$CODE$$

	_ = world
	_ = out
}

func main() {} // Required for c-shared build
`

	script := strings.Replace(template, "$$CODE$$", c.Script, 1)

	currDir, _ := os.Getwd()
	defer os.Chdir(currDir)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Fprintf(out, "Error creating temporary script directory: %s\n", err.Error())
		return
	}
	os.Chdir(tmpDir)
	cmd := exec.Command("go", "mod", "init", "user-script")
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "building script failed: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	tmpScript, err := os.Create("script.go")
	if err != nil {
		fmt.Fprintf(out, "Error creating temporary script file: %s\n", err.Error())
		return
	}
	if _, err := tmpScript.WriteString(script); err != nil {
		fmt.Fprintf(out, "Error writing temporary script file: %s\n", err.Error())
		return
	}
	tmpScript.Close()

	cmd = exec.Command("go", "get", "./...")
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "loading dependencies failed: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	cmd = exec.Command("go", "fmt", "./script.go")
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "formatting script failed: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	cmd = exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest")
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "error installing goimports: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	cmd = exec.Command("goimports", "-w", "./script.go")
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "error installing goimports: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	cmd = exec.Command("go", "build", "-buildmode=c-shared", "-o", "script.so", "./...")
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(out, "building script failed: %s\n", err.Error())
		fmt.Println(string(cmdOut))
		return
	}

	dll := syscall.NewLazyDLL("script.so")
	run := dll.NewProc("RunScript")
	_, _, err = run.Call(
		uintptr(unsafe.Pointer(world)),
		uintptr(unsafe.Pointer(out)),
	)
	if err != nil && err.Error() != "The operation completed successfully." {
		panic(err)
	}
}

func (c runScript) Help(out *strings.Builder) {}

func parseScript(script string) (Command, bool) {
	if !(strings.HasPrefix(script, "$\n") && strings.HasSuffix(script, "\n$")) {
		return nil, false
	}

	script = strings.TrimPrefix(script, "$\n")
	script = strings.TrimSuffix(script, "\n$")

	return runScript{Script: script}, true
}
