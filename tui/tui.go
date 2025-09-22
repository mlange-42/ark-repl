package tui

import (
	"context"
	"log"
	"math/rand"
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
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	spGreen *sparkline.SparkLine
	spRed   *sparkline.SparkLine
}

// newWidgets creates all widgets used by this demo.
func newWidgets(ctx context.Context, t terminalapi.Terminal, c *container.Container) (*widgets, error) {
	spGreen, spRed, err := newSparkLines(ctx)
	if err != nil {
		return nil, err
	}
	return &widgets{
		spGreen: spGreen,
		spRed:   spRed,
	}, nil
}

// gridLayout prepares container options that represent the desired screen layout.
// This function demonstrates the use of the grid builder.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
func gridLayout(w *widgets) ([]container.Option, error) {
	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(50,
			grid.Widget(w.spGreen,
				container.Border(linestyle.Light),
				container.BorderTitle("Green SparkLine"),
			),
		),
		grid.RowHeightPerc(50,
			grid.Widget(w.spRed,
				container.Border(linestyle.Light),
				container.BorderTitle("Red SparkLine"),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}

// rootID is the ID assigned to the root container.
const rootID = "root"

// Terminal implementations
const (
	termboxTerminal = "termbox"
	tcellTerminal   = "tcell"
)

// Run the dashboard
func Run() {
	terminal := "tcell"

	var t terminalapi.Terminal
	var err error
	switch terminal {
	case termboxTerminal:
		t, err = termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	case tcellTerminal:
		t, err = tcell.New(tcell.ColorMode(terminalapi.ColorMode256))
	default:
		log.Fatalf("Unknown terminal implementation '%s' specified. Please choose between 'termbox' and 'tcell'.", terminal)
		return
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

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
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

// newSparkLines creates two new sparklines displaying random values.
func newSparkLines(ctx context.Context) (*sparkline.SparkLine, *sparkline.SparkLine, error) {
	spGreen, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, nil, err
	}

	const max = 100
	go periodic(ctx, 250*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spGreen.Add([]int{v})
	})

	spRed, err := sparkline.New(
		sparkline.Color(cell.ColorRed),
	)
	if err != nil {
		return nil, nil, err
	}
	go periodic(ctx, 500*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spRed.Add([]int{v})
	})
	return spGreen, spRed, nil

}
