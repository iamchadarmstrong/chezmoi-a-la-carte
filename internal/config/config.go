// Package config provides configuration management for the a-la-carte application.
//
// The configuration is loaded with the following precedence (highest to lowest):
// 1. Environment variable (A_LA_CARTE_CONFIG)
// 2. Command line flags (--config)
// 3. XDG config file ($HOME/.config/a-la-carte/a-la-carte.yml)
// 4. Built-in defaults
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// EnvConfigPath is the environment variable name for the config path
	EnvConfigPath = "A_LA_CARTE_CONFIG"

	// DefaultConfigFilename is the default config filename
	DefaultConfigFilename = "a-la-carte.yml"

	// DefaultConfigDirname is the default config directory name under XDG_CONFIG_HOME
	DefaultConfigDirname = "a-la-carte"
)

var (
	// ErrNoConfig is returned when no config file is found
	ErrNoConfig = errors.New("no configuration file found")
)

// Config represents the application configuration
type Config struct {
	// UI configuration settings
	UI struct {
		// Theme controls the color scheme (light, dark, system)
		Theme string `yaml:"theme,omitempty"`
		// DetailHeight is the height of the detail pane
		DetailHeight int `yaml:"detailHeight,omitempty"`
		// ListHeight is the height of the list pane
		ListHeight int `yaml:"listHeight,omitempty"`
		// EmojisEnabled controls whether emojis are displayed in the UI
		EmojisEnabled bool `yaml:"emojisEnabled,omitempty"`
	} `yaml:"ui,omitempty"`

	// Software configuration
	Software struct {
		// ManifestPath is the path to the software manifest
		ManifestPath string `yaml:"manifestPath,omitempty"`
		// PreloadKeys are software keys to preload
		PreloadKeys []string `yaml:"preloadKeys,omitempty"`
	} `yaml:"software,omitempty"`

	// System settings
	System struct {
		// DebugMode enables debug logging
		DebugMode bool `yaml:"debugMode,omitempty"`
	} `yaml:"system,omitempty"`

	// ConfigPath stores the path where the config was loaded from
	ConfigPath string `yaml:"-"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	c := &Config{}

	// UI defaults
	c.UI.Theme = "dark"
	c.UI.DetailHeight = 10
	c.UI.ListHeight = 10
	c.UI.EmojisEnabled = true

	// Software defaults
	c.Software.ManifestPath = "software.yml"
	c.Software.PreloadKeys = []string{}

	// System defaults
	c.System.DebugMode = false

	return c
}

// Validate checks if the configuration is valid
// Returns nil if valid, otherwise returns an error
func (c *Config) Validate() error {
	// Validate UI theme
	validThemes := map[string]bool{
		"dark":   true,
		"light":  true,
		"system": true,
	}
	if !validThemes[c.UI.Theme] {
		return fmt.Errorf("invalid UI theme: %s (must be 'dark', 'light', or 'system')", c.UI.Theme)
	}

	// Validate UI dimensions
	if c.UI.DetailHeight < 1 {
		return fmt.Errorf("invalid detail height: %d (must be > 0)", c.UI.DetailHeight)
	}

	if c.UI.ListHeight < 1 {
		return fmt.Errorf("invalid list height: %d (must be > 0)", c.UI.ListHeight)
	}

	// Validate software manifest path
	if c.Software.ManifestPath == "" {
		return errors.New("software manifest path cannot be empty")
	}

	return nil
}

// Load loads configuration from a specific file path
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, ErrNoConfig
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// If we already have an error, don't overwrite it
			fmt.Fprintf(os.Stderr, "error closing config file: %v\n", closeErr)
		}
	}()

	c := DefaultConfig()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(c); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	c.ConfigPath = configPath
	return c, nil
}

// FindConfigFile searches for a config file in the standard locations:
// 1. Environment variable A_LA_CARTE_CONFIG
// 2. $HOME/.config/a-la-carte/a-la-carte.yml
func FindConfigFile() string {
	// Check environment variable first
	if envPath := os.Getenv(EnvConfigPath); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// Check XDG config directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			xdgConfigHome = filepath.Join(home, ".config")
		}
	}

	if xdgConfigHome != "" {
		configPath := filepath.Join(xdgConfigHome, DefaultConfigDirname, DefaultConfigFilename)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

// Save writes the configuration to the specified file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Create or truncate the file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// If we already have an error, don't overwrite it
			fmt.Fprintf(os.Stderr, "error closing config file: %v\n", closeErr)
		}
	}()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	return nil
}

// SaveToDefaultLocation saves the configuration to the default XDG config location
func (c *Config) SaveToDefaultLocation() error {
	// Get XDG config home
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting user home directory: %w", err)
		}
		xdgConfigHome = filepath.Join(home, ".config")
	}

	// Create the path
	path := filepath.Join(xdgConfigHome, DefaultConfigDirname, DefaultConfigFilename)
	return c.Save(path)
}

// CreateDefault creates a default configuration file in the default XDG location
// only if one doesn't already exist
func CreateDefault() (string, error) {
	// Check if a config file already exists
	if path := FindConfigFile(); path != "" {
		return path, nil // Config file already exists
	}

	// Create a default config
	c := DefaultConfig()

	// Save it
	if err := c.SaveToDefaultLocation(); err != nil {
		return "", err
	}

	// Return the path
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting user home directory: %w", err)
		}
		xdgConfigHome = filepath.Join(home, ".config")
	}

	path := filepath.Join(xdgConfigHome, DefaultConfigDirname, DefaultConfigFilename)
	return path, nil
}

// String returns a string representation of the configuration for debugging
func (c *Config) String() string {
	var b strings.Builder

	b.WriteString("Configuration:\n")
	b.WriteString(fmt.Sprintf("  Config Path: %s\n", c.ConfigPath))
	b.WriteString(fmt.Sprintf("  UI Theme: %s\n", c.UI.Theme))
	b.WriteString(fmt.Sprintf("  UI Detail Height: %d\n", c.UI.DetailHeight))
	b.WriteString(fmt.Sprintf("  UI List Height: %d\n", c.UI.ListHeight))
	b.WriteString(fmt.Sprintf("  UI Emojis Enabled: %v\n", c.UI.EmojisEnabled))
	b.WriteString(fmt.Sprintf("  Software Manifest Path: %s\n", c.Software.ManifestPath))
	b.WriteString(fmt.Sprintf("  System Debug Mode: %v\n", c.System.DebugMode))

	if len(c.Software.PreloadKeys) > 0 {
		b.WriteString("  Preloaded Keys:\n")
		for _, key := range c.Software.PreloadKeys {
			b.WriteString(fmt.Sprintf("    - %s\n", key))
		}
	}

	return b.String()
}
