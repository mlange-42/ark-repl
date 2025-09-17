package arkrepl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

var commands = map[string]command{
	"pause":  &pause{},
	"resume": &resume{},
	"stop":   &stop{},
	"help":   &help{},
	"list":   &list{},
}

// Callbacks for simulation loop control.
type Callbacks struct {
	Pause  func()
	Resume func()
	Stop   func()
}

// Repl is the main entry point.
type Repl struct {
	channel   chan func(*ecs.World)
	world     *ecs.World
	callbacks Callbacks
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World, callbacks Callbacks) *Repl {
	repl := Repl{
		channel:   make(chan func(*ecs.World)),
		world:     world,
		callbacks: callbacks,
	}
	return &repl
}

// RunCommands runs all commands.
func (r *Repl) RunCommands() {
	for {
		select {
		case cmd := <-r.channel:
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

		r.handleCommand(line)
	}

	fmt.Println("REPL exited.")
}

func (r *Repl) handleCommand(cmd string) {
	cmdName, args := parse(cmd)

	if command, ok := commands[cmdName]; ok {
		command.exec(r, args)
	} else {
		fmt.Println("Unknown command:", cmd)
	}
}

func parse(cmd string) (string, []string) {
	tokens := strings.Split(cmd, " ")
	name, args, _ := parseSlice(tokens)
	return name, args
}

func parseSlice(tokens []string) (string, []string, bool) {
	if len(tokens) == 0 {
		return "", nil, false
	}
	cmdName := tokens[0]

	var args []string
	if len(tokens) > 1 {
		args = tokens[1:]
	}

	return cmdName, args, true
}

func (r *Repl) execCommand(fn func(*ecs.World)) {
	done := make(chan struct{})
	r.channel <- func(world *ecs.World) {
		fn(world)
		close(done)
	}
	<-done
}
