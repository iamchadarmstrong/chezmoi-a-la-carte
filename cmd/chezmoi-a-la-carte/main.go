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
	"a-la-carte/internal/config"
	"a-la-carte/internal/flags"
	"a-la-carte/internal/ui/components"
	"a-la-carte/internal/ui/core"
	"a-la-carte/internal/ui/patterns"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	panelWidth                  = core.PanelWidth         // Changed from ui.PanelWidth
	listHeight                  = core.ListHeight         // Changed from ui.ListHeight
	detailHeight                = core.DetailHeight       // Changed from ui.DetailHeight
	detailHeightExpand          = core.DetailHeightExpand // Changed from ui.DetailHeightExpand
	borderAndPadding            = core.BorderWidth        // Changed from ui.BorderWidth
	leftPaneContentWidth        = core.LeftPaneWidth      // Changed from ui.LeftPaneWidth
	rightPaneContentWidth       = core.RightPaneWidth     // Changed from ui.RightPaneWidth
	leftPaneTotalWidth          = leftPaneContentWidth + borderAndPadding
	rightPaneTotalWidth         = rightPaneContentWidth + borderAndPadding
	splitPaneTotalWidth         = leftPaneTotalWidth + rightPaneTotalWidth
	leftRatio                   = float64(leftPaneTotalWidth) / float64(splitPaneTotalWidth)
	splitRatio                  = core.SplitPaneRatio            // Changed from ui.SplitPaneRatio
	cardPadding                 = 1                              // Based on patterns.Card using WithPaddingAll(1)
	cardBorder                  = 1                              // Based on patterns.Card using WithBorderAll()
	cardTotalHorizontalOverhead = (cardPadding + cardBorder) * 2 // For left and right sides
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
//   - uiActiveListIndex:     Index of the currently selected entry.
//   - searchBar:    The search bar model.
//   - focus:        Which panel is focused (list, details, or selected).
//   - detailScroll: Scroll offset for the details panel.
//   - selectedKeys: Keys of software selected for the right pane.
//   - softwarePaneLeft: Track which pane is active in software focus: true=left, false=right
//   - showHelp:     Whether to show the help overlay
//   - layout:       The layout for the TUI
//   - width, height: The window size
type model struct {
	manifest          app.Manifest
	loadErr           error
	entries           []string // sorted keys
	visible           []string // filtered keys (left pane, excludes selected)
	uiActiveListIndex int      // RENAME of 'selected int'. Index in visible (left) or selectedKeys (right)
	searchBar         *components.SearchBarModel
	focus             focusArea
	detailScroll      int

	selectedKeys []string // keys of selected software (right pane)
	// track which pane is active in software focus: true=left, false=right
	softwarePaneLeft bool
	showHelp         bool // whether to show the help overlay

	// Configuration
	config *config.Config

	// Layout
	topSplitPane      patterns.SplitPaneLayout
	width, height     int
	contentWidth      int
	detailsPanelModel tea.Model
}

// layoutMetrics is initialized in Init() to ensure all computed values are available // Changed variable name
var layoutMetrics *core.LayoutMetrics // Changed from ui.LayoutMetrics

