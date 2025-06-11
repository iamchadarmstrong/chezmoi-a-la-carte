package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"a-la-carte/internal/ui/core" // Updated from internal/ui
)

// SearchBarModel represents the search bar component
type SearchBarModel struct {
	search    string
	searching bool
	width     int
}

// NewSearchBarModel creates a new search bar model
func NewSearchBarModel() *SearchBarModel {
	return &SearchBarModel{
		search:    "",
		searching: false,
	}
}

// Init initializes the search bar model
func (s *SearchBarModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the search bar
func (s *SearchBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		key := keyMsg.String()
		if s.searching {
			switch key {
			case "enter", "tab", "esc":
				// Lock in search state when user navigates away, but preserve text
				s.searching = false
				return s, nil
			case "backspace":
				if s.search != "" {
					s.search = s.search[:len(s.search)-1]
				}
				return s, nil
			default:
				if len(key) == 1 && key >= " " && key <= "~" {
					s.search += key
					return s, nil
				}
			}
		} else if key == "/" {
			s.searching = true
			return s, nil
		}
	}
	return s, nil
}

// View renders the search bar
func (s *SearchBarModel) View() string {
	// Get current theme
	t := core.CurrentTheme() // Updated from ui.CurrentTheme()

	// Create base style for the search bar
	searchBarStyle := lipgloss.NewStyle().
		Foreground(t.Text()).
		Background(t.Background()).
		Padding(0, 1)
	// Apply width if it's greater than 0
	if s.width > 0 {
		searchBarStyle = searchBarStyle.Width(s.width)
	}

	// Create style for the search label
	searchLabelStyle := lipgloss.NewStyle().
		Foreground(t.Accent()).
		Background(t.Background()).
		Bold(true)

	// Create style for the search input
	searchInputStyle := lipgloss.NewStyle().
		Foreground(t.Text()).
		Background(t.Background())

	// Create style for the placeholder
	placeholderStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Background(t.Background()).
		Italic(true)

	if s.searching {
		// When in focus, show cursor and current input
		return searchBarStyle.Render(
			searchLabelStyle.Render("Search: ") +
				searchInputStyle.Render(s.search+"_"),
		)
	}

	// When not in focus
	if s.search == "" {
		// If no search input, show placeholder
		return searchBarStyle.Render(
			searchLabelStyle.Render("Search: ") +
				placeholderStyle.Render("(press / to search)"),
		)
	} else {
		// If has search input, show it without cursor
		return searchBarStyle.Render(
			searchLabelStyle.Render("Search: ") +
				searchInputStyle.Render(s.search),
		)
	}
}

// SetWidth sets the width of the search bar.
func (s *SearchBarModel) SetWidth(width int) {
	s.width = width
}

// GetSearch returns the current search query
func (s *SearchBarModel) GetSearch() string {
	return s.search
}

// IsSearching returns whether the search bar is active
func (s *SearchBarModel) IsSearching() bool {
	return s.searching
}

// ResetSearch resets the search state
func (s *SearchBarModel) ResetSearch() {
	s.search = ""
	s.searching = false
}
