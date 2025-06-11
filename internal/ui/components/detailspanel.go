package components

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea" // Added import
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"a-la-carte/internal/ui/core"
)

// DetailsPanelData holds all lines to display in the details panel, already formatted.
type DetailsPanelData struct {
	Lines []string
}

// DetailsPanelModel is a model for rendering the details panel in the TUI.
type DetailsPanelModel struct {
	data      *DetailsPanelData
	scroll    int
	maxScroll int
	focused   bool
	width     int
	height    int
}

// Init does nothing for this model.
func (d *DetailsPanelModel) Init() tea.Cmd { return nil }

// Update does nothing for this model.
func (d *DetailsPanelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return d, nil }

// NewDetailsPanelModel creates a new DetailsPanelModel.
//
// # Parameters
//   - data: pointer to DetailsPanelData
//   - scroll: current scroll offset
//   - maxScroll: maximum scroll offset
//   - focused: whether the panel is focused
//   - width: panel width
//   - height: panel height
//
// # Returns
//   - pointer to DetailsPanelModel
func NewDetailsPanelModel(data *DetailsPanelData, scroll, maxScroll int, focused bool, width, height int) *DetailsPanelModel {
	return &DetailsPanelModel{
		data:      data,
		scroll:    scroll,
		maxScroll: maxScroll,
		focused:   focused,
		width:     width,
		height:    height,
	}
}

// SetData updates the data for the details panel.
func (d *DetailsPanelModel) SetData(data *DetailsPanelData) {
	d.data = data
}

// SetFocused updates the focused state of the panel.
func (d *DetailsPanelModel) SetFocused(focused bool) {
	d.focused = focused
}

// SetScroll updates the scroll position of the panel.
func (d *DetailsPanelModel) SetScroll(scroll int) {
	d.scroll = scroll
	// Optionally, re-calculate maxScroll if data can change independently
	// and affect maxScroll. For now, assume maxScroll is managed externally or
	// alongside data changes.
}

// SetDimensions updates the width and height of the panel.
func (d *DetailsPanelModel) SetDimensions(width, height int) {
	d.width = width
	d.height = height
}

// View renders the details panel as a string.
func (d *DetailsPanelModel) View() string {
	return renderDetailsPanel(d.data, d.scroll, d.focused, d.width, d.height)
}

// Helper: getIndicator returns the indicator string for the focused panel
func getIndicator(scroll, maxScroll int, focused bool) string {
	if !focused || maxScroll <= 0 {
		return ""
	}

	// Use our theme-based indicator style to ensure consistent styling
	indicatorStyle := core.IndicatorStyle(focused) // Updated to use core.IndicatorStyle(focused)

	if scroll == 0 {
		return indicatorStyle.Render("↓")
	}
	if scroll == maxScroll {
		return indicatorStyle.Render("↑")
	}
	return indicatorStyle.Render("↑↓")
}

// Helper: prepareVisibleLines prepares the visible lines for the focused panel
func prepareVisibleLines(data *DetailsPanelData, scroll, height, width int, indicator string) []string {
	visible := make([]string, 0, height)
	for i := 0; i < height; i++ {
		idx := scroll + i
		if idx < len(data.Lines) {
			visible = append(visible, data.Lines[idx])
		} else {
			visible = append(visible, "")
		}
	}
	if indicator != "" && len(visible) > 0 {
		lastIdx := len(visible) - 1
		lastLine := visible[lastIdx]
		// Calculate available width for the last line (panel width minus indicator width)
		indicatorWidth := runewidth.StringWidth(indicator)
		maxLineWidth := width - indicatorWidth
		if maxLineWidth < 0 {
			maxLineWidth = 0
		}
		// Truncate last line if needed
		if runewidth.StringWidth(lastLine) > maxLineWidth {
			lastLine = truncateString(lastLine, maxLineWidth)
		}
		// Append indicator
		padLen := maxLineWidth - runewidth.StringWidth(lastLine)
		if padLen > 0 {
			lastLine += strings.Repeat(" ", padLen)
		}
		visible[lastIdx] = lastLine + indicator
	}
	return visible
}