// filterEntriesByQuery returns entries that match the given search query
func (m *model) filterEntriesByQuery(query string) []string {
	if query == "" {
		return m.entries
	}

	candidateKeys := []string{}
	lowerQuery := strings.ToLower(query)

	for _, key := range m.entries {
		entry := m.manifest[key]
		if strings.Contains(strings.ToLower(entry.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(key), lowerQuery) ||
			strings.Contains(strings.ToLower(entry.Desc), lowerQuery) {
			candidateKeys = append(candidateKeys, key)
		}
	}

	return candidateKeys
}

// excludeSelectedKeys filters out keys that are already in the selected list
func (m *model) excludeSelectedKeys(candidates []string) []string {
	selectedSet := make(map[string]struct{})
	for _, key := range m.selectedKeys {
		selectedSet[key] = struct{}{}
	}

	result := []string{}
	for _, key := range candidates {
		if _, found := selectedSet[key]; !found {
			result = append(result, key)
		}
	}

	return result
}

// clampActiveListIndex ensures the active index is within valid bounds
func (m *model) clampActiveListIndex() {
	if m.softwarePaneLeft {
		if m.uiActiveListIndex >= len(m.visible) {
			m.uiActiveListIndex = len(m.visible) - 1
		}
		if m.uiActiveListIndex < 0 && len(m.visible) > 0 {
			m.uiActiveListIndex = 0
		} else if len(m.visible) == 0 {
			m.uiActiveListIndex = 0 // Or -1, depending on how empty lists are handled
		}
	} else {
		// For right pane (selected keys)
		if m.uiActiveListIndex >= len(m.selectedKeys) {
			m.uiActiveListIndex = len(m.selectedKeys) - 1
		}
		if m.uiActiveListIndex < 0 && len(m.selectedKeys) > 0 {
			m.uiActiveListIndex = 0
		} else if len(m.selectedKeys) == 0 {
			m.uiActiveListIndex = 0
		}
	}
}

func (m *model) filter() {
	query := m.searchBar.GetSearch()
	candidateKeys := m.filterEntriesByQuery(query)
	m.visible = m.excludeSelectedKeys(candidateKeys)
	m.clampActiveListIndex()
}

func (m *model) Init() tea.Cmd {
	metrics := core.DefaultLayoutMetrics() // Get the value
	layoutMetrics = &metrics               // Assign its address

	m.topSplitPane = patterns.NewSplitPane(
		patterns.WithLeftPanel(patterns.Panel(core.EmptyModel())),
		patterns.WithRightPanel(patterns.Panel(core.EmptyModel())),
		patterns.WithRatio(core.SplitPaneRatio),
		// No WithBottomPanel or WithVerticalRatio here
	)
	m.searchBar = components.NewSearchBarModel()

	// Initialize detailsPanelModel
	initialDetailsData := components.DetailsPanelData{Lines: []string{"Initializing details..."}}
	// Use layoutMetrics for initial width, and detailHeight constant for height
	detailsModelWidth := layoutMetrics.PanelWidth // This is the full panel width
	if detailsModelWidth < 0 {
		detailsModelWidth = 0
	}
	detailsModelHeight := detailHeight // This is a line count
	if detailsModelHeight < 0 {
		detailsModelHeight = 0
	}
	m.detailsPanelModel = components.NewDetailsPanelModel(&initialDetailsData, detailsModelWidth, detailsModelHeight, false, 0, 0)

	var initCmds []tea.Cmd
	initCmds = append(initCmds, m.topSplitPane.Init())
	if m.detailsPanelModel != nil {
		initCmds = append(initCmds, m.detailsPanelModel.Init())
	}

	return tea.Batch(initCmds...)
}

func (m *model) handleDetailsInput(key string) *model {
	detailLines := m.detailLines(m.contentWidth) // Pass m.contentWidth
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

// handleHelpKey handles key input when help is shown
func (m *model) handleHelpKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "h":
		m.showHelp = false
		return m, nil
	case "q":
		return m, tea.Quit
	default:
		return m, nil
	}
}

// handleSearchKey handles key input when search is active
func (m *model) handleSearchKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	updatedSearchBar, searchCmd := m.searchBar.Update(msg)
	m.searchBar = updatedSearchBar.(*components.SearchBarModel)
	m.filter()
	return m, searchCmd
}

// handleGeneralKey handles general key input
func (m *model) handleGeneralKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m, tea.Quit
	case "h":
		m.showHelp = !m.showHelp
		return m, nil
	case "tab":
		return m.handleTab(), nil
	}

	if m.loadErr != nil {
		return m, nil
	}

	switch m.focus {
	case focusSoftware:
		return m.handleSoftwareKey(key), nil
	case focusDetails:
		return m.handleDetailsInput(key), nil
	}

	return m, nil
}

// handleWindowSize handles window size changes
func (m *model) handleWindowSize(win tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.width, m.height = win.Width, win.Height

	// Calculate available width for content inside the main card
	m.contentWidth = m.width - cardTotalHorizontalOverhead
	if m.contentWidth < 0 {
		m.contentWidth = 0
	}

	// Update searchBar width
	if m.searchBar != nil {
		m.searchBar.SetWidth(m.contentWidth)
	}

	// Update topSplitPane size
	if m.topSplitPane != nil {
		topSplitCtx := &core.LayoutContext{
			AvailableWidth:  m.contentWidth,
			AvailableHeight: listHeight,
			NestingLevel:    0,
		}
		updateCmd := m.topSplitPane.SetSize(m.contentWidth, listHeight, topSplitCtx)
		cmds = append(cmds, updateCmd)
	}

	// Update DetailsPanelModel's internal width/height
	if dpm, ok := m.detailsPanelModel.(*components.DetailsPanelModel); ok {
		dpm.SetDimensions(m.contentWidth, detailHeight)
	}
	return m, tea.Batch(cmds...)
}

