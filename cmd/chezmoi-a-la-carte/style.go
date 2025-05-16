package main

import "github.com/charmbracelet/lipgloss"

// Theme holds all color values for the TUI.
type Theme struct {
	Primary    lipgloss.Color
	Accent     lipgloss.Color
	Border     lipgloss.Color
	Background lipgloss.Color
	Header     lipgloss.Color
	SelectedBg lipgloss.Color
	SelectedFg lipgloss.Color
	ItemFg     lipgloss.Color
	DetailKey  lipgloss.Color
	DetailVal  lipgloss.Color
	Search     lipgloss.Color
	Footer     lipgloss.Color
	Focus      lipgloss.Color
}

// DefaultTheme is a Nord-inspired color palette.
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("#8be9fd"),
	Accent:     lipgloss.Color("#ff79c6"),
	Border:     lipgloss.Color("#7dcfff"),
	Background: lipgloss.Color("#282a36"),
	Header:     lipgloss.Color("#ff79c6"),
	SelectedBg: lipgloss.Color("#282a36"),
	SelectedFg: lipgloss.Color("#f8f8f2"),
	ItemFg:     lipgloss.Color("#c0caf5"),
	DetailKey:  lipgloss.Color("#7dcfff"),
	DetailVal:  lipgloss.Color("#c0caf5"),
	Search:     lipgloss.Color("#99a7bf"),
	Footer:     lipgloss.Color("#6e738d"),
	Focus:      lipgloss.Color("#51e1a6"),
}

// Styles holds all Lip Gloss styles for the TUI.
type Styles struct {
	BorderStyle       lipgloss.Style
	HeaderStyle       lipgloss.Style
	DetailHeaderStyle lipgloss.Style
	SelectedItemStyle lipgloss.Style
	ItemStyle         lipgloss.Style
	DetailKey         lipgloss.Style
	DetailVal         lipgloss.Style
	SearchStyle       lipgloss.Style
	ListPanel         lipgloss.Style
	DetailPanel       lipgloss.Style
	FooterStyle       lipgloss.Style
	FocusStyle        lipgloss.Style
}

// NewStyles returns all styles for a given theme.
func NewStyles(theme *Theme) Styles {
	return Styles{
		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary).
			Padding(1, 2).
			Width(panelWidth),
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Header).
			Padding(0, 1).
			Width(panelWidth - 4).
			Align(lipgloss.Center),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			Padding(0, 1).
			Width(panelWidth - 10).
			Align(lipgloss.Center),
		SelectedItemStyle: lipgloss.NewStyle().
			Foreground(theme.SelectedFg).
			Background(theme.SelectedBg).
			Bold(true).
			Width(panelWidth - 8),
		ItemStyle: lipgloss.NewStyle().
			Foreground(theme.ItemFg).
			Padding(0, 1).
			Width(panelWidth - 8),
		DetailKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.DetailKey),
		DetailVal: lipgloss.NewStyle().
			Foreground(theme.DetailVal),
		SearchStyle: lipgloss.NewStyle().
			Foreground(theme.Search).
			Bold(true),
		ListPanel: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(theme.Border).
			Padding(0, 2).
			Margin(0, 0).
			Width(panelWidth - 6),
		DetailPanel: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(theme.Border).
			Padding(0, 2).
			Margin(0, 0).
			Width(panelWidth - 6),
		FooterStyle: lipgloss.NewStyle().
			Foreground(theme.Footer).
			Padding(0, 1).
			Width(panelWidth - 4),
		FocusStyle: lipgloss.NewStyle().
			Foreground(theme.Focus).
			Bold(true),
	}
}

// Global styles instance (can be swapped for theming)
var styles = NewStyles(&DefaultTheme)
