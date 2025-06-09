# Configuration System

This document provides an overview of the configuration system for the a-la-carte application, including the configuration loading process, components, and potential improvements.

## Components

The configuration system consists of the following main components:

1. **Config Package** (`internal/config/`):

   - Configuration structure definition
   - File loading and resolution
   - Validation logic
   - Output format support

2. **Flags Package** (`internal/flags/`):

   - Command-line flag parsing
   - Option validation
   - Usage information

3. **Main Package Integration** (`cmd/chezmoi-a-la-carte/main.go`):
   - Configuration loading orchestration
   - Error handling and reporting

## Configuration Loading Process

The system loads configuration in this order of precedence (highest to lowest):

1. **Command-line arguments** (direct overrides like `--debug`)
2. **Environment variable** (`A_LA_CARTE_CONFIG`)
3. **Command-line specified config file** (`--config`)
4. **XDG config location** (`$HOME/.config/a-la-carte/a-la-carte.yml`)
5. **Built-in defaults**

## Implementation Details

The `loadConfig()` function in `main.go` orchestrates this process:

1. It first checks for a config file path from command-line options
2. If not present, it uses `FindConfigFile()` to check environment variables and standard locations
3. It loads the configuration from the file, or uses defaults if no file is found
4. It applies any command-line overrides (like debug mode)
5. It validates the final configuration

## Current Status

All the essential components of a robust configuration system are implemented, including:

- Configuration structure definition
- File loading and precedence handling
- Command-line flag integration
- Validation logic
- Output format support
- Documentation

## Configuration File Format

The configuration file uses YAML format. Here's an example:

```yaml
# UI Configuration
ui:
  # Theme can be light, dark, or system
  theme: dark

  # UI dimensions
  detailHeight: 10
  listHeight: 10

  # Whether to show emojis in the UI
  emojisEnabled: true

# Software configuration
software:
  # Path to the software manifest
  manifestPath: software.yml

  # Software keys to preload (automatically selected when app starts)
  preloadKeys:
    - git
    - vim
    - go

# System settings
system:
  # Enable debug mode
  debugMode: false
```

## Potential Improvements

Based on the current implementation, these improvements could be made:

1. **User settings system** for persistent preferences
2. **Additional output formats** (e.g., YAML)
3. **Environment variable overrides** for individual settings (not just the config file path)
4. **UI adjustments** based on configuration settings

## Module Documentation

For more detailed information about each component:

- [Config Package Documentation](internal/config/README.md)
- [Flags Package Documentation](internal/flags/README.md)
- [Main Package Documentation](cmd/chezmoi-a-la-carte/README.md)
