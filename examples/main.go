package main

import (
	"time"

	arkrepl "github.com/mlange-42/ark-repl"
	"github.com/mlange-42/ark/ecs"
)

func main() {
	world := ecs.NewWorld()
	repl := arkrepl.NewRepl(&world)
	repl.Start()

	paused := false
	running := true

	for running {
		select {
		case event := <-repl.Events:
			switch event {
			case arkrepl.Pause:
				paused = true
			case arkrepl.Resume:
				paused = false
			case arkrepl.Stop:
				running = false
			}
		default:
			repl.RunCommands()
		}

		if paused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		time.Sleep(100 * time.Millisecond)
	}
}
