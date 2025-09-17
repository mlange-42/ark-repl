package main

import (
	"strings"
	"time"

	"github.com/mlange-42/ark-repl/repl"
	"github.com/mlange-42/ark/ecs"
)

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
