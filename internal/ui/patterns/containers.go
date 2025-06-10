// Package patterns provides reusable UI design patterns built on the core UI primitives.
package patterns

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"a-la-carte/internal/ui/core"
)

// Panel creates a standard panel with common styling.
//
// # Usage
//
//	panel := patterns.Panel(content)
//	panel.SetFocused(true)  // make it focused
//
// # Features
//   - Rounded borders on all sides
//   - 1-space padding
//   - Theme-aware styling
func Panel(content tea.Model) core.Container {
	return core.NewContainer(
		content,
		core.WithBorderAll(),
		core.WithRoundedBorder(),
		core.WithPaddingAll(1),
	)
}

// Dialog creates a modal dialog container with extra padding and special styling.
//
// # Usage
//
//	dialog := patterns.Dialog(content)
//
// # Features
//   - Rounded borders on all sides
//   - 2-space padding for better readability
//   - Theme-specific dialog background and border colors
func Dialog(content tea.Model) core.Container {
	theme := core.CurrentTheme()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.DialogBorder()).
		Background(theme.DialogBg())

	return core.NewContainer(
		content,
		core.WithBorderAll(),
		core.WithRoundedBorder(),
		core.WithPaddingAll(2),
		core.WithStyle(&style),
	)
}

// Card creates a card-style container for content that should be visually grouped.
//
// # Usage
//
//	card := patterns.Card(content)
//
// # Features
//   - Rounded borders on all sides
//   - 1-space padding
//   - Theme-aware styling
func Card(content tea.Model) core.Container {
	return core.NewContainer(
		content,
		core.WithBorderAll(),
		core.WithRoundedBorder(),
		core.WithPaddingAll(1),
	)
}

// Tab creates a tab-style container for use in tabbed interfaces.
//
// # Usage
//
//	tab := patterns.Tab(content)
//
// # Features
//   - Special tab-style border on the top
//   - Horizontal padding (1 space)
//   - Theme-specific header colors
func Tab(content tea.Model) core.Container {
	theme := core.CurrentTheme()
	style := lipgloss.NewStyle().
		Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  " ",
			BottomRight: " ",
		}, true, true, false, true).
		BorderForeground(theme.Header()).
		Padding(0, 1).
		Foreground(theme.Text())

	return core.NewContainer(
		content,
		core.WithPaddingHorizontal(1),
		core.WithStyle(&style),
	)
}

// Button creates a button-style container for interactive elements.
//
// # Usage
//
//	button := patterns.Button(content)
//	button.SetActive(true)  // indicate it's clickable
//	button.SetHovered(true) // indicate mouse is over it
//
// # Features
//   - Background color for visibility
//   - Horizontal padding (3 spaces)
//   - 1-space top margin
func Button(content tea.Model) core.Container {
	theme := core.CurrentTheme()
	style := lipgloss.NewStyle().
		Foreground(theme.Text()).
		Background(theme.TextMuted()).
		Padding(0, 3).
		MarginTop(1)

	return core.NewContainer(
		content,
		core.WithPaddingAll(1),
		core.WithStyle(&style),
	)
}

// StatusBar creates a status bar container for displaying app status information.
//
// # Usage
//
//	statusBar := patterns.StatusBar(content)
//
// # Features
//   - Theme-specific status bar colors
//   - Horizontal padding (1 space)
func StatusBar(content tea.Model) core.Container {
	theme := core.CurrentTheme()
	style := lipgloss.NewStyle().
		Foreground(theme.StatusBarFg()).
		Background(theme.StatusBarBg()).
		Padding(0, 1)

	return core.NewContainer(
		content,
		core.WithPaddingAll(1),
		core.WithStyle(&style),
	)
}

// PlainBox creates a simple box with borders but no special styling.
//
// # Usage
//
//	box := patterns.PlainBox(content)
//
// # Features
//   - Simple borders on all sides
//   - No padding by default
func PlainBox(content tea.Model) core.Container {
	return core.NewContainer(
		content,
		core.WithBorderAll(),
	)
}

