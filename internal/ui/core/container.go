// Package core provides the foundational elements for UI components.
package core

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LayoutContext tracks available space and nesting for dynamic layout.
type LayoutContext struct {
	AvailableWidth  int
	AvailableHeight int
	PaddingLeft     int
	PaddingRight    int
	PaddingTop      int
	PaddingBottom   int
	BorderLeft      int
	BorderRight     int
	BorderTop       int
	BorderBottom    int
	NestingLevel    int
}

// Container is the basic building block for all UI elements.
// It provides a consistent way to lay out content with borders, padding, and theming.
//
// # Usage
//
//	c := NewContainer(content, WithPadding(1), WithBorderAll(), WithRoundedBorder())
//
// # Options
//   - Padding (all sides)
//   - Border (any side, any style)
//   - Themed background and border colors
//   - Semantic styling (focused, active, hover states)
type Container interface {
	tea.Model
	SetSize(width, height int, ctx *LayoutContext) tea.Cmd
	ViewWithContext(ctx *LayoutContext) string

	// Border and padding getter methods
	GetBorderTop() bool
	GetBorderRight() bool
	GetBorderBottom() bool
	GetBorderLeft() bool
	GetPaddingTop() int
	GetPaddingRight() int
	GetPaddingBottom() int
	GetPaddingLeft() int
}

// ContainerState represents the current state of a container
type ContainerState struct {
	Focused bool
	Active  bool
	Hovered bool
}

type container struct {
	width, height                                        int
	content                                              tea.Model
	paddingTop, paddingRight, paddingBottom, paddingLeft int
	borderTop, borderRight, borderBottom, borderLeft     bool
	borderStyle                                          lipgloss.Border
	customStyle                                          *lipgloss.Style                // Optional: overrides default style if set
	overlayFunc                                          func(width, height int) string // Optional: overlay to render instead of content
	// State management
	state         ContainerState
	onStateChange func(ContainerState) // Optional: callback for state changes
	// Enhanced features
	minWidth, minHeight int
	maxWidth, maxHeight int
	ariaLabel           string
	debug               bool
	transitionDuration  int
}

func (c *container) Init() tea.Cmd {
	return c.content.Init()
}

func (c *container) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	u, cmd := c.content.Update(msg)
	c.content = u
	return c, cmd
}

// prepareBaseStyle prepares the base style for the container
func (c *container) prepareBaseStyle() lipgloss.Style {
	if c.customStyle != nil {
		return *c.customStyle
	}
	return lipgloss.NewStyle()
}

// calculateContentWidth calculates the width available for content
func (c *container) calculateContentWidth() int {
	width := c.width
	if c.borderLeft {
		width--
	}
	if c.borderRight {
		width--
	}
	return width
}

// applyStateStyle applies styling based on container state
func (c *container) applyStateStyle(style *lipgloss.Style, t Theme) lipgloss.Style {
	if c.state.Focused {
		return style.Background(t.BackgroundFocused())
	}
	if c.state.Active {
		return style.Background(t.BackgroundActive())
	}
	return *style
}

// renderOverlayContent renders the overlay content if present
func (c *container) renderOverlayContent() string {
	if c.overlayFunc == nil {
		return c.content.View()
	}

	overlay := c.overlayFunc(c.width, c.height)
	if overlay == "" {
		return c.content.View()
	}

	w := c.width - c.paddingLeft - c.paddingRight
	h := c.height - c.paddingTop - c.paddingBottom
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	return lipgloss.Place(
		w,
		h,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
	)
}

// addDebugOverlay adds debug information to the view if enabled
func (c *container) addDebugOverlay(view string) string {
	if !c.debug {
		return view
	}

	debugInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				"Size: "+string(rune(c.width))+"x"+string(rune(c.height)),
				"Min: "+string(rune(c.minWidth))+"x"+string(rune(c.minHeight)),
				"Max: "+string(rune(c.maxWidth))+"x"+string(rune(c.maxHeight)),
			),
		)
	return lipgloss.JoinVertical(lipgloss.Left, view, debugInfo)
}

func (c *container) View() string {
	t := CurrentTheme()
	style := c.prepareBaseStyle()

	if c.customStyle == nil {
		width := c.calculateContentWidth()

		// Apply border styling
		if c.borderTop || c.borderRight || c.borderBottom || c.borderLeft {
			style = c.applyBorderStyle(&style, t)
		}

		// Apply state-based styling
		style = c.applyStateStyle(&style, t)

		// Apply dimensions and padding
		style = style.Width(width).
			Height(c.height).
			PaddingTop(c.paddingTop).
			PaddingRight(c.paddingRight).
			PaddingBottom(c.paddingBottom).
			PaddingLeft(c.paddingLeft)
	} else {
		style = style.Height(c.height)
	}

	// Render content
	inner := c.renderOverlayContent()
	view := style.Render(inner)

	// Add debug overlay if enabled
	return c.addDebugOverlay(view)
}

