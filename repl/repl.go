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
	channel   chan func()
	init      chan struct{}
	world     *ecs.World
	callbacks Callbacks
	commands  map[string]commandEntry
	system    System
	started   bool
}

func defaultCommands(r *Repl) map[string]commandEntry {
	return map[string]commandEntry{
		"help":   {help{r}, true},
		"pause":  {pause{r}, true},
		"resume": {resume{r}, true},
		"stop":   {stop{r}, true},
		"exit":   {exit{}, true},

		"stats":   {stats{}, true},
		"list":    {list{}, true},
		"query":   {query{}, true},
		"shrink":  {shrink{}, true},
		"monitor": {runTui{}, true},

		"stats-json": {getStats{r}, false},
	}
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World, callbacks Callbacks) *Repl {
	repl := Repl{
		channel:   make(chan func()),
		init:      make(chan struct{}),
		world:     world,
		callbacks: callbacks,
	}

	commands := map[string]commandEntry{}
	for k, v := range defaultCommands(&repl) {
		commands[k] = v
	}

	repl.commands = commands
	repl.system = System{repl: &repl}
	return &repl
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
//
// Note that a 'monitor' command, if given, is deferred after all other commands.
func (r *Repl) Start(commands ...string) {
	if r.started {
		fmt.Println("ERROR: REPL server is already running.")
		os.Exit(1)
	}
	r.started = true
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Ark REPL started. Type 'help' for commands.")

		if r.runInitialCommands(commands) {
			monitor.New(&localConnection{repl: r})
		}

		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			if line == "monitor" {
				monitor.New(&localConnection{repl: r})
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

func (r *Repl) runInitialCommands(commands []string) bool {
	runMonitor := false
	for _, cmd := range commands {
		fmt.Printf("> %s\n", cmd)
		if cmd == "monitor" {
			runMonitor = true
			continue
		}
		var out strings.Builder
		if !r.handleCommand(cmd, &out) {
			fmt.Print(out.String())
			break
		}
		fmt.Print(out.String())
	}
	close(r.init)
	return runMonitor
}

// StartServer starts a server for the REPL.
//
// The addr argument should be either 'host:port' or just ':port'.
func (r *Repl) StartServer(addr string) {
	if r.started {
		fmt.Println("ERROR: REPL server is already running.")
		os.Exit(1)
	}
	r.started = true
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
				cmd()
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
			cmd()
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
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	if _, err := writer.WriteString("Ark REPL connected. Type 'help' for commands.\n"); err != nil {
		panic(err)
	}
	if err := writer.Flush(); err != nil {
		panic(err)
	}

	for {
		if _, err := writer.WriteString(">\n"); err != nil {
			panic(err)
		}
		if err := writer.Flush(); err != nil {
			panic(err)
		}

		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var out strings.Builder
		if !r.handleCommand(line, &out) {
			if _, err := writer.WriteString(out.String()); err != nil {
				panic(err)
			}
			if err := writer.Flush(); err != nil {
				panic(err)
			}
			break
		}
		if _, err := writer.WriteString(out.String()); err != nil {
			panic(err)
		}
		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}
}

func (r *Repl) handleCommand(cmdString string, out *strings.Builder) bool {
	cmd, help, err := parseInput(cmdString, r.commands)
	if err != nil {
		out.WriteString(err.Error() + "\n")
		return true
	}
	if help {
		if err := extractHelp(cmd, out); err != nil {
			panic(err)
		}
		return true
	}
	cmdType := reflect.TypeOf(cmd)
	if cmdType == exitCmd {
		return false
	}
	r.execCommand(cmd, out)
	return true
}

func (r *Repl) execCommand(cmd Command, out *strings.Builder) {
	done := make(chan struct{})
	r.channel <- func() {
		cmd.Execute(r.world, out)
		close(done)
	}
	<-done
}
