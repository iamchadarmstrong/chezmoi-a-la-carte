// Package core provides the foundational elements for UI components.
// This file defines the Theme interface and provides a DefaultTheme implementation.
// Themes are used to define the color palette of the application.
//
// Usage:
// 1. Define a new theme by implementing the Theme interface.
// 2. Register the theme using `RegisterTheme("myThemeName", myThemeInstance)`.
// 3. Set the current theme using `SetThemeName("myThemeName")` or `SetTheme(myThemeInstance)`.
// 4. Access the current theme using `CurrentTheme()`.
//
// Example:
//
// type MyCustomTheme struct{}
// func (t MyCustomTheme) Primary() lipgloss.AdaptiveColor { return lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"} }
// // ... implement other Theme methods ...
//
//	func main() {
//	  core.RegisterTheme("custom", MyCustomTheme{})
//	  core.SetThemeName("custom")
//	  // ... rest of your application ...
//	  styles := core.BuildStyles() // Styles will now use MyCustomTheme
//	}
package core

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the interface for all UI themes in the application.
// It provides methods to access various color properties used throughout the UI.
type Theme interface {
	// Base colors
	Primary() lipgloss.AdaptiveColor      // Primary color, often used for main interactive elements or headers.
	Secondary() lipgloss.AdaptiveColor    // Secondary color, used for accents or less prominent elements.
	Accent() lipgloss.AdaptiveColor       // Accent color, for highlighting or calling attention to specific items.
	AccentActive() lipgloss.AdaptiveColor // Accent color for active or focused states.

	// Text colors
	Text() lipgloss.AdaptiveColor       // Default text color.
	TextMuted() lipgloss.AdaptiveColor  // Muted text color, for less important text or disabled states.
	TextActive() lipgloss.AdaptiveColor // Text color for active or selected items.

	// Background colors
	Background() lipgloss.AdaptiveColor        // Default background color for UI elements.
	BackgroundActive() lipgloss.AdaptiveColor  // Background color for active or selected items.
	BackgroundFocused() lipgloss.AdaptiveColor // Background color for focused elements, often a subtle variation.

	// Border colors
	Border() lipgloss.AdaptiveColor       // Default border color.
	BorderActive() lipgloss.AdaptiveColor // Border color for active or focused elements.

	// Dialog colors
	DialogBg() lipgloss.AdaptiveColor     // Background color for dialog boxes.
	DialogBorder() lipgloss.AdaptiveColor // Border color for dialog boxes.

	// Status bar colors
	StatusBarBg() lipgloss.AdaptiveColor // Background color for the status bar.
	StatusBarFg() lipgloss.AdaptiveColor // Foreground (text) color for the status bar.

	// Header and tab colors
	Header() lipgloss.AdaptiveColor // Color for headers or tab elements.
	SoftwarePickerHeight() int      // Defines the height of the software picker component.
	ShowSectionHeaders() bool       // Determines if section headers should be visible in components like detail views.
}

// currentTheme holds the currently active theme.
// It is a global variable within the package to allow easy access to the active theme.
var currentTheme Theme

// SetTheme sets the global currentTheme.
// This function is used to change the active theme of the application.
func SetTheme(theme Theme) {
	currentTheme = theme
}

// CurrentTheme returns the currently active theme.
// If no theme has been explicitly set, it might return nil or a default,
// depending on initialization logic (see init function).
func CurrentTheme() Theme {
	return currentTheme
}

// DefaultTheme provides a standard, fallback theme if no other theme is specified.
// It implements the Theme interface with a predefined set of colors.
type DefaultTheme struct{}

// Implement the Theme interface with the original default theme colors

// Primary returns the primary color for the DefaultTheme.
func (t DefaultTheme) Primary() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#874BFD")) // purple
}

// Secondary returns the secondary color for the DefaultTheme.
func (t DefaultTheme) Secondary() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#F25D94")) // fluorescent pink
}

// Accent returns the accent color for the DefaultTheme.
func (t DefaultTheme) Accent() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#F25D94")) // fluorescent pink
}

