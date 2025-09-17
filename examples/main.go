package main

import (
	"fmt"

	arkrepl "github.com/mlange-42/ark-repl"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// CommandSystem executes commands
type CommandSystem struct {
	Repl *arkrepl.Repl
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
	app := app.New()
	app.TPS = 10
	app.FPS = 10

	callbacks := arkrepl.Callbacks{
		Pause: func() {
			app.Paused = true
			fmt.Println("Simulation paused")
		},
		Resume: func() {
			app.Paused = false
			fmt.Println("Simulation resumed")
		},
		Stop: func() {
			ecs.GetResource[resource.Termination](&app.World).Terminate = true
			fmt.Println("Simulation terminated")
		},
	}

	repl := arkrepl.NewRepl(&app.World, callbacks)

	app.AddUISystem(&CommandSystem{repl})

	repl.Start()
	app.Run()
}