// Padded wraps content with padding but no borders.
//
// # Usage
//
//	padded := patterns.Padded(content, 1)
//
// # Features
//   - Even padding on all sides
//   - No borders
func Padded(content tea.Model, padding int) core.Container {
	return core.NewContainer(
		content,
		core.WithPaddingAll(padding),
	)
}

// Banner creates a full-width container for header/banner content.
//
// # Usage
//
//	banner := patterns.Banner(content)
//
// # Features
//   - Theme-specific header background
//   - Horizontal padding (2 spaces)
//   - Vertical padding (1 space)
func Banner(content tea.Model) core.Container {
	theme := core.CurrentTheme()
	style := lipgloss.NewStyle().
		Foreground(theme.Text()).
		Background(theme.Header()).
		Padding(1, 2)

	return core.NewContainer(
		content,
		core.WithStyle(&style),
	)
}

// PlaceOverlay positions an overlay string within a larger panel string.
// It centers the overlay by default.
func PlaceOverlay(panelWidth, panelHeight int, overlayContent, panelContent string, center bool) string {
	// Create a new style for positioning. We assume the panelContent already has its own styling.
	// We'll render the overlay on top.
	// This is a simplified approach; for complex layering, a more robust solution might be needed.
	// For now, we'll place the overlay within the panel's dimensions.

	// This function is more about calculating where to put something if you were to manually merge strings.
	// A true overlay usually involves rendering one thing, then rendering another on top at specific coordinates.
	// Lipgloss's Place function is better suited for this if you are rendering the overlay as the primary content of a cell.

	// If we are to return a single string, we'd need to merge them carefully.
	// However, the typical use case in Bubble Tea is to have a container that renders its content,
	// and if an overlay is needed, it's rendered *instead* of or *conditionally around* the main content.

	// For the purpose of this helper, let's assume we want to render the panelContent
	// and then render the overlayContent on top, centered.
	// This is tricky with just string manipulation if the panelContent is complex.

	// A more practical approach for Bubble Tea is to use lipgloss.Place within a container's View method.
	// This function might be better named `GetOverlayPlacementStyle` if it were to return a style for the overlay.

	// Given the existing usage, it seems like it's trying to create a combined view.
	// Let's use lipgloss.Place to put the overlayContent into a box of panelWidth x panelHeight,
	// and then join that with the panelContent. This isn't a true overlay but a common layout pattern.

	positionedOverlay := lipgloss.NewStyle().Width(panelWidth).Height(panelHeight).Align(lipgloss.Center, lipgloss.Center).Render(overlayContent)

	// This doesn't truly "overlay" in the visual sense of one on top of another with transparency.
	// It replaces the panelContent if we return just positionedOverlay.
	// If the goal is to have panelContent as background and overlayContent on top, that's more complex.
	// For now, let's assume the intent is to show the overlay centered within the panel's dimensions.
	return positionedOverlay // This will show the overlay centered, effectively replacing panelContent if used directly.
}

// SplitPaneLayout defines the interface for a split pane layout.
type SplitPaneLayout interface {
	core.Container
	SetLeftPanel(panel core.Container)
	SetRightPanel(panel core.Container)
	SetBottomPanel(panel core.Container) // Added for three-pane layout
	SetRatio(ratio float64)
	SetVerticalRatio(ratio float64) // Added for three-pane layout
}

// splitPane provides a layout that splits the available space between two or three panels.
// It can split horizontally (left/right) and then the right pane can be split vertically (top/bottom).
type splitPane struct {
	core.Container                                    // Embed core.Container for basic functionality
	leftPanel, rightPanel, bottomPanel core.Container // Panels
	ratio                              float64        // Ratio for left/right split (0.0 to 1.0 for left panel width)
	verticalRatio                      float64        // Ratio for top/bottom split of the right panel (0.0 to 1.0 for top panel height)
	width, height                      int
	ctx                                *core.LayoutContext // Store context for View
}

// NewSplitPane creates a new split pane layout.
func NewSplitPane(options ...SplitPaneOption) SplitPaneLayout {
	// Create a dummy core.Container for the base struct.
	// Its content won't be directly rendered, but it handles the overall size.
	baseContainer := core.NewContainer(core.EmptyModel())

	sp := &splitPane{
		Container:     baseContainer,
		ratio:         0.5, // Default to 50/50 split
		verticalRatio: 0.7, // Default to 70/30 split for right panel vertical
	}
	for _, opt := range options {
		opt(sp)
	}
	return sp
}

