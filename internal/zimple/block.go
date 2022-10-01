package zimple

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Block represents a single block in the statusbar
type Block struct {
	output        chan string
	rerun         chan interface{}
	ticker        *time.Ticker
	Command       string        `yaml:"command"`
	Icon          string        `yaml:"icon"`
	Enabled       string        `yaml:"enabled"`
	Args          []string      `yaml:"args"`
	UpdateSignals []int         `yaml:"update_signals"`
	Interval      time.Duration `yaml:"interval"`
}

// Start starts executing the block and returns a channel where output
// can be listened to.
// The output channel will be closed when the context is canceled
func (b *Block) Start(ctx context.Context) <-chan string {
	if b.Interval == 0 {
		b.Interval = 30 * 24 * time.Hour
	}

	b.ticker = time.NewTicker(b.Interval)
	b.output = make(chan string)
	b.rerun = make(chan interface{}, 100)

	go func() {
		b.runAndSend(ctx)

		for {
			select {
			case <-ctx.Done():
				b.ticker.Stop() // Stop the ticker
				close(b.output) // Close the output channel
				close(b.rerun)  // Close the rerun channel
				return

			case <-b.rerun:
				// Reset the ticker due to an out-of-flow rerun
				for len(b.ticker.C) > 0 {
					<-b.ticker.C
				}
				b.ticker.Reset(b.Interval)

				b.runAndSend(ctx)

			case <-b.ticker.C:
				b.runAndSend(ctx)
			}
		}
	}()

	return b.output
}

// runAndSend runs the block and sends the result to the output channel
func (b *Block) runAndSend(ctx context.Context) {
	o, err := b.runCmd(ctx)
	if err != nil {
		b.output <- "err: " + err.Error()
	} else {
		b.output <- strings.TrimSpace(o)
	}
}

// runCmd runs the block's command and returns the output including the icon
func (b *Block) runCmd(ctx context.Context) (string, error) {
	res, err := exec.CommandContext(ctx, b.Command, b.Args...).CombinedOutput()
	if b.Icon != "" {
		return fmt.Sprintf("%s%s", b.Icon, res), err
	}

	return string(res), err
}

// Rerun will re-run this block's command asynchronously
func (b *Block) Rerun() {
	b.rerun <- 0
}