// prepareContainerStyle creates and configures the container's style
func (c *container) prepareContainerStyle(ctx *LayoutContext) (style lipgloss.Style, padding int) {
	t := CurrentTheme()
	style = lipgloss.NewStyle()
	width := ctx.AvailableWidth

	if c.customStyle != nil {
		return *c.customStyle, width
	}

	// Apply min/max constraints
	if c.minWidth > 0 && width < c.minWidth {
		width = c.minWidth
	}
	if c.maxWidth > 0 && width > c.maxWidth {
		width = c.maxWidth
	}

	// Apply border styles
	if c.borderTop || c.borderRight || c.borderBottom || c.borderLeft {
		if c.borderLeft {
			width--
		}
		if c.borderRight {
			width--
		}
		style = c.applyBorderStyle(&style, t)
	}

	// Apply background based on state
	if c.state.Focused {
		style = style.Background(t.BackgroundFocused())
	} else if c.state.Active {
		style = style.Background(t.BackgroundActive())
	}

	// Apply padding
	style = style.Width(width).
		PaddingTop(c.paddingTop).
		PaddingRight(c.paddingRight).
		PaddingBottom(c.paddingBottom).
		PaddingLeft(c.paddingLeft)

	return style, width
}

// applyBorderStyle applies the border styling based on container state
func (c *container) applyBorderStyle(style *lipgloss.Style, t Theme) lipgloss.Style {
	var borderColor lipgloss.AdaptiveColor
	switch {
	case c.state.Focused:
		borderColor = t.BorderActive()
	case c.state.Active:
		borderColor = t.Accent()
	case c.state.Hovered:
		borderColor = t.AccentActive()
	default:
		borderColor = t.Border()
	}

	return style.Border(c.borderStyle, c.borderTop, c.borderRight, c.borderBottom, c.borderLeft).
		BorderForeground(borderColor)
}

// adjustHeight applies min/max constraints to the height
func (c *container) adjustHeight(height int) int {
	if c.minHeight > 0 && height < c.minHeight {
		height = c.minHeight
	}
	if c.maxHeight > 0 && height > c.maxHeight {
		height = c.maxHeight
	}
	return height
}

// createInnerContext creates a new LayoutContext for child components
func (c *container) createInnerContext(ctx *LayoutContext) *LayoutContext {
	return &LayoutContext{
		AvailableWidth:  ctx.AvailableWidth - c.paddingLeft - c.paddingRight - btoi(c.borderLeft) - btoi(c.borderRight),
		AvailableHeight: ctx.AvailableHeight - c.paddingTop - c.paddingBottom - btoi(c.borderTop) - btoi(c.borderBottom),
		PaddingLeft:     c.paddingLeft,
		PaddingRight:    c.paddingRight,
		PaddingTop:      c.paddingTop,
		PaddingBottom:   c.paddingBottom,
		BorderLeft:      btoi(c.borderLeft),
		BorderRight:     btoi(c.borderRight),
		BorderTop:       btoi(c.borderTop),
		BorderBottom:    btoi(c.borderBottom),
		NestingLevel:    ctx.NestingLevel + 1,
	}
}

// renderInnerContent renders the content or overlay
func (c *container) renderInnerContent(innerCtx *LayoutContext) string {
	// Handle overlay if present
	if c.overlayFunc != nil {
		overlay := c.overlayFunc(innerCtx.AvailableWidth, innerCtx.AvailableHeight)
		if overlay != "" {
			return c.renderOverlay(innerCtx, overlay)
		}
	}

	// Render content
	return c.renderContent(innerCtx)
}

// renderOverlay renders the overlay content centered
func (c *container) renderOverlay(ctx *LayoutContext, overlay string) string {
	w := ctx.AvailableWidth
	h := ctx.AvailableHeight
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return lipgloss.Place(
		w,
		h,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
	)
}

// renderContent renders the container's content
func (c *container) renderContent(ctx *LayoutContext) string {
	if vwc, ok := c.content.(interface{ ViewWithContext(*LayoutContext) string }); ok {
		return vwc.ViewWithContext(ctx)
	}
	return c.content.View()
}

