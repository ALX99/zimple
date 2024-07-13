package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alx99/zimple/internal/zimple"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	cfg, err := zimple.GetConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	run(ctx, cfg)
	cancel()
}

func run(ctx context.Context, cfg zimple.Config) {
	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}
	outputs := make([]string, len(cfg.Blocks))
	sigRedraw := make(chan struct{})
	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan)

	// initialize default text
	for i := range outputs {
		if outputs[i] == "" {
			outputs[i] = cfg.Blocks[i].Icon + "..."
		}
	}

	// Start all of the blocks
	for i := range cfg.Blocks {
		wg.Add(1)

		go func() {
			for output := range cfg.Blocks[i].Start(ctx) {
				if ctx.Err() != nil { // context has been canceled
					return
				}
				mu.Lock()
				outputs[i] = output.Stdout
				mu.Unlock()
				if output.Stderr != "" {
					fmt.Fprint(os.Stderr, output.Stderr)
				}
				sigRedraw <- struct{}{}
			}

			wg.Done()
		}()
	}

	// Goroutine that handles received signals
	go func() {
		for sig := range sigChan {
			for i := range cfg.Blocks {
				cfg.Blocks[i].NotifySignal(sig.(syscall.Signal)) // nolint:forcetypeassert
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			signal.Stop(sigChan)
			close(sigChan)
			close(sigRedraw)

			// Drain all redraw signals, we are shutting down
			for range sigRedraw { //nolint:revive
			}

			wg.Wait() // wait for all blocks to shut down

			newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			setStatusBar(newCtx, "Zimple has shut down", cfg)
			return

		case <-sigRedraw:
			mu.RLock()
			setStatusBar(ctx, strings.Join(outputs, cfg.Settings.Separator), cfg)
			mu.RUnlock()
		}
	}
}

func setStatusBar(ctx context.Context, output string, cfg zimple.Config) {
	if cfg.Settings.WriteToStdout {
		fmt.Fprintln(os.Stdout, output)
		return
	}

	err := exec.CommandContext(ctx, "xsetroot", "-name", output).Run()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
}
