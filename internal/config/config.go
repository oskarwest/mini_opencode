package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	APIBaseURL            string `yaml:"api_base_url"`
	DefaultModel          string `yaml:"default_model"`
	Temperature           float64 `yaml:"temperature"`
	MaxToolIterations     int    `yaml:"max_tool_iterations"`
	CommandTimeoutSeconds int    `yaml:"command_timeout_seconds"`
	WorkspaceDir          string `yaml:"workspace_dir"`
	LogFile               string `yaml:"log_file"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		APIBaseURL:            "https://lmstudiomacmini.gse.com.co:2443/v1",
		DefaultModel:          "qwen/qwen3-coder-30b",
		Temperature:           0.7,
		MaxToolIterations:     10,
		CommandTimeoutSeconds: 30,
		WorkspaceDir:          "./workspace",
		LogFile:               "./logs/chat.log",
	}
}

// Load reads configuration from a YAML file
// If the file doesn't exist, it creates it with default values
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, create with defaults
		cfg := DefaultConfig()
		if err := cfg.Save(path); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