// propagateUpdates propagates updates to child components
func (m *model) propagateUpdates(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Propagate update to topSplitPane
	if m.topSplitPane != nil {
		var topSplitCmd tea.Cmd
		updatedTopSplitPane, topSplitCmd := m.topSplitPane.Update(msg)
		if updatedTopSplit, ok := updatedTopSplitPane.(patterns.SplitPaneLayout); ok {
			m.topSplitPane = updatedTopSplit
		}
		cmds = append(cmds, topSplitCmd)
	}

	// Propagate update to detailsPanelModel
	if m.detailsPanelModel != nil {
		var detailsCmd tea.Cmd
		m.detailsPanelModel, detailsCmd = m.detailsPanelModel.Update(msg)
		cmds = append(cmds, detailsCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle help mode
	if m.showHelp && !m.searchBar.IsSearching() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleHelpKey(keyMsg.String())
		}
		return m, nil
	}

	// Handle search mode
	if m.searchBar.IsSearching() {
		return m.handleSearchKey(msg)
	}

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleGeneralKey(keyMsg.String())
	}

	// Handle window size changes
	if win, ok := msg.(tea.WindowSizeMsg); ok {
		return m.handleWindowSize(win)
	}

	// Propagate updates to child components
	return m.propagateUpdates(msg)
}

// handleTab toggles focus between software and details
func (m *model) handleTab() *model {
	if m.focus == focusSoftware {
		m.focus = focusDetails
		m.detailScroll = 0
		// Clamp uiActiveListIndex to valid range for visible or selectedKeys
		if m.softwarePaneLeft && len(m.visible) > 0 {
			if m.uiActiveListIndex >= len(m.visible) {
				m.uiActiveListIndex = len(m.visible) - 1
			}
			if m.uiActiveListIndex < 0 {
				m.uiActiveListIndex = 0
			}
		}
		if !m.softwarePaneLeft && len(m.selectedKeys) > 0 {
			if m.uiActiveListIndex >= len(m.selectedKeys) {
				m.uiActiveListIndex = len(m.selectedKeys) - 1
			}
			if m.uiActiveListIndex < 0 {
				m.uiActiveListIndex = 0
			}
		}
	} else {
		m.focus = focusSoftware
		// keep softwarePaneLeft as is
	}
	return m
}

