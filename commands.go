package arkrepl

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

type command interface {
	exec(repl *Repl, args []string) error
}

type pause struct{}

func (c *pause) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		if len(args) > 0 {
			fmt.Println("Command has no subcommands and no arguments")
			return
		}
		if repl.callbacks.Pause == nil {
			fmt.Println("No pause callback provided")
			return
		}
		repl.callbacks.Pause()
	})
	return nil
}

type resume struct{}

func (c *resume) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		if len(args) > 0 {
			fmt.Println("Command has no subcommands and no arguments")
			return
		}
		if repl.callbacks.Resume == nil {
			fmt.Println("No resume callback provided")
			return
		}
		repl.callbacks.Resume()
	})
	return nil
}

type stop struct{}

func (c *stop) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		if len(args) > 0 {
			fmt.Println("Command has no subcommands and no arguments")
			return
		}
		if repl.callbacks.Stop == nil {
			fmt.Println("No stop callback provided")
			return
		}
		repl.callbacks.Stop()
	})
	return nil
}

type help struct{}

func (c *help) exec(repl *Repl, args []string) error {
	repl.execCommand(func(world *ecs.World) {
		if len(args) > 0 {
			fmt.Println("Command has no subcommands and no arguments")
			return
		}
		cmds := []string{}
		for cmd := range commands {
			cmds = append(cmds, cmd)
		}
		slices.Sort(cmds)
		fmt.Println("Commands:", strings.Join(cmds, ", "))
		fmt.Println("For help on a command, use: <command> help")
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
		fmt.Println("list subcommands:", strings.Join(cmds, ", "))
	})
	return nil
}

type listEntities struct{}

func (c *listEntities) exec(repl *Repl, args []string) error {
	limit := 25
	if len(args) > 0 {
		var err error
		limit, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Invalid argument '%s'\n", args[0])
			return nil
		}
	}
	repl.execCommand(func(world *ecs.World) {
		filter := ecs.NewUnsafeFilter(world)
		query := filter.Query()
		cnt := 0
		total := query.Count()
		for query.Next() {
			fmt.Println(query.Entity())
			cnt++
			if cnt >= limit {
				query.Close()
				break
			}
		}
		if cnt == 0 {
			fmt.Println("No entities")
		} else {
			if total > cnt {
				fmt.Printf("Skipping %d of %d entities\n", total-cnt, total)
			}
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
			fmt.Printf("%d: %#v\n", id.Index(), res)
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
		ecs.ComponentID[list](world)
		allComp := ecs.ComponentIDs(world)
		cnt := 0
		for _, id := range allComp {
			if info, ok := ecs.ComponentInfo(world, id); ok {
				fmt.Printf("%d: %s\n", id.Index(), info.Type.Name())
				cnt++
			}
		}
		if cnt == 0 {
			fmt.Println("No components")
		}
	})
	return nil
}
