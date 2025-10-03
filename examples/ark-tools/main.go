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
	app := app.New(32)
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
		Ticks: func() int {
			return int(ecs.GetResource[resource.Tick](&app.World).Tick)
		},
	}

	repl := repl.NewRepl(&app.World, callbacks)

	app.AddSystem(&UpdateSystem{})
	app.AddUISystem(repl.System())

	// For control from this terminal:
	repl.Start()

	// For control from another terminal:
	//repl.StartServer(":9000")

	app.Run()
}

// UpdateSystem calls [examples.Update] every tick
type UpdateSystem struct{}

// Initialize the system.
func (s *UpdateSystem) Initialize(w *ecs.World) {}

// Update the system.
func (s *UpdateSystem) Update(w *ecs.World) {
	examples.Update(w)
}

// Finalize the system.
func (s *UpdateSystem) Finalize(w *ecs.World) {}
