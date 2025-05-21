// Package main provides the entry point and TUI logic for chezmoi-a-la-carte.
//
// # Overview
//
// This package implements a terminal user interface (TUI) for browsing and managing
// software manifests using the Bubble Tea framework. It features:
//   - Searchable, scrollable list of software entries
//   - Details panel with rich formatting and emoji icons
//   - Keyboard navigation and accessibility
//
// # Usage
//
//	go run ./cmd/chezmoi-a-la-carte
//
// # Keyboard Controls
//
//   - ↑/↓/j/k: Move selection
//   - /:       Start search
//   - q:       Quit
//   - Enter:   Show details
//   - esc:     Cancel search
//   - TAB:     Toggle focus between list and details
//
// # Example
//
//	$ go run ./cmd/chezmoi-a-la-carte
//	# Launches the TUI
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"a-la-carte/internal/app"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/mattn/go-runewidth"
)

const (
	panelWidth         = 120
	listHeight         = 12
	detailHeight       = 7
	detailHeightExpand = 16
)

type focusArea int

const (
	focusSoftware focusArea = iota // either left or right pane
	focusDetails
)

// model defines the state of the TUI.
//
// # Fields
//
//   - manifest:     The loaded software manifest.
//   - loadErr:      Any error encountered during manifest loading.
//   - entries:      All manifest keys, sorted.
//   - visible:      Filtered keys based on search.
//   - selected:     Index of the currently selected entry.
//   - search:       Current search query.
//   - searching:    Whether the search input is active.
//   - focus:        Which panel is focused (list, details, or selected).
//   - detailScroll: Scroll offset for the details panel.
//   - selectedKeys: Keys of software selected for the right pane.
//   - softwarePaneLeft: Track which pane is active in software focus: true=left, false=right
type model struct {
	manifest     app.Manifest
	loadErr      error
	entries      []string // sorted keys
	visible      []string // filtered keys (left pane, excludes selected)
	selected     int      // index in visible (left) or selectedKeys (right)
	search       string
	searching    bool
	focus        focusArea
	detailScroll int

	selectedKeys []string // keys of selected software (right pane)
	// track which pane is active in software focus: true=left, false=right
	softwarePaneLeft bool
}

func (m *model) filter() {
	if m.search == "" {
		m.visible = make([]string, 0, len(m.entries))
		for _, k := range m.entries {
			if !m.isSelected(k) {
				m.visible = append(m.visible, k)
			}
		}
	} else {
		var filtered []string
		q := strings.ToLower(m.search)
		for _, k := range m.entries {
			if m.isSelected(k) {
				continue
			}
			e := m.manifest[k]
			if strings.Contains(strings.ToLower(k), q) || strings.Contains(strings.ToLower(e.Name), q) || strings.Contains(strings.ToLower(e.Desc), q) {
				filtered = append(filtered, k)
			}
		}
		m.visible = filtered
	}
	if m.selected >= len(m.visible) {
		m.selected = len(m.visible) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m *model) isSelected(key string) bool {
	for _, k := range m.selectedKeys {
		if k == key {
			return true
		}
	}
	return false
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) handleSearchInput(key string) *model {
	if key == "enter" {
		m.searching = false
		return m
	}
	if key == "esc" {
		m.searching = false
		m.search = ""
		m.filter()
		return m
	}
	if key == "backspace" && m.search != "" {
		m.search = m.search[:len(m.search)-1]
		m.filter()
		return m
	}
	if len(key) == 1 && key >= " " && key <= "~" {
		m.search += key
		m.filter()
		return m
	}
	return m
}

func (m *model) handleDetailsInput(key string) *model {
	detailLines := m.detailLines()
	maxScroll := len(detailLines) - detailHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	switch key {
	case "up", "k":
		if m.detailScroll > 0 {
			m.detailScroll--
		}
		return m
	case "down", "j":
		if m.detailScroll < maxScroll {
			m.detailScroll++
		}
		return m
	}
	return m
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		key := keyMsg.String()
		if key == "q" {
			return m, tea.Quit
		}
		if m.loadErr != nil {
			return m, nil
		}
		if m.searching {
			return m.handleSearchInput(key), nil
		}
		if key == "tab" {
			return m.handleTab(), nil
		}
		if m.focus == focusSoftware {
			return m.handleSoftwareKey(key), nil
		}
		if m.focus == focusDetails {
			return m.handleDetailsInput(key), nil
		}
	}
	return m, nil
}

