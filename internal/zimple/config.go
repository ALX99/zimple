package zimple

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// Config represents all of the configuration options
type Config struct {
	Blocks []Block `yaml:"blocks"`
}

// GetConfig reads, pareses and returns the configuration
func GetConfig() (Config, error) {
	f, err := os.ReadFile(getCfgFileLoc())
	if err != nil {
		return Config{}, err
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
