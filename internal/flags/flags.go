// Package flags provides command line flag handling for the a-la-carte application
package flags

import (
	"flag"
	"fmt"
)

// Options defines the command line options for the application
type Options struct {
	// ConfigPath is the path to the configuration file
	ConfigPath string

	// ManifestPath is the path to the software manifest file
	ManifestPath string

	// Debug enables debug mode
	Debug bool

	// Version shows the version and exits
	Version bool

	// Help shows the help message and exits
	Help bool

	// OutputFormat defines the output format for non-interactive commands
	OutputFormat string

	// Quiet suppresses non-essential output
	Quiet bool

	// NoEmojis disables emoji display in the UI
	NoEmojis bool
}

// Parse parses command line flags and returns the options
func Parse() *Options {
	opts := &Options{}

	// Define long-form flags
	flag.StringVar(&opts.ConfigPath, "config", "", "Path to configuration file")
	flag.StringVar(&opts.ManifestPath, "manifest", "", "Path to software manifest file")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&opts.Version, "version", false, "Show version and exit")
	flag.BoolVar(&opts.Help, "help", false, "Show help message")
	flag.StringVar(&opts.OutputFormat, "output", "text", "Output format (text, json)")
	flag.BoolVar(&opts.Quiet, "quiet", false, "Suppress non-essential output")
	flag.BoolVar(&opts.NoEmojis, "no-emojis", false, "Disable emojis in the UI")

	// Define short aliases
	flag.StringVar(&opts.ConfigPath, "c", "", "Path to configuration file (shorthand)")
	flag.StringVar(&opts.ManifestPath, "m", "", "Path to software manifest file (shorthand)")
	flag.BoolVar(&opts.Debug, "d", false, "Enable debug mode (shorthand)")
	flag.BoolVar(&opts.Version, "v", false, "Show version and exit (shorthand)")
	flag.BoolVar(&opts.Help, "h", false, "Show help message (shorthand)")
	flag.StringVar(&opts.OutputFormat, "o", "text", "Output format (shorthand)")
	flag.BoolVar(&opts.Quiet, "q", false, "Suppress non-essential output (shorthand)")
	flag.BoolVar(&opts.NoEmojis, "E", false, "Disable emojis in the UI (shorthand)")

	flag.Parse()
	return opts
}

// Usage prints usage information
func Usage() {
	fmt.Println("Usage: chezmoi-a-la-carte [options]")
	fmt.Println("\nA terminal user interface (TUI) for browsing and managing software manifests.")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()

	fmt.Println("\nConfiguration:")
	fmt.Println("  Configuration is loaded from the following sources in order of precedence:")
	fmt.Println("  1. Environment variable: A_LA_CARTE_CONFIG=/path/to/config.yml")
	fmt.Println("  2. Command line flag: --config /path/to/config.yml")
	fmt.Println("  3. Default location: $HOME/.config/a-la-carte/a-la-carte.yml")
	fmt.Println("  4. Built-in defaults")

	fmt.Println("\nKeyboard Controls:")
	fmt.Println("  ↑/↓/j/k:  Move selection")
	fmt.Println("  /:        Start search")
	fmt.Println("  q:        Quit")
	fmt.Println("  Enter:    Show details")
	fmt.Println("  esc:      Cancel search")
	fmt.Println("  TAB:      Toggle focus between list and details")

	fmt.Println("\nExamples:")
	fmt.Println("  # Run with a custom config file")
	fmt.Println("  chezmoi-a-la-carte --config /path/to/config.yml")
	fmt.Println()
	fmt.Println("  # Run with a specific manifest file")
	fmt.Println("  chezmoi-a-la-carte --manifest /path/to/software.yml")
	fmt.Println()
	fmt.Println("  # Run in debug mode")
	fmt.Println("  chezmoi-a-la-carte --debug")
	fmt.Println()
	fmt.Println("  # Disable emoji display in the UI")
	fmt.Println("  chezmoi-a-la-carte --no-emojis")
	fmt.Println()
	fmt.Println("  # Output in JSON format (for scripting)")
	fmt.Println("  chezmoi-a-la-carte --output json --quiet")
}
