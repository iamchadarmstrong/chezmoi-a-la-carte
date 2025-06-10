package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"a-la-carte/internal/ui/core"
)

// SplitPaneLayout manages left, right, and bottom panels with dynamic sizing and layout.
//
// # Usage
//
//	layout := NewSplitPane(
//	  WithLeftPanel(left),
//	  WithRightPanel(right),
//	  WithBottomPanel(bottom),
//	  WithRatio(0.7),
//	  WithVerticalRatio(0.9),
//	)
type SplitPaneLayout interface {
	tea.Model
	SetLeftPanel(panel core.Container) tea.Cmd
	SetRightPanel(panel core.Container) tea.Cmd
	SetBottomPanel(panel core.Container) tea.Cmd
	ClearLeftPanel() tea.Cmd
	ClearRightPanel() tea.Cmd
	ClearBottomPanel() tea.Cmd
	SetSize(width, height int, ctx *core.LayoutContext) tea.Cmd
	GetSize() (width, height int)
	ViewWithContext(ctx *core.LayoutContext) string
}

type splitPaneLayout struct {
	width, height                      int
	ratio, verticalRatio               float64
	rightPanel, leftPanel, bottomPanel core.Container
}

func (s *splitPaneLayout) Init() tea.Cmd {
	var cmds []tea.Cmd
	if s.leftPanel != nil {
		cmds = append(cmds, s.leftPanel.Init())
	}
	if s.rightPanel != nil {
		cmds = append(cmds, s.rightPanel.Init())
	}
	if s.bottomPanel != nil {
		cmds = append(cmds, s.bottomPanel.Init())
	}
	return tea.Batch(cmds...)
}

