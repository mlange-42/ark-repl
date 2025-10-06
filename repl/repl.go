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
	init      chan struct{}
	world     *ecs.World
	callbacks Callbacks
	commands  map[string]commandEntry
	system    System
}

var defaultCommands = map[string]commandEntry{
	"help":    {help{}, true},
	"pause":   {pause{}, true},
	"resume":  {resume{}, true},
	"stop":    {stop{}, true},
	"exit":    {exit{}, true},
	"stats":   {stats{}, true},
	"list":    {list{}, true},
	"query":   {query{}, true},
	"shrink":  {shrink{}, true},
	"monitor": {runTui{}, true},

	"stats-json": {getStats{}, false},
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World, callbacks Callbacks) *Repl {
	commands := map[string]commandEntry{}
	for k, v := range defaultCommands {
		commands[k] = v
	}
	repl := Repl{
		channel:   make(chan func(*ecs.World)),
		init:      make(chan struct{}),
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
	r.commands[name] = commandEntry{cmd, true}
	return nil
}

// Start the REPL.
//
// Commands to execute at the first [Repl.Poll] call can be given as arguments (e.g. "pause", "monitor", ...).
func (r *Repl) Start(commands ...string) {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Ark REPL started. Type 'help' for commands.")

		// Initial commands
		runMonitor := false
		for _, cmd := range commands {
			fmt.Printf("> %s\n", cmd)
			var out strings.Builder
			if !r.handleCommand(cmd, &out) {
				fmt.Print(out.String())
				break
			}
			s := out.String()
			if s == "MONITOR\n" {
				runMonitor = true
				continue
			}
			fmt.Print(s)
		}
		close(r.init)

		if runMonitor {
			monitor.New(&localConnection{repl: r})
		}

		// Normal commands
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
			s := out.String()
			if s == "MONITOR\n" {
				monitor.New(&localConnection{repl: r})
			}
			fmt.Print(s)
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
	close(r.init)

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
	// Block for initial commands
	if !isClosed(r.init) {
		for {
			select {
			case cmd := <-r.channel:
				cmd(r.world)
			case <-r.init:
				// init closed, switch to single-command mode
				return
			}
		}
	}

	// Non-blocking for normal commands
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
		//_ = monitor.New(&localConnection{repl: r})
		//r.execCommand(cmd, out)
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
