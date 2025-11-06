package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

const ConfigDir = "./"

// Config holds the complete configuration.
type Config struct {
	Ethereum EthereumConfig `toml:"ethereum"`
}

// Load loads configuration from a TOML file.
func Load(path string) (*Config, error) {
	var cfg Config

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if err := c.Ethereum.Validate(); err != nil {
		return fmt.Errorf("ethereum configuration error: %w", err)
	}

	return nil
}