// Helper: renderNotFocusedPanel handles the not-focused rendering logic
func renderNotFocusedPanel(data *DetailsPanelData, width, height int) string {
	details := lipgloss.NewStyle().Width(width).Height(height)
	maxScroll := len(data.Lines) - height
	header := ""
	if len(data.Lines) > 0 {
		header = data.Lines[0]
	}
	indicatorLine := ""
	if maxScroll > 0 {
		// Use muted style for not-focused indicators
		indicatorStyle := core.IndicatorStyle(false) // Updated to use core.IndicatorStyle(false)
		indicatorLine = "  " + indicatorStyle.Render("▼") + " more..."
		if runewidth.StringWidth(indicatorLine) > width {
			maxLen := width - runewidth.StringWidth("  ") - runewidth.StringWidth(indicatorStyle.Render("▼"))
			if maxLen > 0 {
				indicatorLine = "  " + indicatorStyle.Render("▼") + " " + truncateString("more...", maxLen)
			} else {
				indicatorLine = "  " + indicatorStyle.Render("▼")
			}
		}
	}
	contentLines := []string{}
	for i := 1; i < len(data.Lines); i++ {
		contentLines = append(contentLines, data.Lines[i])
	}
	linesAvailable := height
	if header != "" {
		linesAvailable--
	}
	if indicatorLine != "" {
		linesAvailable--
	}
	if linesAvailable < 0 {
		linesAvailable = 0
	}
	visible := []string{}
	if header != "" {
		visible = append(visible, header)
	}
	for i := 0; i < linesAvailable && i < len(contentLines); i++ {
		visible = append(visible, contentLines[i])
	}
	if indicatorLine != "" {
		visible = append(visible, indicatorLine)
	}
	if len(visible) > height {
		visible = visible[:height]
	}
	for len(visible) < height {
		visible = append(visible, "")
	}
	content := lipgloss.JoinVertical(lipgloss.Left, visible...)
	return details.Render(content)
}

// stripANSI removes ANSI escape codes from a string.
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(str string) string {
	return ansiRegexp.ReplaceAllString(str, "")
}

func renderDetailsPanel(data *DetailsPanelData, scroll int, focused bool, width, height int) string {
	// The `width` passed here is the internal content width for the DetailsPanelModel.
	// The container around it will handle its own padding and borders.
	detailsStyle := lipgloss.NewStyle().Width(width).Height(height) // Use the passed width directly

	if data == nil || len(data.Lines) == 0 {
		return detailsStyle.Render("No details available.")
	}

	if !focused {
		return renderNotFocusedPanel(data, width, height)
	}

	maxScroll := len(data.Lines) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	indicator := getIndicator(scroll, maxScroll, focused)
	// Pass the model's internal width to prepareVisibleLines
	visible := prepareVisibleLines(data, scroll, height, width, "")
	content := lipgloss.JoinVertical(lipgloss.Left, visible...)
	panel := detailsStyle.Render(content) // Render with the constrained width

	if indicator != "" {
		indicatorStyle := core.IndicatorStyle(focused)
		indicatorStyled := indicatorStyle.Render(indicator)
		panelLines := strings.Split(panel, "\n")
		lastIdx := len(panelLines) - 1
		if lastIdx < 0 { // Should not happen if panel has content
			return panel
		}
		lastLine := panelLines[lastIdx]
		visibleLastLine := stripANSI(lastLine)
		lastLineWidth := runewidth.StringWidth(visibleLastLine)
		indicatorWidth := runewidth.StringWidth(indicatorStyled) // Use styled indicator width

		// The indicator should be placed within the `width` of the DetailsPanelModel's content area.
		// No extra padding like `rightPadding` is needed here as the container handles outer padding.
		padLen := width - (lastLineWidth + indicatorWidth)

		if padLen < 0 {
			// Truncate visibleLastLine to make space for the indicator within `width`
			truncateAt := width - indicatorWidth
			if truncateAt < 0 {
				truncateAt = 0
			}
			visibleLastLine = runewidth.Truncate(visibleLastLine, truncateAt, "...") // Use "..." for truncation
			lastLineWidth = runewidth.StringWidth(visibleLastLine)                   // Recalculate width after truncation
			padLen = width - (lastLineWidth + indicatorWidth)                        // Recalculate padLen
			if padLen < 0 {
				padLen = 0
			} // Ensure padLen is not negative
		}
		paddedLastLine := visibleLastLine + strings.Repeat(" ", padLen) + indicatorStyled
		panelLines[lastIdx] = paddedLastLine
		return strings.Join(panelLines, "\n")
	}
	return panel
}

// truncateString truncates s to fit maxWidth (in runewidth columns), appending '…' if truncated.
func truncateString(s string, maxWidth int) string {
	w := 0
	for i, r := range s {
		w += runewidth.RuneWidth(r)
		if w > maxWidth {
			if i == 0 {
				return "…"
			}
			return s[:i] + "…"
		}
	}
	return s
}
