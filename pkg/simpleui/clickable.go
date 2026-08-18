package simpleui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Clickable wraps any widget with hover/press/click handling and an
// optional highlight background, e.g. an icon+text row that should react
// like a button without being a Button (which only supports a text label).
type Clickable struct {
	BaseWidget

	Child    Widget
	Theme    *Theme
	OnClick  func()
	Disabled bool

	hovered bool
	pressed bool
}

// NewClickable wraps child so it reacts to hover/click like a button.
func NewClickable(child Widget, theme *Theme, onClick func()) *Clickable {
	return &Clickable{Child: child, Theme: theme, OnClick: onClick}
}

func (c *Clickable) PreferredSize(maxWidth float64) (w, h float64) {
	return c.Child.PreferredSize(maxWidth)
}

func (c *Clickable) Layout(x, y, w, h float64) {
	c.BaseWidget.Layout(x, y, w, h)
	c.Child.Layout(x, y, w, h)
}

func (c *Clickable) Update() {
	if c.Disabled {
		c.hovered, c.pressed = false, false
		c.Child.Update()
		return
	}

	cx, cy := ebiten.CursorPosition()
	c.hovered = c.Contains(float64(cx), float64(cy))

	if c.hovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		c.pressed = true
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if c.pressed && c.hovered && c.OnClick != nil {
			c.OnClick()
		}
		c.pressed = false
	}

	c.Child.Update()
}

func (c *Clickable) Draw(screen *ebiten.Image) {
	x, y, w, h := c.Bounds()

	var bg color.Color
	switch {
	case c.pressed:
		bg = c.Theme.ButtonPressed
	case c.hovered:
		bg = c.Theme.ButtonHover
	}
	if bg != nil {
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bg, false)
	}

	c.Child.Draw(screen)
}