// addDebugInfo adds debug information to the rendered view
func (c *container) addDebugInfo(view string, ctx *LayoutContext) string {
	debugInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				"Size: "+string(rune(ctx.AvailableWidth))+"x"+string(rune(ctx.AvailableHeight)),
				"Min: "+string(rune(c.minWidth))+"x"+string(rune(c.minHeight)),
				"Max: "+string(rune(c.maxWidth))+"x"+string(rune(c.maxHeight)),
			),
		)
	return lipgloss.JoinVertical(lipgloss.Left, view, debugInfo)
}

func (c *container) ViewWithContext(ctx *LayoutContext) string {
	// Prepare style and get adjusted width
	style, _ := c.prepareContainerStyle(ctx)

	// Adjust height
	height := c.adjustHeight(ctx.AvailableHeight)
	style = style.Height(height)

	// Create inner context
	innerCtx := c.createInnerContext(ctx)

	// Render inner content
	inner := c.renderInnerContent(innerCtx)

	// Apply style to content
	view := style.Render(inner)

	// Add debug visualization if enabled
	if c.debug {
		view = c.addDebugInfo(view, ctx)
	}

	return view
}

// Helper to convert bool to int
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// applyConstraints applies min/max constraints to dimensions
func (c *container) applyConstraints(width, height int) (newWidth, newHeight int) {
	if c.minWidth > 0 && width < c.minWidth {
		width = c.minWidth
	}
	if c.maxWidth > 0 && width > c.maxWidth {
		width = c.maxWidth
	}
	if c.minHeight > 0 && height < c.minHeight {
		height = c.minHeight
	}
	if c.maxHeight > 0 && height > c.maxHeight {
		height = c.maxHeight
	}
	return width, height
}

// calculateChildDimensions calculates the dimensions for child components
func (c *container) calculateChildDimensions(width, height int) (childWidth, childHeight int) {
	// Adjust for padding and borders
	hSpace := c.paddingLeft + c.paddingRight
	if c.borderLeft {
		hSpace++
	}
	if c.borderRight {
		hSpace++
	}
	vSpace := c.paddingTop + c.paddingBottom
	if c.borderTop {
		vSpace++
	}
	if c.borderBottom {
		vSpace++
	}

	w := width - hSpace
	h := height - vSpace

	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	return w, h
}

func (c *container) SetSize(width, height int, ctx *LayoutContext) tea.Cmd {
	// Apply min/max constraints
	width, height = c.applyConstraints(width, height)

	c.width = width
	c.height = height

	// Propagate size to content if possible
	if sizeable, ok := c.content.(interface {
		SetSize(int, int, *LayoutContext) tea.Cmd
	}); ok {
		childWidth, childHeight := c.calculateChildDimensions(width, height)
		return sizeable.SetSize(childWidth, childHeight, ctx)
	}
	return nil
}

func (c *container) GetSize() (width, height int) {
	return c.width, c.height
}

// State management methods
func (c *container) SetFocused(focused bool) {
	if c.state.Focused != focused {
		c.state.Focused = focused
		if c.onStateChange != nil {
			c.onStateChange(c.state)
		}
	}
}

func (c *container) SetActive(active bool) {
	if c.state.Active != active {
		c.state.Active = active
		if c.onStateChange != nil {
			c.onStateChange(c.state)
		}
	}
}

func (c *container) SetHovered(hovered bool) {
	if c.state.Hovered != hovered {
		c.state.Hovered = hovered
		if c.onStateChange != nil {
			c.onStateChange(c.state)
		}
	}
}

func (c *container) GetState() ContainerState {
	return c.state
}

// ContainerOption configures a Container.
type ContainerOption func(*container)

// NewContainer creates a new container with the given content and options.
func NewContainer(content tea.Model, options ...ContainerOption) Container {
	c := &container{content: content, borderStyle: lipgloss.NormalBorder()}
	for _, opt := range options {
		opt(c)
	}
	return c
}

// Padding options
func WithPadding(top, right, bottom, left int) ContainerOption {
	return func(c *container) {
		c.paddingTop = top
		c.paddingRight = right
		c.paddingBottom = bottom
		c.paddingLeft = left
	}
}

func WithPaddingAll(p int) ContainerOption {
	return WithPadding(p, p, p, p)
}

func WithPaddingHorizontal(p int) ContainerOption {
	return func(c *container) { c.paddingLeft = p; c.paddingRight = p }
}

