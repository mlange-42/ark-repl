package repl

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"
)

func getComponentIDs(world *ecs.World, compNames []string) ([]ecs.ID, error) {
	ids := []ecs.ID{}

	if len(compNames) > 0 {
		components := map[string]ecs.ID{}
		allIDs := ecs.ComponentIDs(world)

		for _, id := range allIDs {
			info, _ := ecs.ComponentInfo(world, id)
			components[info.Type.String()] = id
		}

		for _, comp := range compNames {
			if id, ok := components[comp]; ok {
				ids = append(ids, id)
			} else {
				return nil, fmt.Errorf("unknown component type '%s'; did you miss to add the package name?", comp)
			}
		}
	}
	return ids, nil
}

func formatMemory(bytes int) string {
	return fmt.Sprintf("%.1fkB", float64(bytes)/1024.0)
}
