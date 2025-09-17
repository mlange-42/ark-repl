package arkinspect

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

type command interface {
	exec(repl *Repl, args []string, out *strings.Builder) error
}

type pause struct{}

func (c *pause) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		if len(args) > 0 {
			fmt.Fprint(out, "Command has no subcommands and no arguments\n")
			return
		}
		if repl.callbacks.Pause == nil {
			fmt.Fprint(out, "No pause callback provided\n")
			return
		}
		repl.callbacks.Pause(out)
	}, out)
	return nil
}

type resume struct{}

func (c *resume) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		if len(args) > 0 {
			fmt.Fprint(out, "Command has no subcommands and no arguments\n")
			return
		}
		if repl.callbacks.Resume == nil {
			fmt.Fprint(out, "No resume callback provided\n")
			return
		}
		repl.callbacks.Resume(out)
	}, out)
	return nil
}

type stop struct{}

func (c *stop) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		if len(args) > 0 {
			fmt.Fprint(out, "Command has no subcommands and no arguments\n")
			return
		}
		if repl.callbacks.Stop == nil {
			fmt.Fprint(out, "No stop callback provided\n")
			return
		}
		repl.callbacks.Stop(out)
	}, out)
	return nil
}

type help struct{}

func (c *help) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		if len(args) > 0 {
			fmt.Fprint(out, "Command has no subcommands and no arguments\n")
			return
		}
		cmds := make([]string, 0, len(commands))
		for cmd := range commands {
			cmds = append(cmds, cmd)
		}
		slices.Sort(cmds)
		fmt.Fprintf(out, "Commands: %s\n", strings.Join(cmds, ", "))
		fmt.Fprint(out, "For help on a command, use: <command> help\n")
	}, out)
	return nil
}

var listCommands = map[string]command{
	"help":       &listHelp{},
	"entities":   &listEntities{},
	"resources":  &listResources{},
	"components": &listComponents{},
}

type list struct{}

func (c *list) exec(repl *Repl, args []string, out *strings.Builder) error {
	subCmd, subArgs, ok := parseSlice(args)
	if !ok {
		(&listHelp{}).exec(repl, subArgs, out)
		return nil
	}
	if command, ok := listCommands[subCmd]; ok {
		command.exec(repl, subArgs, out)
	} else {
		fmt.Fprintf(out, "Unknown subcommand: %s\n", subCmd)
	}
	return nil
}

type listHelp struct{}

func (c *listHelp) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		cmds := make([]string, 0, len(listCommands))
		for cmd := range listCommands {
			cmds = append(cmds, cmd)
		}
		slices.Sort(cmds)
		fmt.Fprintf(out, "list subcommands: %s\n", strings.Join(cmds, ", "))
	}, out)
	return nil
}

type listEntities struct{}

func (c *listEntities) exec(repl *Repl, args []string, out *strings.Builder) error {
	limit := 25
	if len(args) > 0 {
		var err error
		limit, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(out, "Invalid argument '%s'\n", args[0])
			return nil
		}
	}
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		filter := ecs.NewUnsafeFilter(world)
		query := filter.Query()
		cnt := 0
		total := query.Count()
		for query.Next() {
			fmt.Fprintln(out, query.Entity())
			cnt++
			if cnt >= limit {
				query.Close()
				break
			}
		}
		if cnt == 0 {
			fmt.Fprint(out, "No entities\n")
		} else if total > cnt {
			fmt.Fprintf(out, "Skipping %d of %d entities\n", total-cnt, total)
		}
	}, out)
	return nil
}

type listResources struct{}

func (c *listResources) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		allRes := ecs.ResourceIDs(world)
		cnt := 0
		for _, id := range allRes {
			res := world.Resources().Get(id)
			fmt.Fprintf(out, "%d: %#v\n", id.Index(), res)
			cnt++
		}
		if cnt == 0 {
			fmt.Fprint(out, "No resources\n")
		}
	}, out)
	return nil
}

type listComponents struct{}

func (c *listComponents) exec(repl *Repl, args []string, out *strings.Builder) error {
	repl.execCommand(func(world *ecs.World, out *strings.Builder) {
		allComp := ecs.ComponentIDs(world)
		cnt := 0
		for _, id := range allComp {
			if info, ok := ecs.ComponentInfo(world, id); ok {
				fmt.Fprintf(out, "%d: %s\n", id.Index(), info.Type.Name())
				cnt++
			}
		}
		if cnt == 0 {
			fmt.Fprint(out, "No components\n")
		}
	}, out)
	return nil
}
