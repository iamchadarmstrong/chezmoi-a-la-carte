package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lexnux/a-la-carte/internal/app"
	"github.com/mattn/go-runewidth"
)

const (
	panelWidth         = 80
	listHeight         = 12
	detailHeight       = 7
	detailHeightExpand = 16
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Padding(0, 1).Width(panelWidth - 4)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(true).Width(panelWidth - 8)
	itemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1).Width(panelWidth - 8)
	detailKey     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	detailVal     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	searchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	borderStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2).Width(panelWidth)
	detailPanel   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("60")).Padding(0, 2).Margin(1, 0).Width(panelWidth - 6)
	footerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1).Width(panelWidth - 4)
	focusStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
)

type focusArea int

const (
	focusList focusArea = iota
	focusDetails
)

// model defines the state of the TUI
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
		key := msg.String()
		if key == "q" {
			return m, tea.Quit
		}
		if m.loadErr != nil {
			return m, nil
		}
		if m.searching {
			if key == "enter" {
				m.searching = false
				return m, nil
			}
			if key == "esc" {
				m.searching = false
				m.search = ""
				m.filter()
				return m, nil
			}
			if key == "backspace" && len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
				m.filter()
				return m, nil
			}
			if len(key) == 1 && key >= " " && key <= "~" {
				m.search += key
				m.filter()
				return m, nil
			}
			return m, nil
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
			switch key {
			case "/":
				m.searching = true
				return m, nil
			case "up", "k":
				if m.selected > 0 {
					m.selected--
				}
				return m, nil
			case "down", "j":
				if m.selected < len(m.visible)-1 {
					m.selected++
				}
				return m, nil
			}
		} else if m.focus == focusDetails {
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
				return m, nil
			case "down", "j":
				if m.detailScroll < maxScroll {
					m.detailScroll++
				}
				return m, nil
			}
		}
	}
	return m, nil
}

// normalizeEmoji ensures the emoji is exactly 2 columns wide for consistent alignment
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

// Only use emojis that are reliably 2 columns wide in most terminals.
// Safe emoji set: 🐍, 🟩, 🐹, 🐳, 🌱, 🐧, 🍏, 🍺, 💻, 🧪, 📄, 🔑, 🔄, 📝, 📦, 🧰
// Avoid: 🛠️, 🗂️, ☁️, 🍎, etc.
func emojiForEntry(e app.SoftwareEntry) string {
	n := strings.ToLower(e.Name)
	d := strings.ToLower(e.Desc)
	switch {
	case strings.Contains(n, "python") || strings.Contains(d, "python"):
		return normalizeEmoji("🐍")
	case strings.Contains(n, "node") || strings.Contains(d, "node.js"):
		return normalizeEmoji("🟩")
	case strings.Contains(n, "go") || strings.Contains(d, "golang"):
		return normalizeEmoji("🐹")
	case strings.Contains(n, "docker"):
		return normalizeEmoji("🐳")
	case strings.Contains(n, "git"):
		return normalizeEmoji("🌱")
	case strings.Contains(n, "linux") || strings.Contains(d, "linux"):
		return normalizeEmoji("🐧")
	case strings.Contains(n, "mac") || strings.Contains(d, "macos"):
		return normalizeEmoji("🍏") // green apple, reliably 2 columns
	case strings.Contains(n, "brew") || strings.Contains(d, "homebrew"):
		return normalizeEmoji("🍺")
	case strings.Contains(n, "cli") || strings.Contains(d, "command-line"):
		return normalizeEmoji("💻")
	case strings.Contains(n, "test") || strings.Contains(d, "test"):
		return normalizeEmoji("🧪")
	case strings.Contains(n, "file") || strings.Contains(d, "file"):
		return normalizeEmoji("📄")
	case strings.Contains(n, "ssh"):
		return normalizeEmoji("🔑")
	case strings.Contains(n, "cloud") || strings.Contains(d, "cloud"):
		return normalizeEmoji("💻") // fallback to laptop for cloud
	case strings.Contains(n, "sync") || strings.Contains(d, "sync"):
		return normalizeEmoji("🔄")
	case strings.Contains(n, "markdown"):
		return normalizeEmoji("📝")
	case strings.Contains(n, "package") || strings.Contains(d, "package"):
		return normalizeEmoji("📦")
	case strings.Contains(n, "util") || strings.Contains(d, "utility"):
		return normalizeEmoji("🧰")
	default:
		return normalizeEmoji("💻") // fallback to laptop for all others
	}
}

func wrap(s string, width int) string {
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(s)
}