func (s *splitPaneLayout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if winMsg, ok := msg.(tea.WindowSizeMsg); ok {
		cmd := s.SetSize(winMsg.Width, winMsg.Height, &core.LayoutContext{
			AvailableWidth:  winMsg.Width,
			AvailableHeight: winMsg.Height,
			NestingLevel:    0,
		})
		return s, cmd
	}
	if s.rightPanel != nil {
		u, cmd := s.rightPanel.Update(msg)
		s.rightPanel = u.(core.Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if s.leftPanel != nil {
		u, cmd := s.leftPanel.Update(msg)
		s.leftPanel = u.(core.Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if s.bottomPanel != nil {
		u, cmd := s.bottomPanel.Update(msg)
		s.bottomPanel = u.(core.Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return s, tea.Batch(cmds...)
}

func (s *splitPaneLayout) View() string {
	var topSection string
	switch {
	case s.leftPanel != nil && s.rightPanel != nil:
		leftView := s.leftPanel.View()
		rightView := s.rightPanel.View()
		topSection = lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	case s.leftPanel != nil:
		topSection = s.leftPanel.View()
	case s.rightPanel != nil:
		topSection = s.rightPanel.View()
	default:
		topSection = ""
	}
	var finalView string
	switch {
	case s.bottomPanel != nil && topSection != "":
		bottomView := s.bottomPanel.View()
		finalView = lipgloss.JoinVertical(lipgloss.Left, topSection, bottomView)
	case s.bottomPanel != nil:
		finalView = s.bottomPanel.View()
	default:
		finalView = topSection
	}
	if finalView != "" {
		style := lipgloss.NewStyle().Width(s.width).Height(s.height)
		return style.Render(finalView)
	}
	return finalView
}

// calculatePanelDimensions calculates the dimensions for all panels
func (s *splitPaneLayout) calculatePanelDimensions(ctx *core.LayoutContext) (leftWidth, rightWidth, bottomHeight int) {
	pickerHeightVal := core.CurrentTheme().SoftwarePickerHeight()

	// Calculate bottom panel height
	if s.bottomPanel != nil {
		bottomHeight = ctx.AvailableHeight - pickerHeightVal
		if bottomHeight < 0 {
			bottomHeight = 0
		}
	}

	// Calculate left and right panel widths
	s.ratio = 0.5 // Always enforce even split for two-pane layout
	width := ctx.AvailableWidth

	switch {
	case s.leftPanel != nil && s.rightPanel != nil:
		if width%2 != 0 {
			width--
		}
		leftWidth = width / 2
		rightWidth = width / 2
		// ENFORCE: Both panels must have identical border/padding for correct split alignment
		leftBorder, leftPad := getPanelBorderPadding(s.leftPanel)
		rightBorder, rightPad := getPanelBorderPadding(s.rightPanel)
		if leftBorder != rightBorder || leftPad != rightPad {
			panic("SplitPaneLayout: Left and right panels must have identical border and padding configuration for correct layout. (left: border=" + leftBorder + ", pad=" + leftPad + ", right: border=" + rightBorder + ", pad=" + rightPad + ")")
		}
	case s.leftPanel != nil:
		leftWidth = width
		rightWidth = 0
	case s.rightPanel != nil:
		leftWidth = 0
		rightWidth = width
	}

	return leftWidth, rightWidth, bottomHeight
}

// preparePanelContexts prepares the layout contexts for each panel
func (s *splitPaneLayout) preparePanelContexts(ctx *core.LayoutContext, leftWidth, rightWidth, bottomHeight int) (leftCtx, rightCtx, bottomCtx *core.LayoutContext) {
	pickerHeightVal := core.CurrentTheme().SoftwarePickerHeight()

	leftCtx = &core.LayoutContext{
		AvailableWidth:  leftWidth,
		AvailableHeight: pickerHeightVal,
		NestingLevel:    ctx.NestingLevel + 1,
	}

	rightCtx = &core.LayoutContext{
		AvailableWidth:  rightWidth,
		AvailableHeight: pickerHeightVal,
		NestingLevel:    ctx.NestingLevel + 1,
	}

	bottomCtx = &core.LayoutContext{
		AvailableWidth:  ctx.AvailableWidth,
		AvailableHeight: bottomHeight,
		NestingLevel:    ctx.NestingLevel + 1,
	}

	return leftCtx, rightCtx, bottomCtx
}

// renderTopSection renders the top section of the split pane
func (s *splitPaneLayout) renderTopSection(leftCtx, rightCtx *core.LayoutContext) string {
	switch {
	case s.leftPanel != nil && s.rightPanel != nil:
		leftView := s.leftPanel.ViewWithContext(leftCtx)
		rightView := s.rightPanel.ViewWithContext(rightCtx)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	case s.leftPanel != nil:
		return s.leftPanel.ViewWithContext(leftCtx)
	case s.rightPanel != nil:
		return s.rightPanel.ViewWithContext(rightCtx)
	default:
		return ""
	}
}

// renderFinalView renders the final view combining top and bottom sections
func (s *splitPaneLayout) renderFinalView(topSection string, bottomCtx *core.LayoutContext) string {
	switch {
	case s.bottomPanel != nil && topSection != "":
		bottomView := s.bottomPanel.ViewWithContext(bottomCtx)
		return lipgloss.JoinVertical(lipgloss.Left, topSection, bottomView)
	case s.bottomPanel != nil:
		return s.bottomPanel.ViewWithContext(bottomCtx)
	default:
		return topSection
	}
}

func (s *splitPaneLayout) ViewWithContext(ctx *core.LayoutContext) string {
	// Calculate panel dimensions
	leftWidth, rightWidth, bottomHeight := s.calculatePanelDimensions(ctx)

	// Prepare contexts for each panel
	leftCtx, rightCtx, bottomCtx := s.preparePanelContexts(ctx, leftWidth, rightWidth, bottomHeight)

	// Render top section
	topSection := s.renderTopSection(leftCtx, rightCtx)

	// Render final view
	finalView := s.renderFinalView(topSection, bottomCtx)

	// Apply final styling if needed
	if finalView != "" {
		style := lipgloss.NewStyle().Width(ctx.AvailableWidth).Height(ctx.AvailableHeight)
		return style.Render(finalView)
	}
	return finalView
}

// getPanelBorderPadding returns a string representation of border and padding config for a panel
func getPanelBorderPadding(panel core.Container) (leftPadding, rightPadding string) {
	border := "T" + b2s(panel.GetBorderTop()) +
		"R" + b2s(panel.GetBorderRight()) +
		"B" + b2s(panel.GetBorderBottom()) +
		"L" + b2s(panel.GetBorderLeft())

	pad := "T" + itos(panel.GetPaddingTop()) +
		"R" + itos(panel.GetPaddingRight()) +
		"B" + itos(panel.GetPaddingBottom()) +
		"L" + itos(panel.GetPaddingLeft())

	return border, pad
}

func b2s(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func itos(i int) string {
	return fmt.Sprintf("%d", i)
}

func (s *splitPaneLayout) SetSize(width, height int, ctx *core.LayoutContext) tea.Cmd {
	s.width = width
	s.height = height
	if ctx == nil {
		panic("LayoutContext must not be nil: all SetSize and ViewWithContext calls must provide a context")
	}
	pickerHeightVal := core.CurrentTheme().SoftwarePickerHeight() // Call the method to get the int value
	var bottomHeight int
	if s.bottomPanel != nil {
		bottomHeight = height - pickerHeightVal
		if bottomHeight < 0 {
			bottomHeight = 0
		}
	} else {
		bottomHeight = 0
	}
	var leftWidth, rightWidth int
	s.ratio = 0.5 // Always enforce even split for two-pane layout
	w := width
	switch {
	case s.leftPanel != nil && s.rightPanel != nil:
		if w%2 != 0 {
			w--
		}
		leftWidth = w / 2
		rightWidth = w / 2
	case s.leftPanel != nil:
		leftWidth = w
		rightWidth = 0
	case s.rightPanel != nil:
		leftWidth = 0
		rightWidth = w
	}
	var cmds []tea.Cmd
	if s.leftPanel != nil {
		panelCtx := &core.LayoutContext{
			AvailableWidth:  leftWidth,
			AvailableHeight: pickerHeightVal,
			NestingLevel:    ctx.NestingLevel + 1,
		}
		cmd := s.leftPanel.SetSize(leftWidth, pickerHeightVal, panelCtx)
		cmds = append(cmds, cmd)
	}
	if s.rightPanel != nil {
		panelCtx := &core.LayoutContext{
			AvailableWidth:  rightWidth,
			AvailableHeight: pickerHeightVal,
			NestingLevel:    ctx.NestingLevel + 1,
		}
		cmd := s.rightPanel.SetSize(rightWidth, pickerHeightVal, panelCtx)
		cmds = append(cmds, cmd)
	}
	if s.bottomPanel != nil {
		panelCtx := &core.LayoutContext{
			AvailableWidth:  width,
			AvailableHeight: bottomHeight,
			NestingLevel:    ctx.NestingLevel + 1,
		}
		cmd := s.bottomPanel.SetSize(width, bottomHeight, panelCtx)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (s *splitPaneLayout) GetSize() (width, height int) {
	return s.width, s.height
}

func (s *splitPaneLayout) SetLeftPanel(panel core.Container) tea.Cmd {
	s.leftPanel = panel
	if s.width > 0 && s.height > 0 {
		ctx := &core.LayoutContext{
			AvailableWidth:  s.width,
			AvailableHeight: s.height,
			NestingLevel:    0,
		}
		return s.SetSize(s.width, s.height, ctx)
	}
	return nil
}

func (s *splitPaneLayout) SetRightPanel(panel core.Container) tea.Cmd {
	s.rightPanel = panel
	if s.width > 0 && s.height > 0 {
		ctx := &core.LayoutContext{
			AvailableWidth:  s.width,
			AvailableHeight: s.height,
			NestingLevel:    0,
		}
		return s.SetSize(s.width, s.height, ctx)
	}
	return nil
}

func (s *splitPaneLayout) SetBottomPanel(panel core.Container) tea.Cmd {
	s.bottomPanel = panel
	if s.width > 0 && s.height > 0 {
		ctx := &core.LayoutContext{
			AvailableWidth:  s.width,
			AvailableHeight: s.height,
			NestingLevel:    0,
		}
		return s.SetSize(s.width, s.height, ctx)
	}
	return nil
}

func (s *splitPaneLayout) ClearLeftPanel() tea.Cmd {
	s.leftPanel = nil
	if s.width > 0 && s.height > 0 {
		ctx := &core.LayoutContext{
			AvailableWidth:  s.width,
			AvailableHeight: s.height,
			NestingLevel:    0,
		}
		return s.SetSize(s.width, s.height, ctx)
	}
	return nil
}

func (s *splitPaneLayout) ClearRightPanel() tea.Cmd {
	s.rightPanel = nil
	if s.width > 0 && s.height > 0 {
		ctx := &core.LayoutContext{
			AvailableWidth:  s.width,
			AvailableHeight: s.height,
			NestingLevel:    0,
		}
		return s.SetSize(s.width, s.height, ctx)
	}
	return nil
}

func (s *splitPaneLayout) ClearBottomPanel() tea.Cmd {
	s.bottomPanel = nil
	if s.width > 0 && s.height > 0 {
		ctx := &core.LayoutContext{
			AvailableWidth:  s.width,
			AvailableHeight: s.height,
			NestingLevel:    0,
		}
		return s.SetSize(s.width, s.height, ctx)
	}
	return nil
}

// SplitPaneOption configures a SplitPaneLayout.
type SplitPaneOption func(*splitPaneLayout)

// NewSplitPane creates a new SplitPaneLayout with the given options.
func NewSplitPane(options ...SplitPaneOption) SplitPaneLayout {
	layout := &splitPaneLayout{ratio: 0.7, verticalRatio: 0.9}
	for _, opt := range options {
		opt(layout)
	}
	return layout
}

func WithLeftPanel(panel core.Container) SplitPaneOption {
	return func(s *splitPaneLayout) { s.leftPanel = panel }
}

func WithRightPanel(panel core.Container) SplitPaneOption {
	return func(s *splitPaneLayout) { s.rightPanel = panel }
}

func WithBottomPanel(panel core.Container) SplitPaneOption {
	return func(s *splitPaneLayout) { s.bottomPanel = panel }
}

func WithRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) { s.ratio = ratio }
}

func WithVerticalRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) { s.verticalRatio = ratio }
}
