package repl

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

// Callbacks for simulation loop control.
type Callbacks struct {
	Pause  func(out *strings.Builder)
	Resume func(out *strings.Builder)
	Stop   func(out *strings.Builder)
}

// Repl is the main entry point.
type Repl struct {
	channel   chan func(*ecs.World)
	world     *ecs.World
	callbacks Callbacks
	commands  map[string]Command
}

// NewRepl creates a new [Repl].
func NewRepl(world *ecs.World, callbacks Callbacks) *Repl {
	repl := Repl{
		channel:   make(chan func(*ecs.World)),
		world:     world,
		callbacks: callbacks,
		commands: map[string]Command{
			"help":   help{},
			"pause":  pause{},
			"resume": resume{},
			"stop":   stop{},
			"stats":  stats{},
			"list":   list{},
		},
	}
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

// Start the REPL.
func (r *Repl) Start() {
	go r.startLocal()
}

func (r *Repl) startLocal() {
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
		r.handleCommand(line, &out)
		fmt.Print(out.String())
	}
}

// StartServer starts a server for the REPL.
func (r *Repl) StartServer(addr string) {
	go r.startServer(addr)
}

func (r *Repl) startServer(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start REPL server: %w", err)
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

	return nil
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
		r.handleCommand(line, &out)
		writer.WriteString(out.String())
		writer.Flush()
	}
}

func (r *Repl) handleCommand(cmdString string, out *strings.Builder) {
	cmd, help, err := parseInput(cmdString, r.commands)
	if err != nil {
		out.WriteString(err.Error() + "\n")
		return
	}
	if help {
		cmd.Help(r, out)
		return
	}
	r.execCommand(cmd, out)
}

func (r *Repl) execCommand(cmd Command, out *strings.Builder) {
	done := make(chan struct{})
	r.channel <- func(world *ecs.World) {
		cmd.Execute(r, out)
		close(done)
	}
	<-done
}