// handleTab toggles focus between software and details
func (m *model) handleTab() *model {
	if m.focus == focusSoftware {
		m.focus = focusDetails
		m.detailScroll = 0
	} else {
		m.focus = focusSoftware
		// keep softwarePaneLeft as is
	}
	return m
}

// handleSoftwareKey handles key input for the software panes
func (m *model) handleSoftwareKey(key string) *model {
	if key == "/" {
		m.searching = true
		return m
	}
	if m.softwarePaneLeft {
		return m.handleLeftPaneKey(key)
	} else {
		return m.handleRightPaneKey(key)
	}
}

// handleLeftPaneKey handles key input for the left (unselected) pane
func (m *model) handleLeftPaneKey(key string) *model {
	switch key {
	case "enter":
		m.moveToSelected()
	case "down", "j":
		if m.selected < len(m.visible)-1 {
			m.selected++
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "right":
		// switch to right pane if any selected
		if len(m.selectedKeys) > 0 {
			m.softwarePaneLeft = false
			m.selected = 0
		}
	}
	return m
}

// handleRightPaneKey handles key input for the right (selected) pane
func (m *model) handleRightPaneKey(key string) *model {
	switch key {
	case "enter":
		m.moveToDeselected()
	case "down", "j":
		if m.selected < len(m.selectedKeys)-1 {
			m.selected++
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "left":
		// switch to left pane if any visible
		if len(m.visible) > 0 {
			m.softwarePaneLeft = true
			m.selected = 0
		}
	}
	return m
}

// normalizeEmoji ensures the emoji is exactly 2 columns wide for consistent alignment.
//
// # Example
//
//	emoji := normalizeEmoji("🐍")
func normalizeEmoji(e string) string {
	w := runewidth.StringWidth(e)
	if w == 2 {
		return e
	}
	if w > 2 {
		runes := []rune(e)
		acc := 0
		for i, r := range runes {
			acc += runewidth.RuneWidth(r)
			if acc >= 2 {
				return string(runes[:i+1])
			}
		}
		return string(runes[:1]) + " "
	}
	// pad with space if too narrow
	return e + strings.Repeat(" ", 2-w)
}

// checkContains checks if either name or desc contains any of the given strings.
//
// # Parameters
//   - name:    The name string to check.
//   - desc:    The description string to check.
//   - matches: List of substrings to search for.
//
// # Returns
//   - true if any match is found in name or desc; false otherwise.
func checkContains(name, desc string, matches ...string) bool {
	n := strings.ToLower(name)
	d := strings.ToLower(desc)
	for _, m := range matches {
		if strings.Contains(n, m) || strings.Contains(d, m) {
			return true
		}
	}
	return false
}

type emojiRule struct {
	matches []string
	emoji   string
}

var emojiRules = []emojiRule{
	{matches: []string{"python"}, emoji: "🐍"},
	{matches: []string{"node", "node.js"}, emoji: "🟩"},
	{matches: []string{"go", "golang"}, emoji: "🐹"},
	{matches: []string{"docker"}, emoji: "🐳"},
	{matches: []string{"git"}, emoji: "🌱"},
	{matches: []string{"linux"}, emoji: "🐧"},
	{matches: []string{"mac", "apple"}, emoji: "🍏"},
	{matches: []string{"brew"}, emoji: "🍺"},
	{matches: []string{"terminal", "cli", "tui"}, emoji: "💻"},
	{matches: []string{"test", "testing"}, emoji: "🧪"},
	{matches: []string{"file", "document"}, emoji: "📄"},
	{matches: []string{"key", "password", "secret"}, emoji: "🔑"},
	{matches: []string{"sync", "update"}, emoji: "🔄"},
	{matches: []string{"note", "write"}, emoji: "📝"},
	{matches: []string{"package", "install"}, emoji: "📦"},
	{matches: []string{"tool", "utility"}, emoji: "🧰"},
}

func emojiForEntry(e *app.SoftwareEntry) string {
	for _, rule := range emojiRules {
		if checkContains(e.Name, e.Desc, rule.matches...) {
			return normalizeEmoji(rule.emoji)
		}
	}
	return normalizeEmoji("📦") // default emoji
}

// wrap returns the string s wrapped to the given width using lipgloss styling.
//
// # Example
//
//	wrapped := wrap("some long text", 40)
func wrap(s string, width int) string {
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(s)
}

// detailLines returns the lines to display in the details panel for the selected entry.
//
// # Returns
//   - []string: Each string is a line to display in the details panel.
func (m *model) detailLines() []string {
	if m.focus == focusSoftware && !m.softwarePaneLeft {
		// Right pane (selected)
		if len(m.selectedKeys) == 0 || m.selected < 0 || m.selected >= len(m.selectedKeys) {
			return m.noDetails()
		}
		return m.detailsForKey(m.selectedKeys[m.selected])
	} else {
		// Left pane (unselected)
		if len(m.visible) == 0 || m.selected < 0 || m.selected >= len(m.visible) {
			return m.noDetails()
		}
		return m.detailsForKey(m.visible[m.selected])
	}
}

// detailsForKey returns the details lines for a given manifest key
func (m *model) detailsForKey(key string) []string {
	entry := m.manifest[key]
	logical := []string{
		styles.HeaderStyle.Render("Details"),
		styles.DetailKey.Render("Name: ") + styles.DetailVal.Render(entry.Name),
		styles.DetailKey.Render("Key: ") + styles.DetailVal.Render(key),
		styles.DetailKey.Render("Desc: ") + styles.DetailVal.Render(entry.Desc),
	}
	if len(entry.Bin) > 0 {
		logical = append(logical, styles.DetailKey.Render("Bin: ")+styles.DetailVal.Render(strings.Join(entry.Bin, ", ")))
	}
	if len(entry.Brew) > 0 {
		logical = append(logical, styles.DetailKey.Render("Brew: ")+styles.DetailVal.Render(strings.Join(entry.Brew, ", ")))
	}
	if len(entry.Apt) > 0 {
		logical = append(logical, styles.DetailKey.Render("Apt: ")+styles.DetailVal.Render(strings.Join(entry.Apt, ", ")))
	}
	if len(entry.Pacman) > 0 {
		logical = append(logical, styles.DetailKey.Render("Pacman: ")+styles.DetailVal.Render(strings.Join(entry.Pacman, ", ")))
	}
	if entry.Docs != "" {
		logical = append(logical, styles.DetailKey.Render("Docs: ")+styles.DetailVal.Render(entry.Docs))
	}
	if entry.Github != "" {
		logical = append(logical, styles.DetailKey.Render("GitHub: ")+styles.DetailVal.Render(entry.Github))
	}
	if entry.Home != "" {
		logical = append(logical, styles.DetailKey.Render("Home: ")+styles.DetailVal.Render(entry.Home))
	}
	// Flatten to terminal lines
	var lines []string
	wrapWidth := panelWidth - 10
	for _, l := range logical {
		wrapped := wrap(l, wrapWidth)
		lines = append(lines, strings.Split(wrapped, "\n")...)
	}
	return lines
}

// noDetails returns the placeholder lines for when no details are available
func (m *model) noDetails() []string {
	return []string{
		styles.HeaderStyle.Render("Details"),
		styles.ItemStyle.Render("No details available."),
	}
}

// renderHeader renders the TUI header.
//
// # Returns
//   - string: The rendered header.
func (m *model) renderHeader() string {
	return styles.HeaderStyle.Render("chezmoi-a-la-carte 🛒") + "\n"
}

// renderSearch renders the search bar or prompt.
//
// # Returns
//   - string: The rendered search bar.
func (m *model) renderSearch() string {
	if m.searching {
		return styles.SearchStyle.Render(fmt.Sprintf("Search: %s_\n", m.search))
	}
	return styles.FooterStyle.Render("Search: (press / to search)") + "\n"
}

// renderList renders the list of visible entries (left pane)
func (m *model) renderList() string {
	var list strings.Builder
	listHeight := listHeight
	paneWidth := (panelWidth - 4) / 2
	start := m.selected - listHeight/2
	if start < 0 {
		start = 0
	}
	end := start + listHeight
	if end > len(m.visible) {
		end = len(m.visible)
	}

	if len(m.visible) == 0 {
		for i := 0; i < listHeight; i++ {
			line := ""
			if i == listHeight/2 {
				msg := styles.ItemStyle.Render("No results found. Press / to search, q to quit.")
				pad := (paneWidth - 8 - len("No results found. Press / to search, q to quit.")) / 2
				line = strings.Repeat(" ", pad) + msg
			}
			list.WriteString(lipgloss.NewStyle().Width(paneWidth).Render(line) + "\n")
		}
		return list.String()
	}

	linesRendered := 0
	for i := start; i < end; i++ {
		k := m.visible[i]
		entry := m.manifest[k]
		emoji := emojiForEntry(&entry)
		prefix := "  "
		line := fmt.Sprintf("%s %-20s %s", emoji, k, entry.Name)
		line = wrap(line, paneWidth-8)
		switch {
		case m.softwarePaneLeft && i == m.selected && m.focus == focusSoftware:
			line = styles.FocusStyle.Render(styles.SelectedItemStyle.Render(prefix + line))
		case i == m.selected:
			line = styles.SelectedItemStyle.Render(prefix + line)
		default:
			line = styles.ItemStyle.Render(prefix + line)
		}
		list.WriteString(lipgloss.NewStyle().Width(paneWidth).Render(line) + "\n")
		linesRendered++
	}
	for ; linesRendered < listHeight; linesRendered++ {
		list.WriteString(lipgloss.NewStyle().Width(paneWidth).Render("") + "\n")
	}
	return list.String()
}

// renderScrollIndicator returns a scroll indicator string based on scroll position.
//
// # Parameters
//   - scroll:    Current scroll offset.
//   - maxScroll: Maximum scroll offset.
//
// # Returns
//   - string: The scroll indicator (e.g., "↑", "↓", "↑↓", or "").
func (m *model) renderScrollIndicator(scroll, maxScroll int) string {
	switch {
	case scroll > 0 && scroll < maxScroll:
		return "↑↓"
	case scroll > 0:
		return "↑"
	case scroll < maxScroll:
		return "↓"
	default:
		return ""
	}
}

// renderDetailLines renders the details panel lines with optional scroll indicator.
//
// # Parameters
//   - detailLines: Lines to display.
//   - scroll:      Current scroll offset.
//   - maxScroll:   Maximum scroll offset.
//   - focused:     Whether the details panel is focused.
//
// # Returns
//   - string: The rendered details panel.
func (m *model) renderDetailLines(detailLines []string, scroll, maxScroll int, focused bool) string {
	var details strings.Builder
	dh := detailHeight

	if focused {
		indicator := m.renderScrollIndicator(scroll, maxScroll)
		for i := 0; i < dh; i++ {
			idx := scroll + i
			if idx < len(detailLines) {
				details.WriteString(detailLines[idx] + "\n")
			} else {
				details.WriteString("\n")
			}
		}
		if indicator != "" {
			lines := strings.Split(details.String(), "\n")
			if len(lines) > 1 {
				lines[len(lines)-2] = lines[len(lines)-2] + "  " + styles.FocusStyle.Render(indicator)
			}
			details.Reset()
			details.WriteString(strings.Join(lines, "\n"))
		}
	} else {
		for i := 0; i < dh; i++ {
			if i < len(detailLines) {
				details.WriteString(detailLines[i] + "\n")
			} else {
				details.WriteString("\n")
			}
		}
		if maxScroll > 0 {
			lines := strings.Split(details.String(), "\n")
			if len(lines) > 1 {
				lines[len(lines)-2] = lines[len(lines)-2] + "  " + styles.FocusStyle.Render("▼ more...")
			}
			details.Reset()
			details.WriteString(strings.Join(lines, "\n"))
		} else {
			details.WriteString("\n")
		}
	}
	return details.String()
}

// renderDetails renders the details panel for the selected entry.
//
// # Returns
//   - string: The rendered details panel.
func (m *model) renderDetails() string {
	detailLines := m.detailLines()
	maxScroll := len(detailLines) - detailHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := m.detailScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	return m.renderDetailLines(detailLines, scroll, maxScroll, m.focus == focusDetails)
}

// renderSelectedList renders the list of selected software in the right pane.
func (m *model) renderSelectedList() string {
	var list strings.Builder
	listHeight := listHeight
	paneWidth := (panelWidth - 4) - ((panelWidth - 4) / 2)
	start := m.selected - listHeight/2
	if start < 0 {
		start = 0
	}
	end := start + listHeight
	if end > len(m.selectedKeys) {
		end = len(m.selectedKeys)
	}

	if len(m.selectedKeys) == 0 {
		for i := 0; i < listHeight; i++ {
			line := ""
			if i == listHeight/2 {
				msg := styles.ItemStyle.Render("No software selected. Use ←/Enter to remove.")
				pad := (paneWidth - 8 - len("No software selected. Use ←/Enter to remove.")) / 2
				line = strings.Repeat(" ", pad) + msg
			}
			list.WriteString(lipgloss.NewStyle().Width(paneWidth).Render(line) + "\n")
		}
		return list.String()
	}

	linesRendered := 0
	for i := start; i < end; i++ {
		k := m.selectedKeys[i]
		entry := m.manifest[k]
		emoji := emojiForEntry(&entry)
		prefix := "  "
		line := fmt.Sprintf("%s %-20s %s", emoji, k, entry.Name)
		line = wrap(line, paneWidth-8)
		switch {
		case !m.softwarePaneLeft && i == m.selected && m.focus == focusSoftware:
			line = styles.FocusStyle.Render(styles.SelectedItemStyle.Render(prefix + line))
		case i == m.selected:
			line = styles.SelectedItemStyle.Render(prefix + line)
		default:
			line = styles.ItemStyle.Render(prefix + line)
		}
		list.WriteString(lipgloss.NewStyle().Width(paneWidth).Render(line) + "\n")
		linesRendered++
	}
	for ; linesRendered < listHeight; linesRendered++ {
		list.WriteString(lipgloss.NewStyle().Width(paneWidth).Render("") + "\n")
	}
	return list.String()
}

// View renders the entire TUI, including header, search, both panes, and footer.
//
// # Returns
//   - string: The full rendered TUI.
func (m *model) View() string {
	if m.loadErr != nil {
		return styles.BorderStyle.Render(styles.HeaderStyle.Render(fmt.Sprintf("Error loading manifest: %v\nPress q or Ctrl+C to quit.", m.loadErr)))
	}

	header := m.renderHeader()
	search := m.renderSearch()
	listContent := m.renderList()
	selectedContent := m.renderSelectedList()

	// Panel styles for focus (no border)
	listPanelStyle := styles.ListPanel
	selectedPanelStyle := styles.ListPanel
	paneHeight := listHeight
	if m.focus == focusSoftware {
		if m.softwarePaneLeft {
			listPanelStyle = styles.ListPanel.Background(DefaultTheme.FocusBg).Foreground(lipgloss.Color("51")).Bold(true)
		} else {
			selectedPanelStyle = styles.ListPanel.Background(DefaultTheme.FocusBg).Foreground(lipgloss.Color("51")).Bold(true)
		}
	}

	// Render left and right panes
	leftPane := listPanelStyle.Width((panelWidth - 4) / 2).Render(listContent)
	rightPane := selectedPanelStyle.Width((panelWidth - 4) - (panelWidth-4)/2).Render(selectedContent)
	if m.focus == focusSoftware && m.softwarePaneLeft {
		leftPane = lipgloss.NewStyle().Background(DefaultTheme.FocusBg).Height(paneHeight).Render(leftPane)
	} else {
		leftPane = lipgloss.NewStyle().Height(paneHeight).Render(leftPane)
	}
	if m.focus == focusSoftware && !m.softwarePaneLeft {
		rightPane = lipgloss.NewStyle().Background(DefaultTheme.FocusBg).Height(paneHeight).Render(rightPane)
	} else {
		rightPane = lipgloss.NewStyle().Height(paneHeight).Render(rightPane)
	}

	// Build a true vertical separator that fills the pane height
	separatorStyle := lipgloss.NewStyle().
		Foreground(DefaultTheme.Border).
		Width(1)
	separatorChar := "│"

	leftLines := strings.Split(leftPane, "\n")
	rightLines := strings.Split(rightPane, "\n")
	maxLines := paneHeight
	if len(leftLines) > maxLines {
		maxLines = len(leftLines)
	}
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}
	// Pad lines to maxLines
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, strings.Repeat(" ", (panelWidth-4)/2))
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, strings.Repeat(" ", (panelWidth-4)-(panelWidth-4)/2))
	}
	// Build the main panes row by row
	mainPaneRows := make([]string, maxLines)
	for i := 0; i < maxLines; i++ {
		mainPaneRows[i] = leftLines[i] + separatorStyle.Render(separatorChar) + rightLines[i]
	}
	mainPanes := strings.Join(mainPaneRows, "\n")

	detailPanelStyle := styles.DetailPanel
	if m.focus == focusDetails {
		detailPanelStyle = styles.DetailPanel.BorderForeground(lipgloss.Color("51")).Bold(true)
	}

	footer := styles.FooterStyle.Render(
		"↑/↓/j/k: Move  Enter: Select/Deselect  →/←: Switch pane  TAB: Toggle details  /: Search  q: Quit  esc: Cancel search",
	)

	// Compose the full UI with an outer border
	mainPanel := lipgloss.JoinVertical(lipgloss.Left,
		header,
		search,
		mainPanes,
		detailPanelStyle.Render(m.renderDetails()),
		footer,
	)
	return styles.BorderStyle.Render(mainPanel)
}