// handleSoftwareKey handles key input for the software panes
func (m *model) handleSoftwareKey(key string) *model {
	if key == "/" {
		m.searchBar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
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
		if m.uiActiveListIndex < len(m.visible)-1 {
			m.uiActiveListIndex++
		}
	case "up", "k":
		if m.uiActiveListIndex > 0 {
			m.uiActiveListIndex--
		}
	case "right":
		// switch to right pane if any selected
		if len(m.selectedKeys) > 0 {
			m.softwarePaneLeft = false
			// Adjust uiActiveListIndex for the new pane
			if m.uiActiveListIndex >= len(m.selectedKeys) {
				m.uiActiveListIndex = len(m.selectedKeys) - 1
			}
			if m.uiActiveListIndex < 0 { // Should not happen if len > 0
				m.uiActiveListIndex = 0
			}
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
		if m.uiActiveListIndex < len(m.selectedKeys)-1 {
			m.uiActiveListIndex++
		}
	case "up", "k":
		if m.uiActiveListIndex > 0 {
			m.uiActiveListIndex--
		}
	case "left":
		// switch to left pane if any visible
		if len(m.visible) > 0 {
			m.softwarePaneLeft = true
			// Adjust uiActiveListIndex for the new pane
			if m.uiActiveListIndex >= len(m.visible) {
				m.uiActiveListIndex = len(m.visible) - 1
			}
			if m.uiActiveListIndex < 0 { // Should not happen if len > 0
				m.uiActiveListIndex = 0
			}
		}
	}
	return m
}

// wrap returns the string s wrapped to the given width using lipgloss styling.
//
// # Example
//
//	wrapped := wrap("some long text", 40)
func wrap(s string, width int) string {
	// Ensure width is not negative, lipgloss might panic or misbehave.
	if width < 0 {
		width = 0
	}
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(s)
}

// detailLines returns the lines to display in the details panel for the selected entry.
//
// # Returns
//   - []string: Each string is a line to display in the details panel.
func (m *model) detailLines(availableWidth int) []string { // Added availableWidth parameter
	if m.focus == focusSoftware && !m.softwarePaneLeft {
		// Right pane (selected)
		if len(m.selectedKeys) == 0 || m.uiActiveListIndex < 0 || m.uiActiveListIndex >= len(m.selectedKeys) {
			return m.noDetails(availableWidth) // Pass availableWidth
		}
		return m.detailsForKey(m.selectedKeys[m.uiActiveListIndex], availableWidth) // Pass availableWidth
	} else {
		// Left pane (unselected)
		if len(m.visible) == 0 || m.uiActiveListIndex < 0 || m.uiActiveListIndex >= len(m.visible) {
			return m.noDetails(availableWidth) // Pass availableWidth
		}
		return m.detailsForKey(m.visible[m.uiActiveListIndex], availableWidth) // Pass availableWidth
	}
}

// detailsForKey returns the details lines for a given manifest key
func (m *model) detailsForKey(key string, availableWidth int) []string { // Added availableWidth parameter
	entry := m.manifest[key]
	focused := m.focus == focusDetails
	styles := core.CurrentStyles() // Changed from ui.CurrentStyles()
	detailValueStyle := styles.DetailValueStyle
	if focused {
		detailValueStyle = styles.DetailValueActiveStyle
	}

	logical := []string{
		styles.HeaderStyle.Render("Details"),
		styles.DetailKey.Render("Name: ") + detailValueStyle.Render(entry.Name),
		styles.DetailKey.Render("Key: ") + detailValueStyle.Render(key),
		styles.DetailKey.Render("Desc: ") + detailValueStyle.Render(entry.Desc),
	}
	if len(entry.Bin) > 0 {
		logical = append(logical, styles.DetailKey.Render("Bin: ")+detailValueStyle.Render(strings.Join(entry.Bin, ", ")))
	}
	if len(entry.Brew) > 0 {
		logical = append(logical, styles.DetailKey.Render("Brew: ")+detailValueStyle.Render(strings.Join(entry.Brew, ", ")))
	}
	if len(entry.Apt) > 0 {
		logical = append(logical, styles.DetailKey.Render("Apt: ")+detailValueStyle.Render(strings.Join(entry.Apt, ", ")))
	}
	if len(entry.Pacman) > 0 {
		logical = append(logical, styles.DetailKey.Render("Pacman: ")+detailValueStyle.Render(strings.Join(entry.Pacman, ", ")))
	}
	if entry.Docs != "" {
		logical = append(logical, styles.DetailKey.Render("Docs: ")+detailValueStyle.Render(entry.Docs))
	}
	if entry.Github != "" {
		logical = append(logical, styles.DetailKey.Render("GitHub: ")+detailValueStyle.Render(entry.Github))
	}
	if entry.Home != "" {
		logical = append(logical, styles.DetailKey.Render("Home: ")+detailValueStyle.Render(entry.Home))
	}
	// Flatten to terminal lines
	var lines []string
	// Use availableWidth for wrapping, adjusted by DetailsPanelWrapPadding
	wrapWidth := availableWidth - core.DetailsPanelWrapPadding
	if wrapWidth < 0 { // Ensure wrapWidth is not negative
		wrapWidth = 0
	}
	for _, l := range logical {
		wrapped := wrap(l, wrapWidth) // Use calculated wrapWidth
		lines = append(lines, strings.Split(wrapped, "\\\\n")...)
	}
	return lines
}

// noDetails returns the placeholder lines for when no details are available
func (m *model) noDetails(_ int) []string { // Added availableWidth parameter
	// Potentially use availableWidth if "No details available." should be wrapped or styled based on it.
	// For now, it's simple text.
	return []string{
		core.CurrentStyles().HeaderStyle.Render("Details"),
		core.CurrentStyles().ItemStyle.Render("No details available."),
	}
}

// renderHelpView renders the help screen content.
func (m *model) renderHelpView(width int) string {
	helpStyle := lipgloss.NewStyle().Width(width).Padding(1, 2)
	helpTitle := core.CurrentStyles().HeaderStyle.Render("Help")
	helpBody := `
Keyboard Controls:
  ↑/↓/j/k:  Move selection
  Enter:    Select/Deselect item (in software lists)
            (No action in details panel from Enter)
  Tab:      Toggle focus (Software Lists ↔ Details Panel)
  /:        Start search (when focus is on Software Lists)
  Esc:      Cancel search / Close Help
  h:        Toggle Help
  q:        Quit

Focus Areas:
  - Software Lists: Left (Available) and Right (Selected) panes.
    - Use ←/→ to switch between Left and Right panes when focus is on Software Lists.
  - Details Panel: Shows information about the currently highlighted item.
    - Use ↑/↓/j/k to scroll content within the Details Panel.
`
	return helpStyle.Render(lipgloss.JoinVertical(lipgloss.Left, helpTitle, helpBody))
}

func renderHeader(title string, width int) string {
	style := core.CurrentStyles().HeaderStyle.Width(width).Align(lipgloss.Center)
	return style.Render(title)
}

func renderFooter(text string, width int) string {
	style := core.CurrentStyles().FooterStyle.Width(width).Align(lipgloss.Center)
	return style.Render(text)
}

func (m *model) moveToSelected() {
	// This function moves an item from the left pane (m.visible) to the right pane (m.selectedKeys)
	if !m.softwarePaneLeft || len(m.visible) == 0 || m.uiActiveListIndex < 0 || m.uiActiveListIndex >= len(m.visible) {
		return // Not in left pane, or list is empty, or index is out of bounds
	}

	keyToMove := m.visible[m.uiActiveListIndex]

	// Add to selectedKeys
	m.selectedKeys = append(m.selectedKeys, keyToMove)
	// Sort selectedKeys for consistent order (optional, but good for UX)
	sort.Strings(m.selectedKeys)

	// Re-filter, which will remove the keyToMove from m.visible
	m.filter()

	// Adjust uiActiveListIndex for m.visible
	if len(m.visible) == 0 {
		m.uiActiveListIndex = 0 // Or -1 if you prefer for empty lists
	} else if m.uiActiveListIndex >= len(m.visible) {
		m.uiActiveListIndex = len(m.visible) - 1
	}
	// If m.uiActiveListIndex became < 0 due to list emptying and then repopulating, reset to 0
	if m.uiActiveListIndex < 0 && len(m.visible) > 0 {
		m.uiActiveListIndex = 0
	}
}

func (m *model) moveToDeselected() {
	// This function moves an item from the right pane (m.selectedKeys) to the left pane (m.visible)
	if m.softwarePaneLeft || len(m.selectedKeys) == 0 || m.uiActiveListIndex < 0 || m.uiActiveListIndex >= len(m.selectedKeys) {
		return // Not in right pane, or list is empty, or index is out of bounds
	}

	// Remove the selected item at m.uiActiveListIndex from selectedKeys
	newSelectedKeys := make([]string, 0, len(m.selectedKeys)-1)
	for i, k := range m.selectedKeys {
		if i != m.uiActiveListIndex {
			newSelectedKeys = append(newSelectedKeys, k)
		}
	}
	m.selectedKeys = newSelectedKeys

	// Re-filter, which will make keyToMove available in m.visible again (if it matches search)
	m.filter()

	// Adjust uiActiveListIndex for m.selectedKeys
	if len(m.selectedKeys) == 0 {
		m.uiActiveListIndex = 0 // Or -1
	} else if m.uiActiveListIndex >= len(m.selectedKeys) {
		m.uiActiveListIndex = len(m.selectedKeys) - 1
	}
	// If m.uiActiveListIndex became < 0 due to list emptying and then repopulating, reset to 0
	if m.uiActiveListIndex < 0 && len(m.selectedKeys) > 0 {
		m.uiActiveListIndex = 0
	}
}

// Version is the application version
const Version = "0.1.0"

// loadConfig loads the application configuration based on command line flags
// and environment variables in the correct precedence order
func loadConfig(opts *flags.Options) (*config.Config, error) {
	var cfg *config.Config
	var configPath string

	// 1. Check command line flag for config path
	if opts.ConfigPath != "" {
		configPath = opts.ConfigPath
		// Verify the file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", configPath)
		}
	} else {
		// 2. Check environment variable or standard locations
		configPath = config.FindConfigFile()
	}

	// Load config from file or use defaults
	if configPath != "" {
		var err error
		cfg, err = config.Load(configPath)
		if err != nil {
			return nil, fmt.Errorf("error loading config from %s: %w", configPath, err)
		}
	} else {
		// No config file found, use defaults
		cfg = config.DefaultConfig()
	}

	// Override with command line flags if provided
	if opts.Debug {
		cfg.System.DebugMode = true
	}

	// Override manifest path if specified on command line
	if opts.ManifestPath != "" {
		cfg.Software.ManifestPath = opts.ManifestPath
	}

	// Override emoji setting if no-emojis flag is specified
	if opts.NoEmojis {
		cfg.UI.EmojisEnabled = false
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// initializeModel creates a new model with the given configuration
func initializeModel(cfg *config.Config) (*model, error) {
	// Validate the manifest path
	if err := cfg.ValidateManifestPath(); err != nil {
		return nil, fmt.Errorf("manifest validation error: %w", err)
	}

	// Resolve the manifest path to its absolute form
	manifestPath := cfg.ResolveManifestPath()

	// Load the software manifest
	manifestData, err := app.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error loading manifest from %s: %w", manifestPath, err)
	}

	// Get sorted keys from the manifest
	var entries []string
	for k := range manifestData {
		entries = append(entries, k)
	}
	sort.Strings(entries)

	// Create the initial model
	m := &model{
		manifest:          manifestData,
		entries:           entries,
		visible:           append([]string{}, entries...), // Initially all entries are visible
		selectedKeys:      []string{},                     // Initially no keys are selected
		softwarePaneLeft:  true,
		focus:             focusSoftware,
		uiActiveListIndex: 0,
		config:            cfg,
	}

	// Add preloaded keys to selected keys if they exist in the manifest
	for _, key := range cfg.Software.PreloadKeys {
		if _, exists := manifestData[key]; exists {
			m.selectedKeys = append(m.selectedKeys, key)
		}
	}

	// Sort the selected keys for consistency
	if len(m.selectedKeys) > 0 {
		sort.Strings(m.selectedKeys)
	}

	// Ensure valid index when entries list is empty
	if len(entries) == 0 {
		m.uiActiveListIndex = 0
	}

	return m, nil
}

func (m *model) View() string {
	if m.loadErr != nil {
		return fmt.Sprintf("Error loading manifest: %v\n", m.loadErr)
	}
	if m.width == 0 || m.height == 0 { // Not yet initialized
		return "Initializing..."
	}

	// Header
	titleText := "à la carte"
	if m.config.UI.EmojisEnabled {
		titleText += " 🛒"
	}
	header := renderHeader(titleText, m.contentWidth) // Use m.contentWidth

	// Search Bar
	searchBarView := m.searchBar.View()

	// Main Content Area (Top Split Pane + Details Panel)
	// Top Split Pane (Software Lists)
	leftPaneActualContentWidth := int(float64(m.contentWidth)*core.SplitPaneRatio) - (cardPadding+cardBorder)*2
	rightPaneActualContentWidth := m.contentWidth - int(float64(m.contentWidth)*core.SplitPaneRatio) - (cardPadding+cardBorder)*2
	if leftPaneActualContentWidth < 0 {
		leftPaneActualContentWidth = 0
	}
	if rightPaneActualContentWidth < 0 {
		rightPaneActualContentWidth = 0
	}

	leftPaneContent := m.renderList(m.visible, m.softwarePaneLeft && m.focus == focusSoftware, leftPaneActualContentWidth, true)
	rightPaneContent := m.renderList(m.selectedKeys, !m.softwarePaneLeft && m.focus == focusSoftware, rightPaneActualContentWidth, false)

	// Update the content of the panels within the SplitPaneLayout interface
	m.topSplitPane.SetLeftPanel(patterns.Panel(core.StringModel(leftPaneContent)))
	m.topSplitPane.SetRightPanel(patterns.Panel(core.StringModel(rightPaneContent)))
	topSplitPaneView := m.topSplitPane.View()

	// Details Panel
	currentDetailsData := &components.DetailsPanelData{
		Lines: m.detailLines(m.contentWidth),
	}
	if dpm, ok := m.detailsPanelModel.(*components.DetailsPanelModel); ok {
		dpm.SetData(currentDetailsData)
		dpm.SetFocused(m.focus == focusDetails)
		dpm.SetScroll(m.detailScroll)
	}
	detailsPanelContent := m.detailsPanelModel.View()

	// Container for Details Panel
	detailsContainer := core.NewContainer(
		core.StringModel(detailsPanelContent),
		core.WithBorderAll(),     // Restore the border around the details panel
		core.WithRoundedBorder(), // Match the rounded border style used in other panels
		core.WithPaddingAll(1),   // Match padding used in other panels
	)
	detailsContainerCtx := &core.LayoutContext{
		AvailableWidth:  m.contentWidth,
		AvailableHeight: detailHeight, // This is the target height for the container
		NestingLevel:    1,            // Assuming this is nested inside the main card's content area
	}
	detailsContainer.SetSize(m.contentWidth, detailHeight, detailsContainerCtx)
	detailsContainerView := detailsContainer.View()

	// Vertically join top split pane and details panel
	mainContentRendered := lipgloss.JoinVertical(lipgloss.Left, topSplitPaneView, detailsContainerView)

	// Footer
	var footerText string
	if m.showHelp {
		footerText = "Esc/h: Close Help | q: Quit"
	} else {
		footerText = "h: Help | /: Search | Tab: Focus | q: Quit"
	}
	footer := renderFooter(footerText, m.contentWidth)

	// Assemble all parts into a vertical layout
	panelLayout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		searchBarView,
		mainContentRendered,
		footer,
	)

	// Wrap the entire layout in a Card.
	finalViewCard := patterns.Card(core.StringModel(panelLayout))
	// The card itself needs to be sized to the full window width (m.width)
	// Its internal content (panelLayout) is m.contentWidth.
	// The Card pattern (core.Container) will handle its own padding/border.
	// When View() is called on the card, it uses its internally set width/height.
	// We need to ensure the card's SetSize is called appropriately.
	// This happens if the card is the root model or part of a chain where WindowSizeMsg propagates.
	// Here, we construct it fresh in each View. Let's set its size.
	cardCtx := &core.LayoutContext{AvailableWidth: m.width, AvailableHeight: m.height} // Card uses full window size
	finalViewCard.SetSize(m.width, m.height, cardCtx)
	finalView := finalViewCard.View()

	if m.showHelp {
		helpView := m.renderHelpView(m.contentWidth)
		// Help view should also be wrapped in a card for consistent styling if it's a full takeover
		helpCard := patterns.Card(core.StringModel(helpView))
		helpCard.SetSize(m.width, m.height, cardCtx) // Help card also uses full window size
		return helpCard.View()
	}

	return finalView
}

