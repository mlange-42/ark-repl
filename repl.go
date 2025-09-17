package arkrepl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Callbacks for simulation loop control.
type Callbacks struct {
	Pause  func()
	Resume func()
	Stop   func()
}

// Repl is the main entry point.
type Repl struct {
	commands  chan func(*ecs.World)
	world     *ecs.World
	callbacks Callbacks
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World, callbacks Callbacks) *Repl {
	repl := Repl{
		commands:  make(chan func(*ecs.World)),
		world:     world,
		callbacks: callbacks,
	}

	return &repl
}

// RunCommands runs all commands.
func (r *Repl) RunCommands() {
	for {
		select {
		case cmd := <-r.commands:
			cmd(r.world)
		default:
			return
		}
	}
}

// Start the repl.
func (r *Repl) Start() {
	go r.start()
}

func (r *Repl) start() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ark REPL started. Type 'help' for commands.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break // EOF or error
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		handleCommand(line, r)
	}

	fmt.Println("REPL exited.")
}

func handleCommand(cmd string, repl *Repl) {
	switch cmd {
	case "pause":
		repl.execFunc(repl.callbacks.Pause)
	case "resume":
		repl.execFunc(repl.callbacks.Resume)
	case "stop":
		repl.execFunc(repl.callbacks.Stop)
	case "tick":
		repl.execCommand(func(world *ecs.World) {
			fmt.Println("Tick: ", ecs.GetResource[resource.Tick](world).Tick)
		})
	case "help":
		fmt.Println("Commands: pause, resume, stop, tick, help")
	default:
		fmt.Println("Unknown command:", cmd)
	}
}

func (r *Repl) execFunc(fn func()) {
	if fn == nil {
		return
	}
	done := make(chan struct{})
	r.commands <- func(world *ecs.World) {
		fn()
		close(done)
	}
	<-done
}

func (r *Repl) execCommand(fn func(*ecs.World)) {
	done := make(chan struct{})
	r.commands <- func(world *ecs.World) {
		fn(world)
		close(done)
	}
	<-done
}