// main is the entry point for the chezmoi-a-la-carte TUI application.
//
// # Steps
//
//  1. Loads environment variables from .env if present.
//  2. Loads the software manifest from SOFTWARE_MANIFEST_PATH or software.yml.
//  3. Initializes the TUI model and starts the Bubble Tea program.
//
// # Example
//
//	$ go run ./cmd/chezmoi-a-la-carte
func main() {
	// Load .env if present
	_ = godotenv.Load()
	// Load manifest
	manifestPath := os.Getenv("SOFTWARE_MANIFEST_PATH")
	if manifestPath == "" {
		manifestPath = "software.yml"
	}
	manifest, err := app.LoadManifest(manifestPath)
	if err != nil {
		fmt.Printf("Error loading manifest: %v\n", err)
		os.Exit(1)
	}
	m := &model{
		manifest:         manifest,
		entries:          make([]string, 0, len(manifest)),
		visible:          make([]string, 0, len(manifest)),
		selectedKeys:     make([]string, 0, len(manifest)),
		focus:            focusSoftware,
		softwarePaneLeft: true,
	}
	for k := range manifest {
		m.entries = append(m.entries, k)
	}
	sort.Strings(m.entries)
	m.visible = m.entries
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// moveToSelected moves the currently highlighted entry from visible to selectedKeys
func (m *model) moveToSelected() {
	if m.softwarePaneLeft && len(m.visible) > 0 {
		key := m.visible[m.selected]
		if !m.isSelected(key) {
			m.selectedKeys = append(m.selectedKeys, key)
			m.filter()
			if m.selected >= len(m.visible) {
				m.selected = len(m.visible) - 1
			}
			if m.selected < 0 {
				m.selected = 0
			}
			// If right pane is not focused, reset its index to 0 so it always shows the top
			if !m.softwarePaneLeft && len(m.selectedKeys) > 0 {
				m.selected = 0
			}
		}
	}
}

// moveToDeselected moves the currently highlighted entry from selectedKeys back to visible
func (m *model) moveToDeselected() {
	if !m.softwarePaneLeft && len(m.selectedKeys) > 0 {
		idx := m.selected
		if idx >= 0 && idx < len(m.selectedKeys) {
			key := m.selectedKeys[idx]
			m.selectedKeys = append(m.selectedKeys[:idx], m.selectedKeys[idx+1:]...)
			m.filter()
			if m.selected >= len(m.selectedKeys) {
				m.selected = len(m.selectedKeys) - 1
			}
			if m.selected < 0 {
				m.selected = 0
			}
			// Highlight the newly restored item in the left pane
			if len(m.visible) > 0 {
				for i, v := range m.visible {
					if v == key {
						m.selected = i
						break
					}
				}
			}
		}
	}
}
