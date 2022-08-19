package zimple

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Block represents a single block in the statusbar
type Block struct {
	output        chan string
	sigChan       chan os.Signal
	Command       string        `yaml:"command"`
	Icon          string        `yaml:"icon"`
	Args          []string      `yaml:"args"`
	UpdateSignals []int         `yaml:"update_signals"`
	Interval      time.Duration `yaml:"interval"`
}

// Start starts executing the block and returns a channel where output
// can be listened to.
// The output channel will be closed when the context is cancelled
func (b *Block) Start(ctx context.Context) <-chan string {
	b.output = make(chan string)
	b.sigChan = make(chan os.Signal)

	go func() {
		b.runAndSend(ctx)
		for {
			select {
			case <-ctx.Done():
				close(b.output)
				return

			case sig := <-b.sigChan:
				sigNum := int(sig.(syscall.Signal))
				for _, i := range b.UpdateSignals {
					if i == sigNum {
						b.runAndSend(ctx)
						break
					}
				}

			case <-time.After(b.Interval):
				b.runAndSend(ctx)
			}
		}
	}()

	return b.output
}

// runAndSend runs the block and sends the result to the output channel
func (b *Block) runAndSend(ctx context.Context) {
	o, err := b.run(ctx)
	if err != nil {
		b.output <- "err: " + err.Error()
	} else {
		b.output <- strings.TrimSpace(o)
	}
}

// run runs the block and returns the output including the icon
func (b *Block) run(ctx context.Context) (string, error) {
	res, err := exec.CommandContext(ctx, b.Command, b.Args...).CombinedOutput()
	if b.Icon != "" {
		return fmt.Sprintf("%s%s", b.Icon, res), err
	}
	return string(res), err
}

// InformSignal informs the block of a signal
func (b *Block) InformSignal(sig os.Signal) {
	b.sigChan <- sig
}
