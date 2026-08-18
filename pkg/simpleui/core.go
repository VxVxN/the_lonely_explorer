// Package simpleui is a small, no-frills UI toolkit for Ebiten games.
//
// It replaces the ebitenui dependency for this project with a much smaller
// surface area: a handful of widgets (Container, Text, Button, Checkbox,
// List, Dropdown, TextArea) laid out with a single flexbox-like Container,
// styled through one flat Theme struct, and drawn with plain vector shapes
// instead of nine-slice images.
//
// Typical usage:
//
//	theme := simpleui.DefaultTheme(face)
//	root := simpleui.NewContainer(simpleui.Vertical)
//	root.Add(simpleui.NewText("Hello", theme.Face, theme.TextColor))
//	root.Add(simpleui.NewButton("Click me", theme, func() { ... }))
//
//	ui := simpleui.New(root)
//	ui.SetSize(screenWidth, screenHeight)
//
//	// each frame:
//	ui.Update()
//	ui.Draw(screen)
package simpleui

import "github.com/hajimehoshi/ebiten/v2"

// Align controls how a widget is positioned along a container's cross axis.
type Align int

const (
	// AlignInherit means "use the parent Container's default CrossAlign".
	AlignInherit Align = iota
	AlignStart
	AlignCenter
	AlignEnd
	AlignStretch
)

// Insets is padding/margin around a rectangle.
type Insets struct {
	Top, Right, Bottom, Left float64
}

// NewInsets returns Insets with the same value on all four sides.
func NewInsets(v float64) Insets {
	return Insets{Top: v, Right: v, Bottom: v, Left: v}
}

// Widget is anything that can be laid out, updated and drawn.
type Widget interface {
	// Layout is called by a parent Container to assign this widget's
	// position and size.
	Layout(x, y, w, h float64)
	// Bounds returns the rectangle assigned by the last Layout call.
	Bounds() (x, y, w, h float64)
	// PreferredSize returns the size this widget would like to have.
	// maxWidth is a hint for wrapping (0 means "unconstrained") and may be
	// ignored by widgets that don't wrap.
	PreferredSize(maxWidth float64) (w, h float64)
	Update()
	Draw(screen *ebiten.Image)
}

// BaseWidget stores the rectangle assigned by Layout. Embed it in concrete
// widgets to get Bounds/Layout/Contains for free.
type BaseWidget struct {
	x, y, w, h float64
}

func (b *BaseWidget) Layout(x, y, w, h float64) {
	b.x, b.y, b.w, b.h = x, y, w, h
}

func (b *BaseWidget) Bounds() (x, y, w, h float64) {
	return b.x, b.y, b.w, b.h
}

// Contains reports whether the point (px, py) is inside the widget's bounds.
func (b *BaseWidget) Contains(px, py float64) bool {
	return px >= b.x && px < b.x+b.w && py >= b.y && py < b.y+b.h
}

// UI is the root of a widget tree. Call SetSize once (or whenever the
// screen size changes), then Update and Draw once per frame.
type UI struct {
	Root          Widget
	width, height float64
}

// New creates a UI rooted at the given widget (usually a *Container).
func New(root Widget) *UI {
	return &UI{Root: root}
}

// SetSize sets the screen size the UI should lay itself out against.
func (u *UI) SetSize(w, h float64) {
	u.width, u.height = w, h
}

func (u *UI) layout() {
	if u.Root == nil || u.width <= 0 || u.height <= 0 {
		return
	}
	u.Root.Layout(0, 0, u.width, u.height)
}

// Update lays out the tree and dispatches input to it. Call once per frame.
func (u *UI) Update() {
	u.layout()
	if u.Root == nil {
		return
	}
	u.Root.Update()
}

// Draw lays out the tree and draws it. Call once per frame.
func (u *UI) Draw(screen *ebiten.Image) {
	u.layout()
	if u.Root == nil {
		return
	}
	u.Root.Draw(screen)
}
