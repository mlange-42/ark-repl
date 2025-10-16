package repl

import (
	"fmt"
	"math"

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

func numDigits(n int) int {
	if n == 0 {
		return 1
	}
	return int(math.Log10(float64(n))) + 1
}

func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func truncateString(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
