package main

import (
	"strings"

	"github.com/mlange-42/ark-repl/examples"
	"github.com/mlange-42/ark-repl/repl"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

func main() {
	app := app.New()
	app.TPS = 10
	app.FPS = 10

	// Populate the world so there is something to see.
	examples.Populate(&app.World)

	// Callbacks for loop control.
	callbacks := repl.Callbacks{
		Pause: func(out *strings.Builder) {
			app.Paused = true
		},
		Resume: func(out *strings.Builder) {
			app.Paused = false
		},
		Stop: func(out *strings.Builder) {
			ecs.GetResource[resource.Termination](&app.World).Terminate = true
		},
	}

	repl := repl.NewRepl(&app.World, callbacks)

	app.AddUISystem(&CommandSystem{repl})

	// For control from this terminal:
	//repl.Start()

	// For control from another terminal:
	repl.StartServer(":9000")
	app.Run()
}

// CommandSystem executes incoming REPL commands.
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
