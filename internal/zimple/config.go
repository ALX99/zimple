package zimple

import (
	"os"
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
	Separator string `yaml:"separator"`
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
				output:        make(chan string),
				sigChan:       make(chan os.Signal),
				Command:       "printf",
				Icon:          "",
				Args:          []string{"config file %s missing", cfgLoc},
				UpdateSignals: []int{},
				Interval:      time.Hour,
			}},
		}, nil
	}

	cfg := Config{}
	return cfg, yaml.Unmarshal(f, &cfg)
}

// getCfgFileLoc returns the configuration file location
func getCfgFileLoc() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return path.Join(path.Clean(dir), "zimple") + "/config.yaml"
	}
	return "~/.config/zimple/config.yaml"
}
