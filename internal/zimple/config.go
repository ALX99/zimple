package zimple

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents all of the configuration options
type Config struct {
	Settings Settings `yaml:"settings"`
	Blocks   []Block  `yaml:"blocks"`
}

type Settings struct {
	Separator     string `yaml:"separator"`
	WriteToStdout bool   `yaml:"write_to_stdout"`
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	cfgLoc := getCfgFileLoc()

	f, err := os.ReadFile(cfgLoc)
	if err != nil {
		if !os.IsNotExist(err) {
			return Config{}, err
		}

		// Safe default if no config found
		return Config{
			Blocks: []Block{{
				output:        make(chan BlockOutput),
				rerun:         make(chan interface{}),
				ticker:        &time.Ticker{},
				Command:       "printf",
				Icon:          "",
				Enabled:       "",
				Args:          []string{"config file %s missing", cfgLoc},
				UpdateSignals: []int{},
				Interval:      time.Hour,
			}},
		}, nil
	}

	cfg := Config{}
	if err = yaml.Unmarshal(f, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.Blocks, err = filterDisabledBlocks(cfg.Blocks); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// filterDisabledBlocks removes all blocks that are not enabled
func filterDisabledBlocks(blocks []Block) ([]Block, error) {
	bls := make([]Block, 0, len(blocks))
	var ee *exec.ExitError

	for _, block := range blocks {
		if block.Enabled == "" {
			bls = append(bls, block)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("if %s;then exit 0;else exit 21;fi", block.Enabled)).CombinedOutput()
		if err != nil {
			if errors.As(err, &ee) && ee.ExitCode() == 21 {
				continue // Block disabled
			}

			return []Block{}, err
		}

		bls = append(bls, block)
	}

	return bls, nil
}

// getCfgFileLoc returns the configuration file location
func getCfgFileLoc() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return path.Join(path.Clean(dir), "zimple") + "/config.yaml"
	}

	return "~/.config/zimple/config.yaml"
}
