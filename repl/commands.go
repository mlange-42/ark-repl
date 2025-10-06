package repl

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/goccy/go-json"
	"github.com/mlange-42/ark-repl/internal/monitor"
	"github.com/mlange-42/ark/ecs"
)

// Command interface.
//
// Implement this for custom commands.
type Command interface {
	Execute(world *ecs.World, out *strings.Builder)
	Help(out *strings.Builder)
}

type commandEntry struct {
	command Command
	visible bool
}

type help struct {
	repl *Repl
}

func (c help) Execute(_ *ecs.World, out *strings.Builder) {
	cmds := make([]string, 0, len(c.repl.commands))
	help := make(map[string]string, len(c.repl.commands))
	for cmd, obj := range c.repl.commands {
		if !obj.visible {
			continue
		}
		cmds = append(cmds, cmd)
		out := strings.Builder{}
		obj.command.Help(&out)
		parts := strings.SplitN(out.String(), "\n", 2)
		var helpText string
		if len(parts) > 0 {
			helpText = parts[0]
		}
		help[cmd] = helpText
	}
	slices.Sort(cmds)

	fmt.Fprint(out, "For help on a command, use: help <command>\n\n")
	fmt.Fprintf(out, "Commands:\n")
	for _, c := range cmds {
		fmt.Fprintf(out, "  %-12s %s\n", c, help[c])
	}
}

func (c help) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Show this help.")
}

type pause struct {
	repl *Repl
}

func (c pause) Execute(_ *ecs.World, out *strings.Builder) {
	if c.repl.callbacks.Pause == nil {
		fmt.Fprint(out, "No pause callback provided\n")
		return
	}
	c.repl.callbacks.Pause(out)
	fmt.Fprint(out, "Simulation paused\n")
}

func (c pause) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Pause the connected simulation.")
}

type resume struct {
	repl *Repl
}

func (c resume) Execute(_ *ecs.World, out *strings.Builder) {
	if c.repl.callbacks.Resume == nil {
		fmt.Fprint(out, "No resume callback provided\n")
		return
	}
	c.repl.callbacks.Resume(out)
	fmt.Fprint(out, "Simulation resumed\n")
}

func (c resume) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Resume the connected simulation.")
}

type stop struct {
	repl *Repl
}

func (c stop) Execute(_ *ecs.World, out *strings.Builder) {
	if c.repl.callbacks.Stop == nil {
		fmt.Fprint(out, "No stop callback provided\n")
		return
	}
	c.repl.callbacks.Stop(out)
	fmt.Fprint(out, "Simulation terminated\n")
}

func (c stop) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Stop the connected simulation.")
}

type exit struct{}

func (c exit) Execute(_ *ecs.World, _ *strings.Builder) {}

func (c exit) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Exit the REPL without stopping the simulation.")
}

type stats struct{}

func (c stats) Execute(world *ecs.World, out *strings.Builder) {
	stats := world.Stats()
	fmt.Fprint(out, stats)
}

func (c stats) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Prints world statistics.")
}

type query struct {
	N         int      `default:"25" help:"Maximum number of entities to print."`
	Page      int      `help:"Page of entities to show (i'th N)."`
	Comps     []string `help:"Components of the query."`
	With      []string `help:"Additional components to filter for."`
	Without   []string `help:"Only entities without these components."`
	Exclusive bool     `help:"Only entities with exactly the components in 'with'."`
	Full      bool     `help:"Show all components, not only those queried."`
}

func (c query) Execute(world *ecs.World, out *strings.Builder) {
	comps, err := getComponentIDs(world, c.Comps)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}

	allIDs := ecs.ComponentIDs(world)
	compTypes := make([]reflect.Type, 0, len(allIDs))
	for _, id := range allIDs {
		info, _ := ecs.ComponentInfo(world, id)
		compTypes = append(compTypes, info.Type)
	}

	with, err := getComponentIDs(world, c.With)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}
	allComps := make([]ecs.ID, 0, len(comps)+len(with))
	allComps = append(allComps, comps...)
	allComps = append(allComps, with...)

	without, err := getComponentIDs(world, c.Without)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}

	filter := ecs.NewUnsafeFilter(world, allComps...).Without(without...)
	if c.Exclusive {
		filter = filter.Exclusive()
	}
	query := filter.Query()
	cnt := 0
	shown := 0
	total := query.Count()

	if c.N > 0 {
		compStrings := make([]string, 0, len(comps))
		show := []ecs.ID{}

		start := c.Page * c.N
		end := (c.Page + 1) * c.N
		for query.Next() {
			if cnt < start {
				cnt++
				continue
			}

			if c.Full {
				ids := query.IDs()
				for i := range ids.Len() {
					show = append(show, ids.Get(i))
				}
			} else {
				show = append(show, comps...)
			}
			for _, id := range show {
				ptr := query.Get(id)
				val := reflect.NewAt(compTypes[id.Index()], ptr).Elem()
				compStrings = append(compStrings, fmt.Sprintf("%s%+v", compTypes[id.Index()].Name(), val.Interface()))
			}

			fmt.Fprintf(out, "%v: ", query.Entity())
			fmt.Fprintln(out, strings.Join(compStrings, " "))
			cnt++
			shown++
			if cnt >= end {
				query.Close()
				break
			}
			compStrings = compStrings[:0]
			show = show[:0]
		}
	}
	fmt.Fprintf(out, "Listed %d of %d entities (page %d of %d)\n", shown, total, c.Page, (total+c.N-1)/c.N)
}

