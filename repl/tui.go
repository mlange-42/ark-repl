package repl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	arkstats "github.com/mlange-42/ark/ecs/stats"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/sparkline"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	spMemory   *sparkline.SparkLine
	spReserved *sparkline.SparkLine
	spEntities *sparkline.SparkLine
}

// rootID is the ID assigned to the root container.
const rootID = "root"
const spMemoryID = "spMemory"
const spReservedID = "spReserved"
const spEntitiesID = "spEntities"

// Terminal implementations
const (
	termboxTerminal = "termbox"
	tcellTerminal   = "tcell"
)

type monitor struct {
	repl    *Repl
	widgets *widgets
	cont    *container.Container
}

func newMonitor(repl *Repl) *monitor {
	terminal := tcellTerminal

	var t terminalapi.Terminal
	var err error
	switch terminal {
	case termboxTerminal:
		t, err = termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	case tcellTerminal:
		t, err = tcell.New(tcell.ColorMode(terminalapi.ColorMode256))
	default:
		log.Fatalf("Unknown terminal implementation '%s' specified. Please choose between 'termbox' and 'tcell'.", terminal)
		return nil
	}

	if err != nil {
		panic(err)
	}
	defer t.Close()

	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	w, err := newWidgets(ctx, t, c)
	if err != nil {
		panic(err)
	}

	gridOpts, err := gridLayout(w)
	if err != nil {
		panic(err)
	}

	if err := c.Update(rootID, gridOpts...); err != nil {
		panic(err)
	}

	monitor := &monitor{
		repl:    repl,
		widgets: w,
		cont:    c,
	}

	go periodic(ctx, 500*time.Millisecond, monitor.update)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}

	return monitor
}

func (m *monitor) update() error {
	out := strings.Builder{}
	m.repl.execCommand(getStats{}, &out)

	s := arkstats.World{}
	if err := json.Unmarshal([]byte(out.String()), &s); err != nil {
		return err
	}

	m.widgets.spMemory.Add([]int{s.MemoryUsed})
	m.cont.Update(spMemoryID, container.BorderTitle(fmt.Sprintf("Memory %.2fkB", float64(s.MemoryUsed)/1024.0)))

	m.widgets.spReserved.Add([]int{s.Memory})
	m.cont.Update(spReservedID, container.BorderTitle(fmt.Sprintf("Reserved %.2fkB", float64(s.Memory)/1024.0)))

	m.widgets.spEntities.Add([]int{s.Entities.Used})
	m.cont.Update(spEntitiesID, container.BorderTitle(fmt.Sprintf("Entities %d", s.Entities.Used)))

	return nil
}

// newWidgets creates all widgets used by this demo.
func newWidgets(ctx context.Context, t terminalapi.Terminal, c *container.Container) (*widgets, error) {
	spMemory, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, err
	}
	spReserved, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, err
	}
	spEntities, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, err
	}
	return &widgets{
		spMemory:   spMemory,
		spReserved: spReserved,
		spEntities: spEntities,
	}, nil
}

// gridLayout prepares container options that represent the desired screen layout.
// This function demonstrates the use of the grid builder.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
func gridLayout(w *widgets) ([]container.Option, error) {
	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(33,
			grid.Widget(w.spMemory,
				container.ID(spMemoryID),
				container.Border(linestyle.Light),
				container.BorderTitle("Memory 0kB"),
			),
		),
		grid.RowHeightPerc(33,
			grid.Widget(w.spReserved,
				container.ID(spReservedID),
				container.Border(linestyle.Light),
				container.BorderTitle("Reserved 0kB"),
			),
		),
		grid.RowHeightPerc(33,
			grid.Widget(w.spEntities,
				container.ID(spEntitiesID),
				container.Border(linestyle.Light),
				container.BorderTitle("Entities 0"),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}

// periodic executes the provided closure periodically every interval.
// Exits when the context expires.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
