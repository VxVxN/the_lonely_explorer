package simpleui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Checkbox is a labeled toggle. OnChange, if set, is called after Checked
// flips.
type Checkbox struct {
	BaseWidget

	Label    string
	Theme    *Theme
	Checked  bool
	OnChange func(checked bool)
	Disabled bool

	boxSize float64
	hovered bool
}

// NewCheckbox creates a checkbox styled from theme.
func NewCheckbox(label string, theme *Theme, onChange func(checked bool)) *Checkbox {
	return &Checkbox{Label: label, Theme: theme, OnChange: onChange, boxSize: 20}
}

func (c *Checkbox) PreferredSize(maxWidth float64) (w, h float64) {
	inner := maxWidth
	if inner > 0 {
		inner -= c.boxSize + c.Theme.Padding
		if inner < 0 {
			inner = 0
		}
	}
	t := &Text{Label: c.Label, Face: c.Theme.Face}
	tw, th := t.PreferredSize(inner)
	w = c.boxSize + c.Theme.Padding + tw
	h = th
	if c.boxSize > h {
		h = c.boxSize
	}
	return w, h
}

func (c *Checkbox) Update() {
	if c.Disabled {
		c.hovered = false
		return
	}
	cx, cy := ebiten.CursorPosition()
	c.hovered = c.Contains(float64(cx), float64(cy))
	if c.hovered && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		c.Checked = !c.Checked
		if c.OnChange != nil {
			c.OnChange(c.Checked)
		}
	}
}

func (c *Checkbox) Draw(screen *ebiten.Image) {
	x, y, _, h := c.Bounds()
	boxY := y + (h-c.boxSize)/2

	bg := c.Theme.InputBackground
	if c.hovered && !c.Disabled {
		bg = c.Theme.ButtonHover
	}
	vector.DrawFilledRect(screen, float32(x), float32(boxY), float32(c.boxSize), float32(c.boxSize), bg, false)
	vector.StrokeRect(screen, float32(x), float32(boxY), float32(c.boxSize), float32(c.boxSize), 1, c.Theme.Border, false)

	if c.Checked {
		mark := c.Theme.Accent
		if c.Disabled {
			mark = c.Theme.DisabledText
		}
		inset := c.boxSize * 0.25
		vector.DrawFilledRect(screen, float32(x+inset), float32(boxY+inset), float32(c.boxSize-inset*2), float32(c.boxSize-inset*2), mark, false)
	}

	textColor := c.Theme.TextColor
	if c.Disabled {
		textColor = c.Theme.DisabledText
	}
	label := &Text{Label: c.Label, Face: c.Theme.Face, Color: textColor}
	m := c.Theme.Face.Metrics()
	textH := m.HAscent + m.HDescent
	_, _, w, _ := c.Bounds()
	label.Layout(x+c.boxSize+c.Theme.Padding, y+(h-textH)/2, w-c.boxSize-c.Theme.Padding, textH)
	label.Draw(screen)
}
