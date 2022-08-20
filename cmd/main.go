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

	"github.com/alx99/zimple/internal/zimple"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT)
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
	sigRedraw := make(chan any)
	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan)

	// Start all of the blocks
	for i := range cfg.Blocks {
		wg.Add(1)
		go func(b *zimple.Block, i int) {
			for output := range b.Start(ctx) {
				mu.Lock()
				outputs[i] = output
				mu.Unlock()
				sigRedraw <- 0
			}
			wg.Done()
		}(&cfg.Blocks[i], i)
	}

	go func() {
		for sig := range sigChan {
			for i := range cfg.Blocks {
				cfg.Blocks[i].InformSignal(sig)
			}
		}
	}()

	go func() {
		wg.Wait()
		signal.Stop(sigChan)
		close(sigRedraw)
		close(sigChan)
	}()

	for {
		select {
		case <-ctx.Done():
			// Drain the signals
			for range sigRedraw {
			}
			return

		case <-sigRedraw:
			mu.RLock()
			err := exec.CommandContext(ctx, "xsetroot", "-name", strings.Join(outputs, " / ")).Run()
			if err != nil {
				// Give it a second try
				err = exec.CommandContext(ctx, "xsetroot", "-name", fmt.Sprintf("error: %s", err)).Run()
				if err != nil {
					// Give up
					fmt.Fprint(os.Stderr, err.Error())
				}
			}
			mu.RUnlock()
		}
	}

}
