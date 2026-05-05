package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Job represents a single cron job definition.
type Job struct {
	Name       string `json:"name"`
	Schedule   string `json:"schedule"`
	Command    string `json:"command"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

// Config holds the full cronwarden configuration.
type Config struct {
	DBPath     string `json:"db_path"`
	Jobs       []Job  `json:"jobs"`
}

// Load reads and parses a JSON config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}

	if cfg.DBPath == "" {
		cfg.DBPath = "cronwarden.db"
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	for i, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", j.Name)
		}
		if j.Command == "" {
			return fmt.Errorf("job %q: command is required", j.Name)
		}
	}
	return nil
}
