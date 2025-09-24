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

// Tick resource.
type Tick struct {
	Tick int
}

// Populate the world so there is something to see.
func Populate(world *ecs.World) {
	grid := Grid{Width: 100, Height: 60}
	ecs.AddResource(world, &grid)

	ecs.AddResource(world, &Tick{})

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

// Update the world so there are some changes.
func Update(world *ecs.World) {
	grid := ecs.GetResource[Grid](world)

	query := ecs.NewFilter0(world).Query()
	cnt := query.Count()

	cntRemove := rand.IntN(cnt / 10)
	if cntRemove > 0 {
		toRemove := make([]ecs.Entity, 0, cnt)
		for query.Next() {
			toRemove = append(toRemove, query.Entity())
		}
		rand.Shuffle(len(toRemove), func(i, j int) {
			toRemove[i], toRemove[j] = toRemove[j], toRemove[i]
		})
		for i := range cntRemove {
			world.RemoveEntity(toRemove[i])
		}
	} else {
		query.Close()
	}

	toAddPos := rand.IntN(10)
	toAddPosVel := rand.IntN(10)

	if toAddPos > 0 {
		posMap := ecs.NewMap1[Position](world)
		posMap.NewBatchFn(toAddPos, func(entity ecs.Entity, pos *Position) {
			pos.X = float64(rand.IntN(grid.Width))
			pos.Y = float64(rand.IntN(grid.Height))
		})
	}

	if toAddPosVel > 0 {
		posVelMap := ecs.NewMap2[Position, Velocity](world)
		posVelMap.NewBatchFn(toAddPosVel, func(entity ecs.Entity, pos *Position, vel *Velocity) {
			pos.X = float64(rand.IntN(grid.Width))
			pos.Y = float64(rand.IntN(grid.Height))
		})
	}
}
