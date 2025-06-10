# Configuration System for Chezmoi à la Carte

This document explains the configuration system for the Chezmoi à la Carte application.

## Configuration Precedence

Configuration settings are loaded from the following sources in order of precedence (highest to lowest):

1. **Command line arguments**: Direct arguments override any other settings
2. **Environment variables**: `A_LA_CARTE_CONFIG` pointing to a config file
3. **Command line config flag**: `--config /path/to/config.yml`
4. **XDG config file**: `$HOME/.config/a-la-carte/a-la-carte.yml`
5. **Built-in defaults**: Fallback settings when no configuration is provided

## Command Line Arguments

The application supports the following command line arguments:

| Argument          | Short | Description                                        |
| ----------------- | ----- | -------------------------------------------------- |
| `--config FILE`   | `-c`  | Path to configuration file                         |
| `--manifest FILE` | `-m`  | Path to software manifest file                     |
| `--debug`         | `-d`  | Enable debug mode                                  |
| `--version`       | `-v`  | Show version and exit                              |
| `--help`          | `-h`  | Show help message                                  |
| `--output FORMAT` | `-o`  | Output format (text, json) for non-interactive use |
| `--quiet`         | `-q`  | Suppress non-essential output                      |
| `--no-emojis`     | `-E`  | Disable emojis in the UI                           |

### Examples

```bash
# Run with a custom config file
chezmoi-a-la-carte --config ~/.config/my-custom-config.yml

# Run with a specific manifest file
chezmoi-a-la-carte --manifest ~/projects/software-list.yml

# Enable debug mode
chezmoi-a-la-carte --debug

# Show version and exit
chezmoi-a-la-carte --version

# Show help
chezmoi-a-la-carte --help

# Use JSON output format and suppress non-essential output
chezmoi-a-la-carte --output json --quiet
```

## Environment Variables

The application recognizes the following environment variables:

| Variable            | Description                                       |
| ------------------- | ------------------------------------------------- |
| `A_LA_CARTE_CONFIG` | Path to a configuration file (highest precedence) |

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

# List of software keys to preload

preloadKeys: - git - vim - go

# System settings

system:

# Enable debug mode (can also be enabled with --debug flag)

debugMode: false

```

## Command Line Flags

The following command line flags are available:

```

--config, -c Path to configuration file
--debug, -d Enable debug mode
--version, -v Show version and exit
--help, -h Show help message
--no-emojis, -E Disable emojis in the UI

````

## Environment Variables

- `A_LA_CARTE_CONFIG`: Path to a configuration file
- `XDG_CONFIG_HOME`: Base directory for configuration files (defaults to `$HOME/.config`)

## Default Configuration

If no configuration file is found, the application will use the following default settings:

- UI Theme: dark
- Detail Height: 10
- List Height: 10
- Emojis Enabled: true
- Software Manifest Path: software.yml
- Debug Mode: false

## Creating a Configuration File

You can create a custom configuration file by copying the example:

```bash
cp a-la-carte.example.yml ~/.config/a-la-carte/a-la-carte.yml
````

Then edit the file to suit your preferences.
