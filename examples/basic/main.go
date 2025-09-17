package main

import (
	"strings"
	"time"

	repl "github.com/mlange-42/ark-inspect"
	"github.com/mlange-42/ark/ecs"
)

// CommandSystem executes commands
type CommandSystem struct {
	Repl *repl.Repl
}

// InitializeUI the system.
func (s *CommandSystem) InitializeUI(w *ecs.World) {}

// UpdateUI updates the system.
func (s *CommandSystem) UpdateUI(w *ecs.World) {
	s.Repl.RunCommands()
}

// PostUpdateUI does the final part of updating, e.g. update the GL window.
func (s *CommandSystem) PostUpdateUI(w *ecs.World) {}

// FinalizeUI the system.
func (s *CommandSystem) FinalizeUI(w *ecs.World) {}

func main() {
	world := ecs.NewWorld()
	world.NewEntities(100, nil)

	pause := false
	stop := false

	// Callbacks for loop control.
	callbacks := repl.Callbacks{
		Pause: func(out *strings.Builder) {
			pause = true
		},
		Resume: func(out *strings.Builder) {
			pause = false
		},
		Stop: func(out *strings.Builder) {
			stop = true
		},
	}

	repl := repl.NewRepl(&world, callbacks)

	// For control from this terminal:
	repl.Start()

	// For control from another terminal:
	//repl.StartServer(":9000")

	// Update loop.
	for {
		// Execute incoming REPL commands.
		repl.RunCommands()

		if stop {
			break
		}
		if pause {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Update step
		time.Sleep(50 * time.Millisecond)
	}
}
