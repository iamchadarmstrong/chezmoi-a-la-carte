// Package config provides configuration management for the a-la-carte application.
package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputFormat defines the supported output formats for the application
type OutputFormat string

const (
	// OutputFormatText represents plain text output format
	OutputFormatText OutputFormat = "text"

	// OutputFormatJSON represents JSON output format
	OutputFormatJSON OutputFormat = "json"
)

// IsValidOutputFormat checks if the given format string is a valid output format
func IsValidOutputFormat(format string) bool {
	switch OutputFormat(format) {
	case OutputFormatText, OutputFormatJSON:
		return true
	default:
		return false
	}
}

// FormatOutput formats data according to the specified output format
func FormatOutput(data interface{}, format OutputFormat) (string, error) {
	switch format {
	case OutputFormatText:
		return formatAsText(data)
	case OutputFormatJSON:
		return formatAsJSON(data)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// Helper functions for formatting output
func formatAsText(data interface{}) (string, error) {
	// Simple text formatting depends on the data type
	switch v := data.(type) {
	case string:
		return v, nil
	case []string:
		return strings.Join(v, "\n"), nil
	default:
		return fmt.Sprintf("%v", data), nil
	}
}

func formatAsJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}
	return string(jsonBytes), nil
}
