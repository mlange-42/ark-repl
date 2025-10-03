package repl

import "github.com/mlange-42/ark/ecs"

// System is a UI system for the usage in applications using [ark-tools].
// Get a REPL's system with [Repl.System].
//
// [ark-tools]: https://github.com/mlange-42/ark-tools/
type System struct {
	repl *Repl
}

// InitializeUI the system. Called by [ark-tools] scheduler.
//
// [ark-tools]: https://github.com/mlange-42/ark-tools/
func (r *System) InitializeUI(w *ecs.World) {}

// UpdateUI updates the system. Called by [ark-tools] scheduler.
//
// [ark-tools]: https://github.com/mlange-42/ark-tools/
func (r *System) UpdateUI(w *ecs.World) {
	r.repl.Poll()
}

// PostUpdateUI does the final part of updating, e.g. update the GL window. Called by [ark-tools] scheduler.
//
// [ark-tools]: https://github.com/mlange-42/ark-tools/
func (r *System) PostUpdateUI(w *ecs.World) {}

// FinalizeUI the system. Called by [ark-tools] scheduler.
//
// [ark-tools]: https://github.com/mlange-42/ark-tools/
func (r *System) FinalizeUI(w *ecs.World) {}