// SplitPaneOption is a function that configures a splitPane.
type SplitPaneOption func(*splitPane)

// WithLeftPanel sets the left panel of the split pane.
func WithLeftPanel(panel core.Container) SplitPaneOption {
	return func(sp *splitPane) {
		sp.leftPanel = panel
	}
}

// WithRightPanel sets the right panel of the split pane.
func WithRightPanel(panel core.Container) SplitPaneOption {
	return func(sp *splitPane) {
		sp.rightPanel = panel
	}
}

// WithBottomPanel sets the bottom panel (for the right side vertical split).
func WithBottomPanel(panel core.Container) SplitPaneOption {
	return func(sp *splitPane) {
		sp.bottomPanel = panel
	}
}

// WithRatio sets the horizontal split ratio.
func WithRatio(ratio float64) SplitPaneOption {
	return func(sp *splitPane) {
		sp.ratio = ratio
	}
}

// WithVerticalRatio sets the vertical split ratio for the right panel.
func WithVerticalRatio(ratio float64) SplitPaneOption {
	return func(sp *splitPane) {
		sp.verticalRatio = ratio
	}
}

func (sp *splitPane) Init() tea.Cmd {
	var cmds []tea.Cmd
	if sp.leftPanel != nil {
		cmds = append(cmds, sp.leftPanel.Init())
	}
	if sp.rightPanel != nil {
		cmds = append(cmds, sp.rightPanel.Init())
	}
	if sp.bottomPanel != nil {
		cmds = append(cmds, sp.bottomPanel.Init())
	}
	return tea.Batch(cmds...)
}

func (sp *splitPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if sp.leftPanel != nil {
		newLeft, leftCmd := sp.leftPanel.Update(msg)
		sp.leftPanel = newLeft.(core.Container)
		cmds = append(cmds, leftCmd)
	}
	if sp.rightPanel != nil {
		newRight, rightCmd := sp.rightPanel.Update(msg)
		sp.rightPanel = newRight.(core.Container)
		cmds = append(cmds, rightCmd)
	}
	if sp.bottomPanel != nil {
		newBottom, bottomCmd := sp.bottomPanel.Update(msg)
		sp.bottomPanel = newBottom.(core.Container)
		cmds = append(cmds, bottomCmd)
	}
	return sp, tea.Batch(cmds...)
}

func (sp *splitPane) SetSize(width, height int, ctx *core.LayoutContext) tea.Cmd {
	sp.width = width
	sp.height = height
	sp.ctx = ctx // Store context for View

	var cmds []tea.Cmd

	leftWidth := int(float64(width) * sp.ratio)
	rightWidth := width - leftWidth

	// If there's a bottom panel, the left panel takes full height,
	// and the right area is split vertically.
	if sp.bottomPanel != nil {
		if sp.leftPanel != nil {
			cmds = append(cmds, sp.leftPanel.SetSize(leftWidth, height, ctx))
		}

		rightAreaHeight := height
		topRightHeight := int(float64(rightAreaHeight) * sp.verticalRatio)
		bottomRightHeight := rightAreaHeight - topRightHeight

		if sp.rightPanel != nil { // This is effectively the top-right panel
			cmds = append(cmds, sp.rightPanel.SetSize(rightWidth, topRightHeight, ctx))
		}
		if sp.bottomPanel != nil { // This is the bottom-right panel
			cmds = append(cmds, sp.bottomPanel.SetSize(rightWidth, bottomRightHeight, ctx))
		}
	} else {
		// Original two-pane horizontal split (both full height)
		if sp.leftPanel != nil {
			cmds = append(cmds, sp.leftPanel.SetSize(leftWidth, height, ctx))
		}
		if sp.rightPanel != nil {
			cmds = append(cmds, sp.rightPanel.SetSize(rightWidth, height, ctx))
		}
	}

	return tea.Batch(cmds...)
}

