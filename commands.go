package arkrepl

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type command interface {
	exec(repl *Repl, args []string) error
}

type pause struct{}

func (c *pause) exec(repl *Repl, args []string) error {
	repl.execFunc(repl.callbacks.Pause)
	return nil
}

type resume struct{}

func (c *resume) exec(repl *Repl, args []string) error {
	repl.execFunc(repl.callbacks.Resume)
	return nil
}

type stop struct{}

func (c *stop) exec(repl *Repl, args []string) error {
	repl.execFunc(repl.callbacks.Stop)
	return nil
}

type help struct{}

func (c *help) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		cmds := []string{}
		for cmd := range commands {
			cmds = append(cmds, cmd)
		}
		slices.Sort(cmds)
		fmt.Println("Commands: ", strings.Join(cmds, ", "))
	})
	return nil
}

type ticks struct{}

func (c *ticks) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		fmt.Println("Tick: ", ecs.GetResource[resource.Tick](world).Tick)
	})
	return nil
}

var listCommands = map[string]command{
	"help":       &listHelp{},
	"entities":   &listEntities{},
	"resources":  &listResources{},
	"components": &listComponents{},
}

type list struct{}

func (c *list) exec(repl *Repl, args []string) error {
	subCmd, subArgs, ok := parseSlice(args)
	if !ok {
		(&listHelp{}).exec(repl, subArgs)
		return nil
	}

	if command, ok := listCommands[subCmd]; ok {
		command.exec(repl, subArgs)
	} else {
		fmt.Println("Unknown subcommand:", subCmd)
	}
	return nil
}

type listHelp struct{}

func (c *listHelp) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		cmds := []string{}
		for cmd := range listCommands {
			cmds = append(cmds, cmd)
		}
		slices.Sort(cmds)
		fmt.Println("list subcommands: ", strings.Join(cmds, ", "))
	})
	return nil
}

type listEntities struct{}

func (c *listEntities) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		filter := ecs.NewUnsafeFilter(world)
		query := filter.Query()
		cnt := 0
		for query.Next() {
			fmt.Println(query.Entity())
			cnt++
		}
		if cnt == 0 {
			fmt.Println("No entities")
		}
	})
	return nil
}

type listResources struct{}

func (c *listResources) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		allRes := ecs.ResourceIDs(world)
		cnt := 0
		for _, id := range allRes {
			res := world.Resources().Get(id)
			fmt.Printf("%#v\n", res)
			cnt++
		}
		if cnt == 0 {
			fmt.Println("No resources")
		}
	})
	return nil
}

type listComponents struct{}

func (c *listComponents) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		allComp := ecs.ComponentIDs(world)
		cnt := 0
		for _, id := range allComp {
			if info, ok := ecs.ComponentInfo(world, id); ok {
				fmt.Println(info.Type.Name())
				cnt++
			}
		}
		if cnt == 0 {
			fmt.Println("No components")
		}
	})
	return nil
}
