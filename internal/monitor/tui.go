package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

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
	"github.com/mum4k/termdash/widgets/text"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	spMemory   *sparkline.SparkLine
	spReserved *sparkline.SparkLine
	spEntities *sparkline.SparkLine
	spFPS      *sparkline.SparkLine
}

// rootID is the ID assigned to the root container.
const rootID = "root"
const spMemoryID = "spMemory"
const spReservedID = "spReserved"
const spEntitiesID = "spEntities"
const spFpsID = "spFps"

// Terminal implementations
const (
	termboxTerminal = "termbox"
	tcellTerminal   = "tcell"
)

// Monitor TUI.
type Monitor struct {
	stats     Connection
	widgets   *widgets
	cont      *container.Container
	ticks     int
	lastTicks int
	lastTime  time.Time
}

// New monitor TUI.
func New(stats Connection) *Monitor {
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
	w, err := newWidgets()
	if err != nil {
		panic(err)
	}

	gridOpts, err := layout(w)
	if err != nil {
		panic(err)
	}

	if err := c.Update(rootID, gridOpts...); err != nil {
		panic(err)
	}

	monitor := &Monitor{
		stats:   stats,
		widgets: w,
		cont:    c,
	}

	go periodic(ctx, 1000*time.Millisecond, monitor.update)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
		cmd := ""
		switch k.Key {
		case 'p':
			cmd = "pause"
		case 'r':
			cmd = "resume"
		case 's':
			cmd = "shrink"
		}
		if cmd != "" {
			if err := stats.Exec(cmd); err != nil {
				panic(err)
			}
		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}

	return monitor
}

func (m *Monitor) update() error {
	s, err := m.stats.Get()
	if err != nil {
		return err
	}

	if err := m.widgets.spMemory.Add([]int{s.Stats.MemoryUsed}); err != nil {
		return err
	}
	if err := m.cont.Update(spMemoryID, container.BorderTitle(fmt.Sprintf("Memory %.1fkB", float64(s.Stats.MemoryUsed)/1024.0))); err != nil {
		return err
	}

	if err := m.widgets.spReserved.Add([]int{s.Stats.Memory}); err != nil {
		return err
	}
	if err := m.cont.Update(spReservedID, container.BorderTitle(fmt.Sprintf("Reserved %.1fkB", float64(s.Stats.Memory)/1024.0))); err != nil {
		return err
	}

	if err := m.widgets.spEntities.Add([]int{s.Stats.Entities.Used}); err != nil {
		return err
	}
	if err := m.cont.Update(spEntitiesID, container.BorderTitle(fmt.Sprintf("Entities %d", s.Stats.Entities.Used))); err != nil {
		return err
	}

	timePassed := time.Since(m.lastTime)
	fps := float64(s.Ticks-m.lastTicks) / timePassed.Seconds()
	if err := m.widgets.spFPS.Add([]int{max(int(fps), 0)}); err != nil {
		return err
	}
	if err := m.cont.Update(spFpsID, container.BorderTitle(fmt.Sprintf("FPS %.0f", fps))); err != nil {
		return err
	}

	m.lastTime = time.Now()
	m.lastTicks = s.Ticks

	m.ticks++

	return nil
}

// newWidgets creates all widgets used by this demo.
func newWidgets() (*widgets, error) {
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
	spFPS, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, err
	}
	return &widgets{
		spMemory:   spMemory,
		spReserved: spReserved,
		spEntities: spEntities,
		spFPS:      spFPS,
	}, nil
}

func layout(w *widgets) ([]container.Option, error) {
	builder := grid.New()

	leftColumn := grid.ColWidthPerc(25,
		grid.RowHeightPerc(25,
			grid.Widget(w.spFPS,
				container.ID(spFpsID),
				container.Border(linestyle.Light),
				container.BorderTitle("FPS 0"),
			),
		),
		grid.RowHeightPerc(25,
			grid.Widget(w.spMemory,
				container.ID(spMemoryID),
				container.Border(linestyle.Light),
				container.BorderTitle("Memory 0kB"),
			),
		),
		grid.RowHeightPerc(25,
			grid.Widget(w.spReserved,
				container.ID(spReservedID),
				container.Border(linestyle.Light),
				container.BorderTitle("Reserved 0kB"),
			),
		),
		grid.RowHeightPerc(25,
			grid.Widget(w.spEntities,
				container.ID(spEntitiesID),
				container.Border(linestyle.Light),
				container.BorderTitle("Entities 0"),
			),
		),
	)

	builder.Add(
		leftColumn,
		grid.ColWidthPerc(75),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}

	help, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := help.Write("Help: [Esc]ape [P]ause [R]esume [S]hrink"); err != nil {
		panic(err)
	}

	outer := []container.Option{
		container.SplitHorizontal(
			container.Top(gridOpts...),
			container.Bottom(container.PlaceWidget(help)),
			container.SplitFixedFromEnd(1),
		),
	}

	return outer, nil
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
