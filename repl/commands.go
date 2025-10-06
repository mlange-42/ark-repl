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
type Command interface {
	Execute(repl *Repl, out *strings.Builder)
	Help(repl *Repl, out *strings.Builder)
}

type commandEntry struct {
	command Command
	visible bool
}

type help struct{}

func (c help) Execute(repl *Repl, out *strings.Builder) {
	cmds := make([]string, 0, len(repl.commands))
	help := make(map[string]string, len(repl.commands))
	for cmd, obj := range repl.commands {
		if !obj.visible {
			continue
		}
		cmds = append(cmds, cmd)
		out := strings.Builder{}
		obj.command.Help(repl, &out)
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

func (c help) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Show this help.")
}

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

type exit struct{}

func (c exit) Execute(repl *Repl, out *strings.Builder) {
	if repl.callbacks.Stop == nil {
		fmt.Fprint(out, "No stop callback provided\n")
		return
	}
	repl.callbacks.Stop(out)
	fmt.Fprint(out, "Simulation terminated\n")
}

func (c exit) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Exit the REPL without stopping the simulation.")
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
	Page      int      `help:"Page of entities to show (i'th N)."`
	Comps     []string `help:"Components of the query."`
	With      []string `help:"Additional components to filter for."`
	Without   []string `help:"Only entities without these components."`
	Exclusive bool     `help:"Only entities with exactly the components in 'with'."`
	Full      bool     `help:"Show all components, not only those queried."`
}

func (c query) Execute(repl *Repl, out *strings.Builder) {
	comps, err := getComponentIDs(repl.world, c.Comps)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}

	allIDs := ecs.ComponentIDs(repl.world)
	compTypes := make([]reflect.Type, 0, len(allIDs))
	for _, id := range allIDs {
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

func (c query) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Query entities.")
}

type shrink struct {
}

func (c shrink) Execute(repl *Repl, out *strings.Builder) {
	oldMem := repl.World().Stats().Memory
	repl.world.Shrink()
	newMem := repl.World().Stats().Memory

	if newMem != oldMem {
		fmt.Fprintf(out, "Shrinked world memory: %s -> %s\n", formatMemory(oldMem), formatMemory(newMem))
	} else {
		fmt.Fprintf(out, "Shrink had no effect: %s\n", formatMemory(newMem))
	}
}

func (c shrink) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Shrink world memory.")
}

type list struct {
	Resources  listResources
	Components listComponents
	Archetypes listArchetypes
}

func (c list) Execute(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists various things. Run `help list` for details.")
}

func (c list) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists various things.")
}

type listResources struct{}

func (c listResources) Execute(repl *Repl, out *strings.Builder) {
	allRes := ecs.ResourceIDs(repl.World())
	padIDs := numDigits(len(allRes))
	cnt := 0
	for _, id := range allRes {
		res := repl.World().Resources().Get(id)
		fmt.Fprintf(out, "%*d: %#v\n", padIDs, id.Index(), res)
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
	padIDs := numDigits(len(allComp))
	cnt := 0
	for _, id := range allComp {
		if info, ok := ecs.ComponentInfo(repl.World(), id); ok {
			fmt.Fprintf(out, "%*d: %s\n", padIDs, id.Index(), info.Type.String())
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

type listArchetypes struct{}

func (c listArchetypes) Execute(repl *Repl, out *strings.Builder) {
	stats := repl.World().Stats()

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

func (c listArchetypes) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Lists archetypes.")
}

type runTui struct{}

func (c runTui) Execute(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "MONITOR")
}

func (c runTui) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Starts the monitoring TUI app.")
}

type getStats struct{}

func (c getStats) Execute(repl *Repl, out *strings.Builder) {
	stats := repl.world.Stats()

	ticks := 0
	if repl.callbacks.Ticks != nil {
		ticks = repl.callbacks.Ticks()
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

func (c getStats) Help(repl *Repl, out *strings.Builder) {
	fmt.Fprintln(out, "Prints world statistics in JSON format.")
}
