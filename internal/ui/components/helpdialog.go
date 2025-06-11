// helpdialog.go provides a component for displaying help dialogs.
package components

import (
	"a-la-carte/internal/ui/core"
	"a-la-carte/internal/ui/patterns" // Updated from layouts to patterns

	"github.com/charmbracelet/lipgloss"
)

// HelpDialogModel represents a help dialog component.
type HelpDialogModel struct {
	title   string
	content string
	footer  string
	width   int
	height  int
	visible bool
}

// NewHelpDialogModel creates a new help dialog model.
func NewHelpDialogModel() *HelpDialogModel {
	return &HelpDialogModel{
		title:   "Help",
		content: "q: Quit  h: Toggle Help  /: Search  TAB: Toggle Details  ↑/↓/j/k: Move  Enter: Select/Deselect",
		footer:  "Press 'h' to close this help dialog.",
		width:   core.PanelWidth,
		height:  3, // Default height for basic content
		visible: false,
	}
}

// SetTitle sets the help dialog title.
func (m *HelpDialogModel) SetTitle(title string) *HelpDialogModel {
	m.title = title
	return m
}

// SetContent sets the help dialog content.
func (m *HelpDialogModel) SetContent(content string) *HelpDialogModel {
	m.content = content
	return m
}

// SetFooter sets the help dialog footer.
func (m *HelpDialogModel) SetFooter(footer string) *HelpDialogModel {
	m.footer = footer
	return m
}

// Show makes the help dialog visible.
func (m *HelpDialogModel) Show() {
	m.visible = true
}

// Hide hides the help dialog.
func (m *HelpDialogModel) Hide() {
	m.visible = false
}

// Toggle toggles the visibility of the help dialog.
func (m *HelpDialogModel) Toggle() {
	m.visible = !m.visible
}

// IsVisible returns whether the help dialog is visible.
func (m *HelpDialogModel) IsVisible() bool {
	return m.visible
}

// View renders the help dialog.
func (m *HelpDialogModel) View() string {
	if !m.visible {
		return ""
	}

	styles := core.CurrentStyles() // Updated from ui.CurrentStyles()

	title := styles.TitleHeaderStyle.Render(m.title)
	content := styles.ItemStyle.Render(m.content)
	footer := styles.FooterStyle.Render(m.footer)

	// Combine all parts of the dialog
	helpContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		footer,
	)

	// Use Dialog to create the container
	return patterns.Dialog(core.StringModel(helpContent)).View() // Updated from ui.Dialog and ui.StringModel
}

// ApplyOverlay applies the help dialog as an overlay on the base content.
func (m *HelpDialogModel) ApplyOverlay(baseContent string, centerX, centerY int) string {
	if !m.visible {
		return baseContent
	}

	helpContainer := m.View()
	// Assuming centerX maps to panelWidth and centerY to panelHeight for the area
	// in which the helpContainer should be centered.
	return patterns.PlaceOverlay(centerX, centerY, helpContainer, baseContent, true)
}
