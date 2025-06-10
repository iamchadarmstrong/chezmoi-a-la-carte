# Main Package

The main package serves as the entry point for the chezmoi-a-la-carte application and orchestrates the configuration loading process.

## Configuration Loading Process

The `loadConfig()` function in `main.go` coordinates the configuration loading with the correct precedence:

1. It first checks for a config file path from command-line options
2. If not present, it uses `FindConfigFile()` to check environment variables and standard locations
3. It loads the configuration from the file, or uses defaults if no file is found
4. It applies any command-line overrides (like debug mode)
5. It validates the final configuration

## Configuration Precedence

The system loads configuration in this order of precedence (highest to lowest):

1. **Command-line arguments** (direct overrides like `--debug`)
2. **Environment variable** (`A_LA_CARTE_CONFIG`)
3. **Command-line specified config file** (`--config`)
4. **XDG config location** (`$HOME/.config/a-la-carte/a-la-carte.yml`)
5. **Built-in defaults**

## Implementation Flow

```
main()
  |
  ├── Parse command-line flags (flags.Parse())
  |
  ├── Validate command-line options (flags.ValidateOptions())
  |
  ├── Handle special flags (--help, --version)
  |
  ├── Load configuration (loadConfig())
  |     |
  |     ├── Check for config file from --config flag
  |     |
  |     ├── If not found, search standard locations (config.FindConfigFile())
  |     |
  |     ├── Load config from file or use defaults
  |     |
  |     ├── Apply command-line overrides
  |     |
  |     └── Validate configuration
  |
  ├── Initialize application model
  |     |
  |     ├── Validate manifest path
  |     |
  |     ├── Resolve manifest path
  |     |
  |     └── Load software manifest
  |
  └── Run the application
```

## Debug Output

When debug mode is enabled, the application will:

1. Print "Debug mode enabled"
2. Display the complete configuration
3. Show the resolved manifest path

## Error Handling

The application handles various error conditions:

- Configuration errors (file not found, invalid format)
- Manifest validation errors
- Initialization errors
- Runtime errors

## Roadmap Checklist

### Implemented Features

- [x] User settings system for persistent preferences (via XDG config)
- [x] UI adjustments based on configuration settings (theme, dimensions, emojis)
- [x] Command-line configuration overrides (--config, --debug)
- [x] Environment variable for config file path (A_LA_CARTE_CONFIG)
- [x] Configuration validation and error handling
- [x] Default configuration fallback
- [x] Debug mode with detailed output
- [x] Theme system with light/dark/system options

### Planned Improvements

- [ ] Additional output formats (e.g., YAML)
- [ ] Environment variable overrides for individual settings
