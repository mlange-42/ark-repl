package examples

import (
	"math/rand/v2"

	"github.com/mlange-42/ark/ecs"
)

// Position component.
type Position struct {
	X float64
	Y float64
}

// Velocity component.
type Velocity struct {
	X float64
	Y float64
}

// Grid resource.
type Grid struct {
	Width  int
	Height int
}

// Populate the world so there is something to see.
func Populate(world *ecs.World) {
	grid := Grid{Width: 100, Height: 60}

	ecs.AddResource(world, &grid)

	posMap := ecs.NewMap1[Position](world)
	posVelMap := ecs.NewMap2[Position, Velocity](world)

	posMap.NewBatchFn(10, func(entity ecs.Entity, pos *Position) {
		pos.X = float64(rand.IntN(grid.Width))
		pos.Y = float64(rand.IntN(grid.Height))
	})

	posVelMap.NewBatchFn(10, func(entity ecs.Entity, pos *Position, vel *Velocity) {
		pos.X = float64(rand.IntN(grid.Width))
		pos.Y = float64(rand.IntN(grid.Height))
	})
}
