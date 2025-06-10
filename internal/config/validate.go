package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateManifestPath checks if the manifest path exists and is readable
func (c *Config) ValidateManifestPath() error {
	// If the manifest path is relative, prepend the config directory
	manifestPath := c.Software.ManifestPath
	if !filepath.IsAbs(manifestPath) && c.ConfigPath != "" {
		configDir := filepath.Dir(c.ConfigPath)
		manifestPath = filepath.Join(configDir, manifestPath)
	}

	// Check if the manifest file exists and is readable
	info, err := os.Stat(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("manifest file not found: %s", manifestPath)
		}
		return fmt.Errorf("error accessing manifest file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("manifest path is a directory, not a file: %s", manifestPath)
	}

	// Try to open the file to check if it's readable
	f, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("error opening manifest file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// If we already have an error, don't overwrite it
			fmt.Fprintf(os.Stderr, "error closing config file: %v\n", closeErr)
		}
	}()

	return nil
}

// ResolveManifestPath returns the absolute path to the manifest file
func (c *Config) ResolveManifestPath() string {
	manifestPath := c.Software.ManifestPath

	// If it's already absolute, return it
	if filepath.IsAbs(manifestPath) {
		return manifestPath
	}

	// If we have a config file path, make the manifest path relative to it
	if c.ConfigPath != "" {
		configDir := filepath.Dir(c.ConfigPath)
		return filepath.Join(configDir, manifestPath)
	}

	// Otherwise, use the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to just the relative path if we can't get the working directory
		return manifestPath
	}

	return filepath.Join(cwd, manifestPath)
}