func (m model) detailLines() []string {
	if len(m.visible) == 0 {
		return []string{
			headerStyle.Render("Details"),
			itemStyle.Render("No details available."),
		}
	}
	selKey := m.visible[m.selected]
	sel := m.manifest[selKey]
	var logical []string
	logical = append(logical, headerStyle.Render("Details"))
	logical = append(logical, detailKey.Render("Name: ")+detailVal.Render(sel.Name))
	logical = append(logical, detailKey.Render("Key: ")+detailVal.Render(selKey))
	logical = append(logical, detailKey.Render("Desc: ")+detailVal.Render(sel.Desc))
	if len(sel.Bin) > 0 {
		logical = append(logical, detailKey.Render("Bin: ")+detailVal.Render(strings.Join(sel.Bin, ", ")))
	}
	if len(sel.Brew) > 0 {
		logical = append(logical, detailKey.Render("Brew: ")+detailVal.Render(strings.Join(sel.Brew, ", ")))
	}
	if len(sel.Apt) > 0 {
		logical = append(logical, detailKey.Render("Apt: ")+detailVal.Render(strings.Join(sel.Apt, ", ")))
	}
	if len(sel.Pacman) > 0 {
		logical = append(logical, detailKey.Render("Pacman: ")+detailVal.Render(strings.Join(sel.Pacman, ", ")))
	}
	if sel.Docs != "" {
		logical = append(logical, detailKey.Render("Docs: ")+detailVal.Render(sel.Docs))
	}
	if sel.Github != "" {
		logical = append(logical, detailKey.Render("GitHub: ")+detailVal.Render(sel.Github))
	}
	if sel.Home != "" {
		logical = append(logical, detailKey.Render("Home: ")+detailVal.Render(sel.Home))
	}
	// Flatten to terminal lines
	var lines []string
	wrapWidth := panelWidth - 10
	for _, l := range logical {
		wrapped := wrap(l, wrapWidth)
		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, line)
		}
	}
	return lines
}

func (m model) View() string {
	if m.loadErr != nil {
		return borderStyle.Render(headerStyle.Render(fmt.Sprintf("Error loading manifest: %v\nPress q or Ctrl+C to quit.", m.loadErr)))
	}
	// Always render the full UI: header, search bar, list area, details, footer
	var list strings.Builder
	list.WriteString(headerStyle.Render("chezmoi-a-la-carte 🛒") + "\n")
	if m.searching {
		list.WriteString(searchStyle.Render(fmt.Sprintf("Search: %s_\n", m.search)))
	} else {
		list.WriteString(footerStyle.Render("Search: (press / to search)") + "\n")
	}
	max := listHeight
	start := m.selected - max/2
	if start < 0 {
		start = 0
	}
	end := start + max
	if end > len(m.visible) {
		end = len(m.visible)
	}
	linesRendered := 0
	if len(m.visible) == 0 {
		// Render the list area as a fixed-height block, with a centered message and padding
		for i := 0; i < listHeight; i++ {
			if i == listHeight/2 {
				msg := itemStyle.Render("No results found. Press / to search, q to quit.")
				pad := (panelWidth - 8 - len("No results found. Press / to search, q to quit.")) / 2
				list.WriteString(strings.Repeat(" ", pad) + msg + "\n")
			} else {
				list.WriteString("\n")
			}
		}
		linesRendered = listHeight
	} else {
		for i := start; i < end; i++ {
			k := m.visible[i]
			entry := m.manifest[k]
			emoji := emojiForEntry(entry)
			prefix := "  "
			line := fmt.Sprintf("%s %-20s %s", emoji, k, entry.Name)
			line = wrap(line, panelWidth-8)
			if i == m.selected && m.focus == focusList {
				list.WriteString(focusStyle.Render(selectedStyle.Render(prefix+line)) + "\n")
			} else if i == m.selected {
				list.WriteString(selectedStyle.Render(prefix+line) + "\n")
			} else {
				list.WriteString(itemStyle.Render(prefix+line) + "\n")
			}
			linesRendered++
		}
		// Always pad with newlines to reach listHeight
		for ; linesRendered < listHeight; linesRendered++ {
			list.WriteString("\n")
		}
	}
	// Details panel and footer always rendered
	detailLines := m.detailLines()
	dh := detailHeight
	var details strings.Builder
	maxScroll := len(detailLines) - dh
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
	indicator := ""
	if m.focus == focusDetails {
		if maxScroll > 0 {
			if scroll > 0 && scroll < maxScroll {
				indicator = "↑↓"
			} else if scroll > 0 {
				indicator = "↑"
			} else if scroll < maxScroll {
				indicator = "↓"
			}
		}
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
				lines[len(lines)-2] = lines[len(lines)-2] + "  " + focusStyle.Render(indicator)
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
				lines[len(lines)-2] = lines[len(lines)-2] + "  " + focusStyle.Render("▼ more...")
			}
			details.Reset()
			details.WriteString(strings.Join(lines, "\n"))
		} else {
			details.WriteString("\n")
		}
	}
	detailPanelStyle := detailPanel
	if m.focus == focusDetails {
		detailPanelStyle = detailPanel.Copy().BorderForeground(lipgloss.Color("51")).Bold(true)
	}

	footer := footerStyle.Render("↑/↓/j/k: Move  /: Search  q: Quit  Enter: Details  esc: Cancel search  TAB: Toggle focus")

	mainPanel := lipgloss.JoinVertical(lipgloss.Left,
		list.String(),
		detailPanelStyle.Render(details.String()),
		footer,
	)
	return borderStyle.Render(mainPanel)
}

func main() {
	manifest, err := app.LoadManifest("software.yml")
	// Sort keys for consistent order
	var keys []string
	for k := range manifest {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	m := model{
		manifest: manifest,
		loadErr:  err,
		entries:  keys,
		visible:  keys,
		selected: 0,
	}
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
