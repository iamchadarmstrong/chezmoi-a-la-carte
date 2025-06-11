// Package core provides the foundational elements for UI components.
package core

import (
	"github.com/charmbracelet/lipgloss"
)

// colorToAdaptive converts a lipgloss.Color to an AdaptiveColor
// that works well in both light and dark terminal backgrounds.
func colorToAdaptive(color lipgloss.Color) lipgloss.AdaptiveColor {
	// We'll use the same color for both light and dark backgrounds for now
	// In a more sophisticated implementation, we could adjust colors based on their brightness
	colorStr := string(color)
	return lipgloss.AdaptiveColor{
		Light: colorStr,
		Dark:  colorStr,
	}
}
