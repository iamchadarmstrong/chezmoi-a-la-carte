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
	panelWidth         = 80
	listHeight         = 12
	detailHeight       = 7
	detailHeightExpand = 16
)

type focusArea int

const (
	focusList focusArea = iota
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
//   - focus:        Which panel is focused (list or details).
//   - detailScroll: Scroll offset for the details panel.
//
// # Methods
//
//   - filter():         Updates visible entries based on search.
//   - handleSearchInput(): Handles key input in search mode.
//   - handleListInput():   Handles key input in list mode.
//   - handleDetailsInput(): Handles key input in details mode.
//   - detailLines():    Returns the lines to display in the details panel.
//   - View():           Renders the full TUI.
type model struct {
	manifest     app.Manifest
	loadErr      error
	entries      []string // sorted keys
	visible      []string // filtered keys
	selected     int
	search       string
	searching    bool
	focus        focusArea
	detailScroll int
}

func (m *model) filter() {
	if m.search == "" {
		m.visible = m.entries
		return
	}
	var filtered []string
	q := strings.ToLower(m.search)
	for _, k := range m.entries {
		e := m.manifest[k]
		if strings.Contains(strings.ToLower(k), q) || strings.Contains(strings.ToLower(e.Name), q) || strings.Contains(strings.ToLower(e.Desc), q) {
			filtered = append(filtered, k)
		}
	}
	m.visible = filtered
	if m.selected >= len(m.visible) {
		m.selected = len(m.visible) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
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

func (m *model) handleListInput(key string) *model {
	switch key {
	case "/":
		m.searching = true
		return m
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		return m
	case "down", "j":
		if m.selected < len(m.visible)-1 {
			m.selected++
		}
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
			if m.focus == focusList {
				m.focus = focusDetails
				m.detailScroll = 0
			} else {
				m.focus = focusList
			}
			return m, nil
		}
		if m.focus == focusList {
			return m.handleListInput(key), nil
		} else if m.focus == focusDetails {
			return m.handleDetailsInput(key), nil
		}
	}
	return m, nil
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
	if len(m.visible) == 0 {
		return []string{
			styles.HeaderStyle.Render("Details"),
			styles.ItemStyle.Render("No details available."),
		}
	}
	selKey := m.visible[m.selected]
	sel := m.manifest[selKey]
	logical := []string{
		styles.HeaderStyle.Render("Details"),
		styles.DetailKey.Render("Name: ") + styles.DetailVal.Render(sel.Name),
		styles.DetailKey.Render("Key: ") + styles.DetailVal.Render(selKey),
		styles.DetailKey.Render("Desc: ") + styles.DetailVal.Render(sel.Desc),
	}
	if len(sel.Bin) > 0 {
		logical = append(logical, styles.DetailKey.Render("Bin: ")+styles.DetailVal.Render(strings.Join(sel.Bin, ", ")))
	}
	if len(sel.Brew) > 0 {
		logical = append(logical, styles.DetailKey.Render("Brew: ")+styles.DetailVal.Render(strings.Join(sel.Brew, ", ")))
	}
	if len(sel.Apt) > 0 {
		logical = append(logical, styles.DetailKey.Render("Apt: ")+styles.DetailVal.Render(strings.Join(sel.Apt, ", ")))
	}
	if len(sel.Pacman) > 0 {
		logical = append(logical, styles.DetailKey.Render("Pacman: ")+styles.DetailVal.Render(strings.Join(sel.Pacman, ", ")))
	}
	if sel.Docs != "" {
		logical = append(logical, styles.DetailKey.Render("Docs: ")+styles.DetailVal.Render(sel.Docs))
	}
	if sel.Github != "" {
		logical = append(logical, styles.DetailKey.Render("GitHub: ")+styles.DetailVal.Render(sel.Github))
	}
	if sel.Home != "" {
		logical = append(logical, styles.DetailKey.Render("Home: ")+styles.DetailVal.Render(sel.Home))
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

// renderEmptyList renders a placeholder when no results are found.
//
// # Returns
//   - string: The rendered empty list message.
func (m *model) renderEmptyList() string {
	var list strings.Builder
	for i := 0; i < listHeight; i++ {
		if i == listHeight/2 {
			msg := styles.ItemStyle.Render("No results found. Press / to search, q to quit.")
			pad := (panelWidth - 8 - len("No results found. Press / to search, q to quit.")) / 2
			list.WriteString(strings.Repeat(" ", pad) + msg + "\n")
		} else {
			list.WriteString("\n")
		}
	}
	return list.String()
}

// renderList renders the list of visible entries.
//
// # Returns
//   - string: The rendered list.
func (m *model) renderList() string {
	var list strings.Builder
	listHeight := listHeight // rename to avoid shadowing
	start := m.selected - listHeight/2
	if start < 0 {
		start = 0
	}
	end := start + listHeight
	if end > len(m.visible) {
		end = len(m.visible)
	}

	if len(m.visible) == 0 {
		return m.renderEmptyList()
	}

	linesRendered := 0
	for i := start; i < end; i++ {
		k := m.visible[i]
		entry := m.manifest[k]
		emoji := emojiForEntry(&entry)
		prefix := "  "
		line := fmt.Sprintf("%s %-20s %s", emoji, k, entry.Name)
		line = wrap(line, panelWidth-8)
		switch {
		case i == m.selected && m.focus == focusList:
			list.WriteString(styles.FocusStyle.Render(styles.SelectedItemStyle.Render(prefix+line)) + "\n")
		case i == m.selected:
			list.WriteString(styles.SelectedItemStyle.Render(prefix+line) + "\n")
		default:
			list.WriteString(styles.ItemStyle.Render(prefix+line) + "\n")
		}
		linesRendered++
	}
	// Always pad with newlines to reach listHeight
	for ; linesRendered < listHeight; linesRendered++ {
		list.WriteString("\n")
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

// View renders the entire TUI, including header, search, list, details, and footer.
//
// # Returns
//   - string: The full rendered TUI.
func (m *model) View() string {
	if m.loadErr != nil {
		return styles.BorderStyle.Render(styles.HeaderStyle.Render(fmt.Sprintf("Error loading manifest: %v\nPress q or Ctrl+C to quit.", m.loadErr)))
	}

	// Render header and search bar outside the list container
	header := m.renderHeader()
	search := m.renderSearch()
	listContent := m.renderList()

	detailPanelStyle := styles.DetailPanel
	if m.focus == focusDetails {
		detailPanelStyle = styles.DetailPanel.BorderForeground(lipgloss.Color("51")).Bold(true)
	}

	listPanelStyle := styles.ListPanel
	if m.focus == focusList {
		listPanelStyle = styles.ListPanel.BorderForeground(lipgloss.Color("51")).Bold(true)
	}

	footer := styles.FooterStyle.Render("↑/↓/j/k: Move  /: Search  q: Quit  Enter: Details  esc: Cancel search  TAB: Toggle focus")

	mainPanel := lipgloss.JoinVertical(lipgloss.Left,
		header,
		search,
		listPanelStyle.Render(listContent),
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
		manifest: manifest,
		entries:  make([]string, 0, len(manifest)),
		visible:  make([]string, 0, len(manifest)),
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
