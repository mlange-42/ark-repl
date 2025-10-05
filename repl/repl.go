package repl

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/mlange-42/ark-repl/internal/monitor"
	"github.com/mlange-42/ark/ecs"
)

var runTuiCmd = reflect.TypeFor[runTui]()
var exitCmd = reflect.TypeFor[exit]()

// Callbacks for simulation loop control.
// Individual callbacks are optional, but required to enable the resp. functionality.
type Callbacks struct {
	// Pause the simulation.
	Pause func(out *strings.Builder)
	// Resume the simulation.
	Resume func(out *strings.Builder)
	// Stop the simulation.
	Stop func(out *strings.Builder)
	// Get the current simulation tick. Used to calculate frame rate.
	Ticks func() int
}

// Repl is the main entry point.
type Repl struct {
	channel   chan func(*ecs.World)
	world     *ecs.World
	callbacks Callbacks
	commands  map[string]Command
	system    System
}

var defaultCommands = map[string]Command{
	"help":       help{},
	"pause":      pause{},
	"resume":     resume{},
	"stop":       stop{},
	"exit":       exit{},
	"stats":      stats{},
	"stats-json": getStats{},
	"list":       list{},
	"query":      query{},
	"shrink":     shrink{},
	"monitor":    runTui{},
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World, callbacks Callbacks) *Repl {
	commands := map[string]Command{}
	for k, v := range defaultCommands {
		commands[k] = v
	}
	repl := Repl{
		channel:   make(chan func(*ecs.World)),
		world:     world,
		callbacks: callbacks,
		commands:  commands,
	}
	repl.system = System{repl: &repl}
	return &repl
}

// World returns the World associated to this REPL.
func (r *Repl) World() *ecs.World {
	return r.world
}

// AddCommand adds a command to the REPL.
//
// Returns an error if a command with the same name is already registered.
func (r *Repl) AddCommand(name string, cmd Command) error {
	if _, ok := r.commands[name]; ok {
		return fmt.Errorf("command '%s' is already registered", name)
	}
	r.commands[name] = cmd
	return nil
}

// Start the REPL.
func (r *Repl) Start() {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Ark REPL started. Type 'help' for commands.")

		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var out strings.Builder
			if !r.handleCommand(line, &out) {
				fmt.Print(out.String())
				break
			}
			fmt.Print(out.String())
		}
	}()
}

// StartServer starts a server for the REPL.
func (r *Repl) StartServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to start REPL server: %s", err)
		return
	}
	fmt.Println("REPL server listening on", addr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("REPL connection error:", err)
				continue
			}
			go r.handleConnection(conn)
		}
	}()
}

// Poll runs all commands.
func (r *Repl) Poll() {
	for {
		select {
		case cmd := <-r.channel:
			cmd(r.world)
		default:
			return
		}
	}
}

// System returns a UI system for the usage in applications using [ark-tools].
//
// Usage:
//
//	app := app.New()
//	callbacks := repl.Callbacks{
//	    // define callbacks...
//	}
//	repl := repl.NewRepl(&app.World, callbacks)
//
//	// Set up other systems...
//
//	app.AddUISystem(repl.System())
//
// [ark-tools]: https://github.com/mlange-42/ark-tools/
func (r *Repl) System() *System {
	return &r.system
}

func (r *Repl) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	writer.WriteString("Ark REPL connected. Type 'help' for commands.\n")
	writer.Flush()

	for {
		writer.WriteString(">\n")
		writer.Flush()

		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var out strings.Builder
		if !r.handleCommand(line, &out) {
			writer.WriteString(out.String())
			writer.Flush()
			break
		}
		writer.WriteString(out.String())
		writer.Flush()
	}
}

func (r *Repl) handleCommand(cmdString string, out *strings.Builder) bool {
	cmd, help, err := parseInput(cmdString, r.commands)
	if err != nil {
		out.WriteString(err.Error() + "\n")
		return true
	}
	if help {
		if err := extractHelp(r, cmd, out); err != nil {
			panic(err)
		}
		return true
	}
	cmdType := reflect.TypeOf(cmd)
	switch cmdType {
	case runTuiCmd:
		_ = monitor.New(&localStats{repl: r})
		r.execCommand(cmd, out)
	case exitCmd:
		return false
	}
	r.execCommand(cmd, out)
	return true
}

func (r *Repl) execCommand(cmd Command, out *strings.Builder) {
	done := make(chan struct{})
	r.channel <- func(world *ecs.World) {
		cmd.Execute(r, out)
		close(done)
	}
	<-done
}
