package arkinspect

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

// Command interface.
type Command interface {
	Execute(repl *Repl, args []string, out *strings.Builder)
	Help(repl *Repl, out *strings.Builder)
	SubCommands(repl *Repl) map[string]Command
}

type help struct{}

func (c *help) Execute(repl *Repl, args []string, out *strings.Builder) {
	name := "help"
	var command Command = c
	for _, arg := range args {
		subCommands := command.SubCommands(repl)
		if len(subCommands) == 0 {
			fmt.Fprintf(out, "Command '%s' has no subcommands\n", name)
			return
		}
		if cmd, ok := subCommands[arg]; ok {
			name = arg
			command = cmd
		} else {
			fmt.Fprintf(out, "Command '%s' has no subcommand '%s'\n", name, arg)
			return
		}
	}
	command.Help(repl, out)
}

func (c *help) Help(repl *Repl, out *strings.Builder) {
	cmds := make([]string, 0, len(repl.commands))
	for cmd := range repl.commands {
		cmds = append(cmds, cmd)
	}
	slices.Sort(cmds)
	fmt.Fprintf(out, "Commands: %s\n", strings.Join(cmds, ", "))
	fmt.Fprint(out, "For help on a command, use: help <command>\n")
}

func (c *help) SubCommands(repl *Repl) map[string]Command {
	return repl.commands
}

type pause struct{}

func (c *pause) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands or arguments\n")
		return
	}
	if repl.callbacks.Pause == nil {
		fmt.Fprint(out, "No pause callback provided\n")
		return
	}
	repl.callbacks.Pause(out)
	fmt.Fprint(out, "Simulation paused\n")
}

func (c *pause) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Pause the connected simulation")
}

func (c *pause) SubCommands(repl *Repl) map[string]Command {
	return nil
}

type resume struct{}

func (c *resume) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands or arguments\n")
		return
	}
	if repl.callbacks.Resume == nil {
		fmt.Fprint(out, "No resume callback provided\n")
		return
	}
	repl.callbacks.Resume(out)
	fmt.Fprint(out, "Simulation resumed\n")
}

func (c *resume) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Resume the connected simulation")
}

func (c *resume) SubCommands(repl *Repl) map[string]Command {
	return nil
}

type stop struct{}

func (c *stop) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands or arguments\n")
		return
	}
	if repl.callbacks.Stop == nil {
		fmt.Fprint(out, "No stop callback provided\n")
		return
	}
	repl.callbacks.Stop(out)
	fmt.Fprint(out, "Simulation terminated\n")
}

func (c *stop) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Stop the connected simulation")
}

func (c *stop) SubCommands(repl *Repl) map[string]Command {
	return nil
}

type stats struct{}

func (c *stats) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands or arguments\n")
		return
	}
	stats := repl.World().Stats()
	fmt.Fprint(out, stats)
}

func (c *stats) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Prints world statistics")
}

func (c *stats) SubCommands(repl *Repl) map[string]Command {
	return nil
}

var listCommands = map[string]Command{
	"entities":   &listEntities{},
	"resources":  &listResources{},
	"components": &listComponents{},
}

type list struct{}

func (c *list) Execute(repl *Repl, args []string, out *strings.Builder) {
	subCmd, subArgs, ok := parseSlice(args)
	if !ok {
		c.Help(repl, out)
		return
	}
	if command, ok := listCommands[subCmd]; ok {
		command.Execute(repl, subArgs, out)
	} else {
		fmt.Fprintf(out, "Unknown subcommand '%s'\n", subCmd)
	}
}

func (c *list) Help(repl *Repl, out *strings.Builder) {
	cmds := make([]string, 0, len(listCommands))
	for cmd := range listCommands {
		cmds = append(cmds, cmd)
	}
	slices.Sort(cmds)
	fmt.Fprintln(out, "Lists various things.")
	fmt.Fprintf(out, "Subcommands: %s\n", strings.Join(cmds, ", "))
}

func (c *list) SubCommands(repl *Repl) map[string]Command {
	return listCommands
}

type listEntities struct{}

func (c *listEntities) Execute(repl *Repl, args []string, out *strings.Builder) {
	limit := 25
	if len(args) > 0 {
		var err error
		limit, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(out, "Invalid argument '%s'\n", args[0])
			return
		}
	}
	filter := ecs.NewUnsafeFilter(repl.World())
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
}

func (c *listEntities) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists entities. Optional argument to limit the number of entities to list. Default 25")
}

func (c *listEntities) SubCommands(repl *Repl) map[string]Command {
	return nil
}

type listResources struct{}

func (c *listResources) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands or arguments\n")
		return
	}
	allRes := ecs.ResourceIDs(repl.World())
	cnt := 0
	for _, id := range allRes {
		res := repl.World().Resources().Get(id)
		fmt.Fprintf(out, "%d: %#v\n", id.Index(), res)
		cnt++
	}
	if cnt == 0 {
		fmt.Fprint(out, "No resources\n")
	}
}

func (c *listResources) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists resources.")
}

func (c *listResources) SubCommands(repl *Repl) map[string]Command {
	return nil
}

type listComponents struct{}

func (c *listComponents) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands or arguments\n")
		return
	}
	allComp := ecs.ComponentIDs(repl.World())
	cnt := 0
	for _, id := range allComp {
		if info, ok := ecs.ComponentInfo(repl.World(), id); ok {
			fmt.Fprintf(out, "%d: %s\n", id.Index(), info.Type.Name())
			cnt++
		}
	}
	if cnt == 0 {
		fmt.Fprint(out, "No components\n")
	}
}

func (c *listComponents) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists component types.")
}

func (c *listComponents) SubCommands(repl *Repl) map[string]Command {
	return nil
}