func (c query) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Query entities.")
}

type shrink struct {
}

func (c shrink) Execute(world *ecs.World, out *strings.Builder) {
	oldMem := world.Stats().Memory
	world.Shrink()
	newMem := world.Stats().Memory

	if newMem != oldMem {
		fmt.Fprintf(out, "Shrink world memory: %s -> %s\n", formatMemory(oldMem), formatMemory(newMem))
	} else {
		fmt.Fprintf(out, "Shrink had no effect: %s\n", formatMemory(newMem))
	}
}

func (c shrink) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Shrink world memory.")
}

type list struct {
	Resources  listResources
	Components listComponents
	Archetypes listArchetypes
}

func (c list) Execute(_ *ecs.World, out *strings.Builder) {
	fmt.Fprintln(out, "Lists various things. Run `help list` for details.")
}

func (c list) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Lists various things.")
}

type listResources struct{}

func (c listResources) Execute(world *ecs.World, out *strings.Builder) {
	allRes := ecs.ResourceIDs(world)
	padIDs := numDigits(len(allRes))
	cnt := 0
	for _, id := range allRes {
		res := world.Resources().Get(id)
		fmt.Fprintf(out, "%*d: %#v\n", padIDs, id.Index(), res)
		cnt++
	}
	if cnt == 0 {
		fmt.Fprint(out, "No resources\n")
	}
}

func (c listResources) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Lists resources.")
}

type listComponents struct{}

func (c listComponents) Execute(world *ecs.World, out *strings.Builder) {
	allComp := ecs.ComponentIDs(world)
	padIDs := numDigits(len(allComp))
	cnt := 0
	for _, id := range allComp {
		if info, ok := ecs.ComponentInfo(world, id); ok {
			fmt.Fprintf(out, "%*d: %s\n", padIDs, id.Index(), info.Type.String())
			cnt++
		}
	}
	if cnt == 0 {
		fmt.Fprint(out, "No components\n")
	}
}

func (c listComponents) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Lists component types.")
}

type listArchetypes struct{}

func (c listArchetypes) Execute(world *ecs.World, out *strings.Builder) {
	stats := world.Stats()

	maxEntities := 0
	maxTable := 0
	for i := range stats.Archetypes {
		arch := &stats.Archetypes[i]
		if arch.Size > maxEntities {
			maxEntities = arch.Size
		}
		if len(arch.Tables) > maxTable {
			maxTable = len(arch.Tables)
		}
	}
	padIDs := numDigits(len(stats.Archetypes))
	padEntities := numDigits(maxEntities)
	padTables := numDigits(maxTable)
	for i := range stats.Archetypes {
		arch := &stats.Archetypes[i]
		fmt.Fprintf(out, "%*d: %*d entities, %*d table(s)  %s\n", padIDs, i, padEntities, arch.Size, padTables, len(arch.Tables), strings.Join(arch.ComponentTypeNames, " "))
	}
}

func (c listArchetypes) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Lists archetypes.")
}

type runTui struct{}

func (c runTui) Execute(_ *ecs.World, out *strings.Builder) {
	fmt.Fprintln(out, "MONITOR")
}

func (c runTui) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Starts the monitoring TUI app.")
}

type getStats struct {
	repl *Repl
}

func (c getStats) Execute(world *ecs.World, out *strings.Builder) {
	stats := world.Stats()

	ticks := 0
	if c.repl.callbacks.Ticks != nil {
		ticks = c.repl.callbacks.Ticks()
	}

	s := monitor.Stats{
		Stats: stats,
		Ticks: ticks,
	}

	enc, err := json.Marshal(&s)
	if err != nil {
		panic(err)
	}
	out.Write(enc)
	out.WriteRune('\n')
}

func (c getStats) Help(out *strings.Builder) {
	fmt.Fprintln(out, "Prints world statistics in JSON format.")
}
