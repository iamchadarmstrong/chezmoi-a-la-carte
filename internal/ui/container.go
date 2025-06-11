package ui

import (
	"a-la-carte/internal/ui/core"

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

// Container wraps a tea.Model and provides border, padding, and theme support.
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
	GetSize() (int, int)
	// New methods for better theme integration
	SetFocused(focused bool)
	SetActive(active bool)
	SetHovered(hovered bool)
	GetState() ContainerState
	// New: View with context
	ViewWithContext(ctx *LayoutContext) string
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
	// New fields for state management
	state         ContainerState
	onStateChange func(ContainerState) // Optional: callback for state changes
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
		return c.customStyle.Width(c.width).Height(c.height)
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
func (c *container) applyStateStyle(style *lipgloss.Style, t core.Theme) lipgloss.Style {
	switch {
	case c.state.Focused:
		return style.Background(t.BackgroundFocused())
	case c.state.Active:
		return style.Background(t.BackgroundActive())
	default:
		return *style
	}
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

func (c *container) View() string {
	t := core.CurrentTheme()
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
	}

	// Render content
	inner := c.renderOverlayContent()
	return style.Render(inner)
}

// prepareChildContext creates a new LayoutContext for child components
func (c *container) prepareChildContext(ctx *LayoutContext) *LayoutContext {
	width := ctx.AvailableWidth
	height := ctx.AvailableHeight
	borderLeft := btoi(c.borderLeft)
	borderRight := btoi(c.borderRight)
	borderTop := btoi(c.borderTop)
	borderBottom := btoi(c.borderBottom)
	padLeft := c.paddingLeft
	padRight := c.paddingRight
	padTop := c.paddingTop
	padBottom := c.paddingBottom

	childCtx := &LayoutContext{
		AvailableWidth:  width - padLeft - padRight - borderLeft - borderRight,
		AvailableHeight: height - padTop - padBottom - borderTop - borderBottom,
		PaddingLeft:     padLeft,
		PaddingRight:    padRight,
		PaddingTop:      padTop,
		PaddingBottom:   padBottom,
		BorderLeft:      borderLeft,
		BorderRight:     borderRight,
		BorderTop:       borderTop,
		BorderBottom:    borderBottom,
		NestingLevel:    ctx.NestingLevel + 1,
	}

	if childCtx.AvailableWidth < 0 {
		childCtx.AvailableWidth = 0
	}
	if childCtx.AvailableHeight < 0 {
		childCtx.AvailableHeight = 0
	}

	return childCtx
}

// prepareContainerStyle creates and configures the container's style
func (c *container) prepareContainerStyle(ctx *LayoutContext) lipgloss.Style {
	t := core.CurrentTheme()
	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if c.customStyle != nil {
		style := *c.customStyle
		return style.Width(width).Height(height)
	}

	style := lipgloss.NewStyle()
	outerWidth := width

	// Apply border styles
	if c.borderTop || c.borderRight || c.borderBottom || c.borderLeft {
		style = c.applyBorderStyle(&style, t)
	}

	// Apply background based on state
	switch {
	case c.state.Focused:
		style = style.Background(t.BackgroundFocused())
	case c.state.Active:
		style = style.Background(t.BackgroundActive())
	}

	// Apply padding and dimensions
	return style.Width(outerWidth).
		Height(height).
		PaddingTop(c.paddingTop).
		PaddingRight(c.paddingRight).
		PaddingBottom(c.paddingBottom).
		PaddingLeft(c.paddingLeft)
}

// applyBorderStyle applies the border styling based on container state
func (c *container) applyBorderStyle(style *lipgloss.Style, t core.Theme) lipgloss.Style {
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

// renderContent renders the container's content with proper context
func (c *container) renderContent(childCtx *LayoutContext) string {
	if c.overlayFunc != nil {
		overlay := c.overlayFunc(childCtx.AvailableWidth, childCtx.AvailableHeight)
		if overlay != "" {
			w := childCtx.AvailableWidth
			h := childCtx.AvailableHeight
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
	}

	if vwc, ok := c.content.(interface{ ViewWithContext(*LayoutContext) string }); ok {
		return vwc.ViewWithContext(childCtx)
	}
	return c.content.View()
}

func (c *container) ViewWithContext(ctx *LayoutContext) string {
	childCtx := c.prepareChildContext(ctx)
	style := c.prepareContainerStyle(ctx)
	inner := c.renderContent(childCtx)
	return style.Render(inner)
}

// Helper to convert bool to int
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (c *container) SetSize(width, height int, ctx *LayoutContext) tea.Cmd {
	c.width = width
	c.height = height
	// Propagate size to content if possible
	if sizeable, ok := c.content.(interface {
		SetSize(int, int, *LayoutContext) tea.Cmd
	}); ok {
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
		return sizeable.SetSize(w, h, ctx)
	}
	return nil
}

func (c *container) GetSize() (width, height int) {
	return c.width, c.height
}

// ContainerOption configures a Container.
type ContainerOption func(*container)

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
func WithPaddingAll(p int) ContainerOption { return WithPadding(p, p, p, p) }
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
func WithBorderAll() ContainerOption        { return WithBorder(true, true, true, true) }
func WithBorderHorizontal() ContainerOption { return WithBorder(true, false, true, false) }
func WithBorderVertical() ContainerOption   { return WithBorder(false, true, false, true) }

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

// Add a WithWidth option to set the container width
func WithWidth(width int) ContainerOption {
	return func(c *container) {
		c.width = width
	}
}

// StringModel wraps a static string as a tea.Model for use in containers.
//
// # Usage
//
//	m := StringModel("Hello, world!")
//	c := NewContainer(m, ...)
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

// WithStyle allows injecting a custom lipgloss.Style for the container (overrides default logic)
func WithStyle(style *lipgloss.Style) ContainerOption {
	return func(c *container) {
		c.customStyle = style
	}
}

// WithOverlay allows setting an overlay function to render instead of content (e.g., for empty messages)
func WithOverlay(f func(width, height int) string) ContainerOption {
	return func(c *container) {
		c.overlayFunc = f
	}
}

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

// New container options for enhanced theme integration

// WithStateChangeHandler adds a callback for state changes
func WithStateChangeHandler(handler func(ContainerState)) ContainerOption {
	return func(c *container) {
		c.onStateChange = handler
	}
}

// WithSemanticStyle applies semantic styling based on container state
func WithSemanticStyle() ContainerOption {
	return func(c *container) {
		// This is handled in the View method now
	}
}
