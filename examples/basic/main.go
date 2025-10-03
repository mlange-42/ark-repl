package main

import (
	"strings"
	"time"

	"github.com/mlange-42/ark-repl/examples"
	"github.com/mlange-42/ark-repl/repl"
	"github.com/mlange-42/ark/ecs"
)

func main() {
	world := ecs.NewWorld(32)

	// Populate the world so there is something to see.
	examples.Populate(&world)

	pause := true
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
		Ticks: func() int {
			return ecs.GetResource[examples.Tick](&world).Tick
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
		repl.Poll()

		if stop { // Stopped?
			break
		}
		if pause { // Paused?
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Update step
		examples.Update(&world)
		ecs.GetResource[examples.Tick](&world).Tick++
		// Emulate frame time
		time.Sleep(50 * time.Millisecond)
	}
}
