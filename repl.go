package arkrepl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

// Event type
type Event uint8

const (
	// Pause event
	Pause Event = iota
	// Resume event
	Resume
	// Stop event
	Stop
)

// Repl is the main entry point.
type Repl struct {
	Events   chan Event
	commands chan func(*ecs.World)
	world    *ecs.World
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World) *Repl {
	repl := Repl{
		Events:   make(chan Event),
		commands: make(chan func(*ecs.World)),
		world:    world,
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
		repl.Events <- Pause
	case "resume":
		repl.Events <- Resume
	case "stop":
		repl.Events <- Stop
	case "tick":
		repl.commands <- func(world *ecs.World) {
			fmt.Println("Manual tick executed.")
			// RunSystems(world) if you want manual stepping
		}
	case "help":
		fmt.Println("Commands: pause, resume, stop, tick, help")
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
