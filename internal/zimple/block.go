package zimple

import (
	"bytes"
	"context"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"
)

// Block represents a single block in the statusbar
type Block struct {
	output        chan BlockOutput
	rerun         chan struct{}
	ticker        *time.Ticker
	Command       string        `yaml:"command"`
	Icon          string        `yaml:"icon"`
	Enabled       string        `yaml:"enabled"`
	Args          []string      `yaml:"args"`
	UpdateSignals []int         `yaml:"update_signals"`
	Interval      time.Duration `yaml:"interval"`
}

type BlockOutput struct {
	Stdout string
	Stderr string
}

// Start starts executing the block and returns a channel where output
// can be listened to.
// The output channel will be closed when the context is canceled
func (b *Block) Start(ctx context.Context) <-chan BlockOutput {
	if b.Interval == 0 {
		b.Interval = 30 * 24 * time.Hour
	}

	b.ticker = time.NewTicker(b.Interval)
	b.output = make(chan BlockOutput)
	b.rerun = make(chan struct{}, 100)

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
	stdout, stderr, err := b.runCmd(ctx)
	if err != nil {
		b.output <- BlockOutput{Stdout: "err: " + err.Error()}
	} else {
		b.output <- BlockOutput{
			Stdout: strings.TrimSpace(stdout),
			Stderr: strings.TrimSpace(stderr),
		}
	}
}

// runCmd runs the block's command and returns the stdout, stderr and a possible error
func (b *Block) runCmd(ctx context.Context) (string, string, error) {
	cmd := exec.CommandContext(ctx, b.Command, b.Args...)
	stdoutBuf := bytes.NewBufferString(b.Icon)
	stderrBuf := bytes.NewBufferString("")

	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	if err := cmd.Run(); err != nil {
		return "", "", err
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// NotifySignal notifies the block that a signal has been received
func (b *Block) NotifySignal(s syscall.Signal) {
	if slices.Contains(b.UpdateSignals, int(s)) {
		b.rerun <- struct{}{}
	}
}
