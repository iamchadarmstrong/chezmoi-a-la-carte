# Configuration Package

This package provides configuration management for the a-la-carte application.

## Overview

The `config` package manages the application's configuration, including loading from various sources, validation, and providing utilities for manipulating configuration data.

## Configuration Precedence

Configuration settings are loaded from the following sources in order of precedence (highest to lowest):

1. **Command-line arguments** (direct overrides like `--debug`)
2. **Environment variable** (`A_LA_CARTE_CONFIG`) pointing to a config file
3. **Command-line specified config file** (`--config`)
4. **XDG config location** (`$HOME/.config/a-la-carte/a-la-carte.yml`)
5. **Built-in defaults**

## Configuration Structure

The main configuration struct includes:

- **UI settings**: Theme, layout dimensions, emoji support
- **Software settings**: Manifest path, preload keys
- **System settings**: Debug mode, etc.

## Main Functions

- `DefaultConfig()`: Returns a configuration with default values
- `Load(configPath string)`: Loads configuration from a specific file path
- `FindConfigFile()`: Searches for a config file in standard locations
- `Validate()`: Validates the configuration values
- `Save(path string)`: Writes configuration to a file
- `SaveToDefaultLocation()`: Saves to the default XDG config location

## Output Format Handling

The package includes support for different output formats:

- Text output (default human-readable format)
- JSON output (for scripting/programmatic use)

## Validation

Configuration validation includes:

- Validating UI theme values
- Checking UI dimensions are positive
- Verifying manifest path exists and is readable
- Resolving relative paths to absolute paths

## Example Usage

```go
// Load configuration with proper precedence
cfg, err := config.Load(configPath)
if err != nil {
    // Handle error
}

// Override with command-line flags if needed
if commandLineDebug {
    cfg.System.DebugMode = true
}

// Validate configuration
if err := cfg.Validate(); err != nil {
    // Handle validation error
}

// Use configuration
manifestPath := cfg.ResolveManifestPath()
```

## Potential Improvements

- User settings system for persistent preferences
- Additional output formats (e.g., YAML)
- Environment variable overrides for individual settings (not just the config file path)
- UI adjustments based on configuration settings