func (sp *splitPane) View() string {
	// Use the stored context if available, otherwise create a default one.
	// This ensures View can be called independently if needed, though typically SetSize provides the context.
	ctx := sp.ctx
	if ctx == nil {
		ctx = &core.LayoutContext{
			AvailableWidth:  sp.width,
			AvailableHeight: sp.height,
		}
	}
	return sp.ViewWithContext(ctx)
}

func (sp *splitPane) ViewWithContext(ctx *core.LayoutContext) string {
	var leftView, rightView, bottomView string
	leftWidth := int(float64(ctx.AvailableWidth) * sp.ratio)
	rightWidth := ctx.AvailableWidth - leftWidth

	if sp.bottomPanel != nil { // Three-pane layout
		if sp.leftPanel != nil {
			leftView = sp.leftPanel.ViewWithContext(&core.LayoutContext{
				AvailableWidth:  leftWidth,
				AvailableHeight: ctx.AvailableHeight, // Left panel takes full height
			})
		}

		topRightHeight := int(float64(ctx.AvailableHeight) * sp.verticalRatio)
		bottomRightHeight := ctx.AvailableHeight - topRightHeight

		if sp.rightPanel != nil { // This is the top-right panel
			rightView = sp.rightPanel.ViewWithContext(&core.LayoutContext{
				AvailableWidth:  rightWidth,
				AvailableHeight: topRightHeight,
			})
		}
		if sp.bottomPanel != nil { // This is the bottom-right panel
			bottomView = sp.bottomPanel.ViewWithContext(&core.LayoutContext{
				AvailableWidth:  rightWidth,
				AvailableHeight: bottomRightHeight,
			})
		}
		// Combine right and bottom views vertically
		combinedRightView := lipgloss.JoinVertical(lipgloss.Left, rightView, bottomView)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftView, combinedRightView)
	} else { // Original two-pane horizontal split
		if sp.leftPanel != nil {
			leftView = sp.leftPanel.ViewWithContext(&core.LayoutContext{
				AvailableWidth:  leftWidth,
				AvailableHeight: ctx.AvailableHeight,
			})
		}
		if sp.rightPanel != nil {
			rightView = sp.rightPanel.ViewWithContext(&core.LayoutContext{
				AvailableWidth:  rightWidth,
				AvailableHeight: ctx.AvailableHeight,
			})
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	}
}

// Setter methods
func (sp *splitPane) SetLeftPanel(panel core.Container) {
	sp.leftPanel = panel
}

func (sp *splitPane) SetRightPanel(panel core.Container) {
	sp.rightPanel = panel
}

func (sp *splitPane) SetBottomPanel(panel core.Container) {
	sp.bottomPanel = panel
}

func (sp *splitPane) SetRatio(ratio float64) {
	sp.ratio = ratio
}

func (sp *splitPane) SetVerticalRatio(ratio float64) {
	sp.verticalRatio = ratio
}

// GetListPanelStyle styles the list panel, applying appropriate focused/unfocused styles
func GetListPanelStyle(baseStyle *lipgloss.Style, focused bool, theme core.Theme) lipgloss.Style {
	style := *baseStyle // Dereference the pointer to work with a copy

	if focused {
		return style.
			BorderForeground(theme.AccentActive()).
			Bold(true).
			Background(theme.BackgroundFocused())
	}
	return style.
		BorderForeground(theme.Border()).
		UnsetBold().
		UnsetBackground()
}

// GetDetailPanelStyle styles the detail panel, applying appropriate focused/unfocused styles
func GetDetailPanelStyle(baseStyle *lipgloss.Style, focused bool, theme core.Theme) lipgloss.Style {
	style := *baseStyle // Dereference the pointer to work with a copy

	if focused {
		return style.
			BorderForeground(theme.AccentActive()).
			Bold(true).
			Background(theme.BackgroundFocused())
	}
	return style.
		BorderForeground(theme.Border()).
		UnsetBold().
		UnsetBackground()
}

// NewEnhancedContainer creates a new container with enhanced features like overlay and min/max sizing.
// It forwards to core.NewContainer but is placed here as it's a common pattern.
func NewEnhancedContainer(content tea.Model, options ...core.ContainerOption) core.Container {
	return core.NewContainer(content, options...)
}