// renderList renders a list of items for a pane.
func (m *model) renderList(keys []string, focused bool, width int, isLeftPane bool) string {
	displayableItems := listHeight // This is a number of lines, not pixels

	if len(keys) == 0 {
		return m.renderEmptyList(width, isLeftPane)
	}

	start, end := m.calculateVisibleRange(keys, displayableItems)
	content := m.buildListContent(keys, start, end, focused, width)
	return m.ensureConsistentHeight(content, displayableItems)
}

// renderEmptyList handles the case when there are no items to display
func (m *model) renderEmptyList(width int, isLeftPane bool) string {
	styles := core.CurrentStyles()
	var emptyMsg string

	if isLeftPane {
		emptyMsg = core.ListEmptyMsg
	} else {
		emptyMsg = core.SelectedEmptyMsg
	}

	// Create a slice of 14 empty strings
	lines := make([]string, 14)

	// Place the centered message in the middle line
	middleLine := 14 / 2
	for i := 0; i < 14; i++ {
		if i == middleLine {
			lines[i] = styles.ItemStyle.Width(width).Align(lipgloss.Center).Render(emptyMsg)
		} else {
			lines[i] = styles.ItemStyle.Width(width).Render(" ")
		}
	}

	return strings.Join(lines, "\n")
}

// calculateVisibleRange determines which items should be visible in the view
func (m *model) calculateVisibleRange(keys []string, displayableItems int) (start, end int) {
	start = 0
	if m.uiActiveListIndex >= displayableItems {
		start = m.uiActiveListIndex - displayableItems + 1
	}

	end = start + displayableItems
	if end > len(keys) {
		end = len(keys)
	}
	if start > end { // Ensure start is not past end if keys list is very short
		start = end
	}

	return start, end
}

