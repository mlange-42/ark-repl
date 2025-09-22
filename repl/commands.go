package repl

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/mlange-42/ark/ecs"
)

// Command interface.
type Command interface {
	Execute(repl *Repl, out *strings.Builder)
	Help(repl *Repl, out *strings.Builder)
}

type help struct{}

func (c help) Execute(repl *Repl, out *strings.Builder) {
	cmds := make([]string, 0, len(repl.commands))
	for cmd := range repl.commands {
		cmds = append(cmds, cmd)
	}
	slices.Sort(cmds)

	fmt.Fprint(out, "For help on a command, use: help <command>\n\n")
	fmt.Fprintf(out, "Commands:\n")
	for _, c := range cmds {
		fmt.Fprintf(out, "  %s\n", c)
	}
}

func (c help) Help(repl *Repl, out *strings.Builder) {}

type pause struct{}

func (c pause) Execute(repl *Repl, out *strings.Builder) {
	if repl.callbacks.Pause == nil {
		fmt.Fprint(out, "No pause callback provided\n")
		return
	}
	repl.callbacks.Pause(out)
	fmt.Fprint(out, "Simulation paused\n")
}

func (c pause) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Pause the connected simulation.")
}

type resume struct{}

func (c resume) Execute(repl *Repl, out *strings.Builder) {
	if repl.callbacks.Resume == nil {
		fmt.Fprint(out, "No resume callback provided\n")
		return
	}
	repl.callbacks.Resume(out)
	fmt.Fprint(out, "Simulation resumed\n")
}

func (c resume) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Resume the connected simulation.")
}

type stop struct{}

func (c stop) Execute(repl *Repl, out *strings.Builder) {
	if repl.callbacks.Stop == nil {
		fmt.Fprint(out, "No stop callback provided\n")
		return
	}
	repl.callbacks.Stop(out)
	fmt.Fprint(out, "Simulation terminated\n")
}

func (c stop) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Stop the connected simulation.")
}

type stats struct{}

func (c stats) Execute(repl *Repl, out *strings.Builder) {
	stats := repl.World().Stats()
	fmt.Fprint(out, stats)
}

func (c stats) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Prints world statistics.")
}

type query struct {
	N         int      `default:"25" help:"Maximum number of entities to print."`
	Comps     []string `help:"Components of the query."`
	With      []string `help:"Additional components to filter for."`
	Without   []string `help:"Only entities without these components."`
	Exclusive bool     `help:"Only entities with exactly the components in 'with'."`
}

func (c query) Execute(repl *Repl, out *strings.Builder) {
	limit := c.N

	comps, err := getComponentIDs(repl.world, c.Comps)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}
	compTypes := make([]reflect.Type, 0, len(comps))
	for _, id := range comps {
		info, _ := ecs.ComponentInfo(repl.world, id)
		compTypes = append(compTypes, info.Type)
	}

	with, err := getComponentIDs(repl.world, c.With)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}
	allComps := make([]ecs.ID, 0, len(comps)+len(with))
	allComps = append(allComps, comps...)
	allComps = append(allComps, with...)

	without, err := getComponentIDs(repl.world, c.Without)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}

	filter := ecs.NewUnsafeFilter(repl.World(), allComps...).Without(without...)
	if c.Exclusive {
		filter = filter.Exclusive()
	}
	query := filter.Query()
	cnt := 0
	total := query.Count()
	if limit > 0 {
		compStrings := make([]string, len(comps))
		for query.Next() {
			fmt.Fprintf(out, "%v: ", query.Entity())
			for i, id := range comps {
				ptr := query.Get(id)
				val := reflect.NewAt(compTypes[i], ptr).Elem()
				compStrings[i] = fmt.Sprintf("%s%+v", compTypes[i].Name(), val.Interface())
			}
			fmt.Fprintln(out, strings.Join(compStrings, " "))
			cnt++
			if cnt >= limit {
				query.Close()
				break
			}
		}
	}
	if total == 0 {
		fmt.Fprint(out, "No entities\n")
	} else if total > cnt {
		fmt.Fprintf(out, "Skipping %d of %d entities\n", total-cnt, total)
	}
}

func (c query) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Query entities")
}

type list struct {
	Resources  listResources
	Components listComponents
}

func (c list) Execute(repl *Repl, out *strings.Builder) {
	c.Help(repl, out)
}

func (c list) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists various things.")
}

type listResources struct{}

func (c listResources) Execute(repl *Repl, out *strings.Builder) {
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

func (c listResources) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists resources.")
}

type listComponents struct{}

func (c listComponents) Execute(repl *Repl, out *strings.Builder) {
	allComp := ecs.ComponentIDs(repl.World())
	cnt := 0
	for _, id := range allComp {
		if info, ok := ecs.ComponentInfo(repl.World(), id); ok {
			fmt.Fprintf(out, "%d: %s\n", id.Index(), info.Type.String())
			cnt++
		}
	}
	if cnt == 0 {
		fmt.Fprint(out, "No components\n")
	}
}

func (c listComponents) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists component types.")
}
