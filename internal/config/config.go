package config

import (
	"errors"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	ManageGh       bool               `yaml:"manage_gh"`
	DisableUpdates bool               `yaml:"disable_updates"`
	Default        string             `yaml:"default"`
	Profiles       map[string]Profile `yaml:"profiles"`
	Dirs           map[string]string  `yaml:"dirs"`
	SSHHosts       map[string]string  `yaml:"ssh_hosts"`
	RemoteUser     bool               `yaml:"remote_user"`
}

// Load returns an empty Config when the file does not exist, so ghx acts as plain gh.
func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	return Parse(content)
}

func Parse(content []byte) (Config, error) {
	var parsed Config
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := parsed.validate(); err != nil {
		return Config{}, err
	}
	return parsed, nil
}

func (cfg Config) validate() error {
	for name, profile := range cfg.Profiles {
		if profile.User == "" {
			return fmt.Errorf("profile %q: user is required", name)
		}
	}
	for dir, profileName := range cfg.Dirs {
		if _, ok := cfg.Profiles[profileName]; !ok {
			return fmt.Errorf("dir %q: unknown profile %q", dir, profileName)
		}
	}
	for host, profileName := range cfg.SSHHosts {
		if _, ok := cfg.Profiles[profileName]; !ok {
			return fmt.Errorf("ssh host %q: unknown profile %q", host, profileName)
		}
	}
	if _, ok := cfg.Profiles[cfg.Default]; cfg.Default != "" && !ok {
		return fmt.Errorf("default: unknown profile %q", cfg.Default)
	}
	return nil
}
