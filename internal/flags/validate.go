package flags

import (
	"fmt"
	"strings"
)

// ValidateOptions validates the command line options and returns an error if invalid
func ValidateOptions(opts *Options) error {
	// Validate output format
	if !isValidOutputFormat(opts.OutputFormat) {
		return fmt.Errorf("invalid output format: %s (must be 'text' or 'json')", opts.OutputFormat)
	}

	return nil
}

// isValidOutputFormat checks if the given format is valid
func isValidOutputFormat(format string) bool {
	validFormats := map[string]bool{
		"text": true,
		"json": true,
	}

	return validFormats[strings.ToLower(format)]
}
