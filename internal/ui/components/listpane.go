package components

import (
	"a-la-carte/internal/app"
	"a-la-carte/internal/ui/core" // Updated from internal/ui

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListItem struct {
	Key      string
	Name     string
	Emoji    string
	Content  string // formatted string for display
	Selected bool
	Focused  bool
	Style    lipgloss.Style
}

type ListPaneModel struct {
	items    []ListItem
	emptyMsg string
	width    int
	height   int
}

func NewListPaneModel(items []ListItem, emptyMsg string) *ListPaneModel {
	return &ListPaneModel{
		items:    items,
		emptyMsg: emptyMsg,
	}
}

func (l *ListPaneModel) Init() tea.Cmd                           { return nil }
func (l *ListPaneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return l, nil }
func (l *ListPaneModel) SetSize(width, height int) {
	l.width = width
	l.height = height
}
func (l *ListPaneModel) View() string {
	if len(l.items) == 0 {
		return lipgloss.Place(
			l.width,
			l.height,
			lipgloss.Center,
			lipgloss.Center,
			l.emptyMsg,
		)
	}
	return renderPaneList(l.items, l.width, l.height, l.emptyMsg)
}

// Helper to prepare list items for a pane
func PreparePaneListItems(keys []string, manifest app.Manifest, selectedIdx int, focused bool, paneWidth int, isLeftPane bool, currentFocus int) []ListItem {
	items := make([]ListItem, len(keys))
	for i, k := range keys {
		e := manifest[k]
		emoji := core.EmojiForEntry(&e) // Updated from ui.EmojiForEntry
		item := ListItem{
			Key:      k,
			Name:     e.Name,
			Emoji:    emoji,
			Selected: i == selectedIdx,
			Focused:  focused,
		}
		item.Content = item.Emoji + " " + item.Name
		// Style logic can be expanded as needed
		if item.Selected && item.Focused {
			item.Style = core.CurrentStyles().SelectedItemStyle // Updated from ui.CurrentStyles()
		} else {
			// Assuming GetItemStyle was a helper, directly use ItemStyle or ActiveItemStyle based on focus
			if focused {
				item.Style = core.CurrentStyles().ActiveItemStyle // Or some other style for focused but not selected
			} else {
				item.Style = core.CurrentStyles().ItemStyle
			}
		}
		items[i] = item
	}
	return items
}

// Helper to render the list pane
func renderPaneList(items []ListItem, paneWidth, listHeight int, emptyMsg string) string {
	if len(items) == 0 {
		return lipgloss.Place(
			paneWidth,
			listHeight,
			lipgloss.Center,
			lipgloss.Center,
			emptyMsg,
		)
	}
	lines := make([]string, len(items))
	for i := range items {
		lines[i] = items[i].Style.Width(paneWidth).Render(items[i].Content)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.Place(
		paneWidth,
		listHeight,
		lipgloss.Left,
		lipgloss.Top,
		content,
	)
}