// AccentActive returns the active accent color for the DefaultTheme.
func (t DefaultTheme) AccentActive() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#F25D94")) // fluorescent pink active
}

// Text returns the default text color for the DefaultTheme.
func (t DefaultTheme) Text() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#EEEEEE")) // normal text
}

// TextMuted returns the muted text color for the DefaultTheme.
func (t DefaultTheme) TextMuted() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#D9DCCF")) // muted text
}

// TextActive returns the active text color for the DefaultTheme.
func (t DefaultTheme) TextActive() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#EEEEEE")) // normal for active text
}

// Background returns the default background color for the DefaultTheme.
func (t DefaultTheme) Background() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#282a36")) // dark base
}

// BackgroundActive returns the active background color for the DefaultTheme.
func (t DefaultTheme) BackgroundActive() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#6272a4")) // selected text background (darker blue)
}

// BackgroundFocused returns the focused background color for the DefaultTheme.
func (t DefaultTheme) BackgroundFocused() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#2d303f")) // very faint focus background
}

// Border returns the default border color for the DefaultTheme.
func (t DefaultTheme) Border() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#874BFD")) // purple
}

// BorderActive returns the active border color for the DefaultTheme.
func (t DefaultTheme) BorderActive() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#F25D94")) // fluorescent pink
}

// DialogBg returns the dialog background color for the DefaultTheme.
func (t DefaultTheme) DialogBg() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#FFF7DB")) // cream for dialog background
}

// DialogBorder returns the dialog border color for the DefaultTheme.
func (t DefaultTheme) DialogBorder() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#F25D94")) // pink for dialog border
}

// StatusBarBg returns the status bar background color for the DefaultTheme.
func (t DefaultTheme) StatusBarBg() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#D9DCCF")) // Light cream for status bar background
}

// StatusBarFg returns the status bar foreground color for the DefaultTheme.
func (t DefaultTheme) StatusBarFg() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#343433")) // Dark gray for status bar text
}

// Header returns the header color for the DefaultTheme.
func (t DefaultTheme) Header() lipgloss.AdaptiveColor {
	return colorToAdaptive(lipgloss.Color("#874BFD")) // purple header from original
}

// SoftwarePickerHeight returns the height for software picker elements in the DefaultTheme.
func (t DefaultTheme) SoftwarePickerHeight() int {
	return 12 // Default height for software picker elements (matching original)
}

// ShowSectionHeaders determines if section headers are shown in the DefaultTheme.
func (t DefaultTheme) ShowSectionHeaders() bool {
	return true // Default to showing section headers
}

// registeredThemes holds a map of theme names to Theme implementations.
// This allows themes to be registered and switched by name.
var registeredThemes = make(map[string]Theme)

// currentThemeName stores the name of the currently active theme.
var currentThemeName string

// RegisterTheme adds a new theme to the registeredThemes map.
// If it's the first theme being registered, it's automatically set as the current theme.
func RegisterTheme(name string, theme Theme) {
	registeredThemes[name] = theme
	// If this is the first registered theme, set it as current
	if currentThemeName == "" {
		SetThemeName(name)
	}
}

// GetThemeByName retrieves a theme from the registeredThemes map by its name.
// It returns the Theme and a boolean indicating if the theme was found.
func GetThemeByName(name string) (Theme, bool) {
	theme, exists := registeredThemes[name]
	return theme, exists
}

// SetThemeName changes the current theme to the one specified by name.
// It looks up the theme in registeredThemes and, if found, sets it using SetTheme.
func SetThemeName(name string) {
	if theme, exists := registeredThemes[name]; exists {
		SetTheme(theme)
		currentThemeName = name
	}
}

// CurrentThemeName returns the name of the currently active theme.
func CurrentThemeName() string {
	return currentThemeName
}

// init ensures that a DefaultTheme is set when the package is initialized,
// preventing nil pointer exceptions if CurrentTheme() is called before any theme is explicitly set.
func init() {
	// Set the default theme if none is specified
	SetTheme(DefaultTheme{})
}
