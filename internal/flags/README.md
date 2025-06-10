# Flags Package

This package provides command line flag handling for the a-la-carte application.

## Overview

The `flags` package manages command-line options, parses arguments, and provides validation for the application's command-line interface.

## Command Line Options

The package defines the following command-line options:

| Argument          | Short | Description                                        | Default |
| ----------------- | ----- | -------------------------------------------------- | ------- |
| `--config FILE`   | `-c`  | Path to configuration file                         | ""      |
| `--manifest FILE` | `-m`  | Path to software manifest file                     | ""      |
| `--debug`         | `-d`  | Enable debug mode                                  | false   |
| `--version`       | `-v`  | Show version and exit                              | false   |
| `--help`          | `-h`  | Show help message                                  | false   |
| `--output FORMAT` | `-o`  | Output format (text, json) for non-interactive use | "text"  |
| `--quiet`         | `-q`  | Suppress non-essential output                      | false   |
| `--no-emojis`     | `-E`  | Disable emojis in the UI                           | false   |

## Main Functions

- `Parse()`: Parses command line flags and returns the options
- `ValidateOptions(opts *Options)`: Validates the command line options
- `Usage()`: Prints detailed usage information

## Output Format Validation

The package validates output format options:

- Currently supported: "text" and "json"
- Returns an error for invalid formats

## Example Usage

```go
// Parse command-line flags
opts := flags.Parse()

// Validate command-line options
if err := flags.ValidateOptions(opts); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    flags.Usage()
    os.Exit(1)
}

// Handle help flag
if opts.Help {
    flags.Usage()
    return
}

// Use the parsed options
if opts.Debug {
    // Enable debug mode
}

if opts.Version {
    // Show version information
}
```

## Integration with Configuration

The flags package is designed to work seamlessly with the `config` package:

1. Command-line flags are parsed first
2. The `--config` flag (if provided) points to a custom config file
3. Other flags can override settings from the config file

## Potential Improvements

- Support for environment variable overrides for individual settings
- Additional output formats beyond text and JSON
- Command completion for shells like bash/zsh
- Interactive help with examples