// buildListContent creates the content for the visible items
func (m *model) buildListContent(keys []string, start, end int, focused bool, width int) string {
	var s strings.Builder

	for i := start; i < end; i++ {
		if i < 0 || i >= len(keys) {
			continue
		}

		k := keys[i]
		e := m.manifest[k]

		formattedLine := m.formatItemLine(&e, i, focused, width)
		s.WriteString(formattedLine)
		s.WriteString("\n")
	}

	return s.String()
}

// formatItemLine formats a single item line with appropriate styling
func (m *model) formatItemLine(e *app.SoftwareEntry, index int, focused bool, width int) string {
	styles := core.CurrentStyles()
	itemStyle := styles.ItemStyle
	if focused && index == m.uiActiveListIndex {
		itemStyle = styles.ActiveItemStyle
	}

	textWidth := width - 2 // Corrected from width - 1
	if textWidth < 0 {
		textWidth = 0
	}

	line := m.formatItemText(e, textWidth)
	return itemStyle.Render(line)
}

// formatItemText handles text formatting with or without emoji
func (m *model) formatItemText(e *app.SoftwareEntry, textWidth int) string {
	line := e.Name

	if m.config.UI.EmojisEnabled {
		emoji := core.EmojiForEntry(e)
		emojiAdjustedTextWidth := textWidth - 3

		switch {
		case len(line) > emojiAdjustedTextWidth && emojiAdjustedTextWidth > 3:
			return emoji + " " + line[:emojiAdjustedTextWidth-3] + "..."
		case len(line) > emojiAdjustedTextWidth:
			return emoji + " " + line[:emojiAdjustedTextWidth]
		default:
			return emoji + " " + line
		}
	} else {
		switch {
		case len(line) > textWidth && textWidth > 3:
			return line[:textWidth-3] + "..."
		case len(line) > textWidth:
			return line[:textWidth]
		default:
			return line
		}
	}
}

