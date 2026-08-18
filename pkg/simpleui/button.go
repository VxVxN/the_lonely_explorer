package simpleui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Button is a clickable label. Just set Label/OnClick/Disabled; colors come
// from the Theme unless overridden per-widget.
type Button struct {
	BaseWidget

	Label    string
	Theme    *Theme
	OnClick  func()
	Disabled bool

	// Width/Height force a fixed size instead of sizing to the label.
	Width, Height float64

	hovered bool
	pressed bool
}

// NewButton creates a button styled from theme.
func NewButton(label string, theme *Theme, onClick func()) *Button {
	return &Button{Label: label, Theme: theme, OnClick: onClick}
}

func (b *Button) PreferredSize(maxWidth float64) (w, h float64) {
	pad := b.Theme.Padding
	tw, th := text.Advance(b.Label, b.Theme.Face), 0.0
	m := b.Theme.Face.Metrics()
	th = m.HAscent + m.HDescent
	w, h = tw+pad*2, th+pad
	if b.Width > 0 {
		w = b.Width
	}
	if b.Height > 0 {
		h = b.Height
	}
	return w, h
}

func (b *Button) Update() {
	if b.Disabled {
		b.hovered, b.pressed = false, false
		return
	}
	cx, cy := ebiten.CursorPosition()
	b.hovered = b.Contains(float64(cx), float64(cy))

	if b.hovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.pressed = true
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if b.pressed && b.hovered && b.OnClick != nil {
			b.OnClick()
		}
		b.pressed = false
	}
}

func (b *Button) bgColor() color.Color {
	switch {
	case b.Disabled:
		return b.Theme.ButtonDisabled
	case b.pressed:
		return b.Theme.ButtonPressed
	case b.hovered:
		return b.Theme.ButtonHover
	default:
		return b.Theme.ButtonIdle
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	x, y, w, h := b.Bounds()
	if bg := b.bgColor(); bg != nil {
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bg, false)
	}

	textColor := b.Theme.ButtonText
	if b.Disabled {
		textColor = b.Theme.DisabledText
	}

	t := &Text{Label: b.Label, Face: b.Theme.Face, Color: textColor, Align: AlignCenter}
	m := b.Theme.Face.Metrics()
	textH := m.HAscent + m.HDescent
	t.Layout(x, y+(h-textH)/2, w, textH)
	t.Draw(screen)
}