func WithPaddingVertical(p int) ContainerOption {
	return func(c *container) { c.paddingTop = p; c.paddingBottom = p }
}

// Border options
func WithBorder(top, right, bottom, left bool) ContainerOption {
	return func(c *container) {
		c.borderTop = top
		c.borderRight = right
		c.borderBottom = bottom
		c.borderLeft = left
	}
}

func WithBorderAll() ContainerOption {
	return WithBorder(true, true, true, true)
}

func WithBorderHorizontal() ContainerOption {
	return WithBorder(true, false, true, false)
}

func WithBorderVertical() ContainerOption {
	return WithBorder(false, true, false, true)
}

// Package-level border style variables to allow taking their address
var (
	roundedBorderVar = lipgloss.RoundedBorder()
	thickBorderVar   = lipgloss.ThickBorder()
	doubleBorderVar  = lipgloss.DoubleBorder()
)

func WithBorderStyle(style *lipgloss.Border) ContainerOption {
	return func(c *container) { c.borderStyle = *style }
}

func WithRoundedBorder() ContainerOption { return WithBorderStyle(&roundedBorderVar) }
func WithThickBorder() ContainerOption   { return WithBorderStyle(&thickBorderVar) }
func WithDoubleBorder() ContainerOption  { return WithBorderStyle(&doubleBorderVar) }

// Sizing options
func WithWidth(width int) ContainerOption {
	return func(c *container) {
		c.width = width
	}
}

func WithHeight(height int) ContainerOption {
	return func(c *container) {
		c.height = height
	}
}

func WithMinWidth(width int) ContainerOption {
	return func(c *container) {
		c.minWidth = width
	}
}

func WithMinHeight(height int) ContainerOption {
	return func(c *container) {
		c.minHeight = height
	}
}

func WithMaxWidth(width int) ContainerOption {
	return func(c *container) {
		c.maxWidth = width
	}
}

func WithMaxHeight(height int) ContainerOption {
	return func(c *container) {
		c.maxHeight = height
	}
}

// Accessibility options
func WithAriaLabel(label string) ContainerOption {
	return func(c *container) {
		c.ariaLabel = label
	}
}

// Debug options
func WithDebug(enable bool) ContainerOption {
	return func(c *container) {
		c.debug = enable
	}
}

// Transition options
func WithTransition(duration int) ContainerOption {
	return func(c *container) {
		c.transitionDuration = duration
	}
}

// Style options
func WithStyle(style *lipgloss.Style) ContainerOption {
	return func(c *container) {
		c.customStyle = style
	}
}

// Overlay options
func WithOverlay(f func(width, height int) string) ContainerOption {
	return func(c *container) {
		c.overlayFunc = f
	}
}

// State management options
func WithStateChangeHandler(handler func(ContainerState)) ContainerOption {
	return func(c *container) {
		c.onStateChange = handler
	}
}

// Enhanced Feature Options from original EnhancedContainer

// WithMinSize sets the minimum width and height for the container.
func WithMinSize(width, height int) ContainerOption {
	return func(c *container) {
		c.minWidth = width
		c.minHeight = height
	}
}

// WithMaxSize sets the maximum width and height for the container.
func WithMaxSize(width, height int) ContainerOption {
	return func(c *container) {
		c.maxWidth = width
		c.maxHeight = height
	}
}

// StringModel wraps a static string as a tea.Model for use in containers.
type stringModel struct {
	content string
}

func (s stringModel) Init() tea.Cmd                           { return nil }
func (s stringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s stringModel) View() string                            { return s.content }
func (s stringModel) SetSize(width, height int) tea.Cmd       { return nil }
func (s stringModel) GetSize() (width, height int)            { return 0, 0 }

// StringModel returns a tea.Model that renders the given string.
func StringModel(content string) tea.Model {
	return stringModel{content: content}
}

// EmptyModel returns a tea.Model that renders an empty string.
func EmptyModel() tea.Model {
	return stringModel{content: ""}
}

// Border and padding getters
func (c *container) GetBorderTop() bool    { return c.borderTop }
func (c *container) GetBorderRight() bool  { return c.borderRight }
func (c *container) GetBorderBottom() bool { return c.borderBottom }
func (c *container) GetBorderLeft() bool   { return c.borderLeft }
func (c *container) GetPaddingTop() int    { return c.paddingTop }
func (c *container) GetPaddingRight() int  { return c.paddingRight }
func (c *container) GetPaddingBottom() int { return c.paddingBottom }
func (c *container) GetPaddingLeft() int   { return c.paddingLeft }