// ensureConsistentHeight ensures the content has a consistent height
func (m *model) ensureConsistentHeight(content string, displayableItems int) string {
	result := content

	// Count the actual number of lines in the result
	lines := strings.Count(result, "\n") + 1

	// Add more newlines if needed to ensure exact displayableItems height
	if lines < displayableItems {
		result += strings.Repeat("\n", displayableItems-lines)
	}

	// Add an extra newline for height consistency with empty panes
	return result + "\n"
}

func main() {
	// Parse command line flags
	opts := flags.Parse()

	// Validate command line options
	if err := flags.ValidateOptions(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flags.Usage()
		os.Exit(1)
	}

	// Handle help flag
	if opts.Help {
		flags.Usage()
		return
	}

	// Handle version flag
	if opts.Version {
		output := fmt.Sprintf("chezmoi-a-la-carte version %s", Version)

		if opts.OutputFormat == "json" {
			jsonOutput, _ := config.FormatOutput(map[string]string{"version": Version}, config.OutputFormat(opts.OutputFormat))
			fmt.Println(jsonOutput)
		} else {
			fmt.Println(output)
		}
		return
	}

	// Load configuration
	cfg, err := loadConfig(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Print configuration information
	switch {
	case opts.Quiet:
		// Suppress output in quiet mode
	case cfg.System.DebugMode:
		fmt.Printf("Debug mode enabled\n")
		fmt.Println(cfg.String())

		// In debug mode, also print resolved manifest path
		fmt.Printf("Using manifest: %s\n", cfg.ResolveManifestPath())
	case cfg.ConfigPath != "":
		fmt.Printf("Loaded config from: %s\n", cfg.ConfigPath)
	default:
		fmt.Println("Using default settings (no config file found)")
	}

	// Initialize model
	initialModel, err := initializeModel(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	// Run the application
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
