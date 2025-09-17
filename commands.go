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
}

type pause struct{}

func (c *pause) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands and no arguments\n")
		return
	}
	if repl.callbacks.Pause == nil {
		fmt.Fprint(out, "No pause callback provided\n")
		return
	}
	repl.callbacks.Pause(out)
	fmt.Fprint(out, "Simulation paused\n")
}

type resume struct{}

func (c *resume) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands and no arguments\n")
		return
	}
	if repl.callbacks.Resume == nil {
		fmt.Fprint(out, "No resume callback provided\n")
		return
	}
	repl.callbacks.Resume(out)
	fmt.Fprint(out, "Simulation resumed\n")
}

type stop struct{}

func (c *stop) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands and no arguments\n")
		return
	}
	if repl.callbacks.Stop == nil {
		fmt.Fprint(out, "No stop callback provided\n")
		return
	}
	repl.callbacks.Stop(out)
	fmt.Fprint(out, "Simulation terminated\n")
}

type help struct{}

func (c *help) Execute(repl *Repl, args []string, out *strings.Builder) {
	if len(args) > 0 {
		fmt.Fprint(out, "Command has no subcommands and no arguments\n")
		return
	}
	cmds := make([]string, 0, len(repl.commands))
	for cmd := range repl.commands {
		cmds = append(cmds, cmd)
	}
	slices.Sort(cmds)
	fmt.Fprintf(out, "Commands: %s\n", strings.Join(cmds, ", "))
	fmt.Fprint(out, "For help on a command, use: <command> help\n")
}

var listCommands = map[string]Command{
	"help":       &listHelp{},
	"entities":   &listEntities{},
	"resources":  &listResources{},
	"components": &listComponents{},
}

type list struct{}

func (c *list) Execute(repl *Repl, args []string, out *strings.Builder) {
	subCmd, subArgs, ok := parseSlice(args)
	if !ok {
		(&listHelp{}).Execute(repl, subArgs, out)
		return
	}
	if command, ok := listCommands[subCmd]; ok {
		command.Execute(repl, subArgs, out)
	} else {
		fmt.Fprintf(out, "Unknown subcommand: %s\n", subCmd)
	}
}

type listHelp struct{}

func (c *listHelp) Execute(repl *Repl, args []string, out *strings.Builder) {
	cmds := make([]string, 0, len(listCommands))
	for cmd := range listCommands {
		cmds = append(cmds, cmd)
	}
	slices.Sort(cmds)
	fmt.Fprintf(out, "list subcommands: %s\n", strings.Join(cmds, ", "))
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

type listResources struct{}

func (c *listResources) Execute(repl *Repl, args []string, out *strings.Builder) {
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

type listComponents struct{}

func (c *listComponents) Execute(repl *Repl, args []string, out *strings.Builder) {
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
