package simpleui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TextArea is a read-only, mouse-wheel-scrollable block of wrapped text.
// Use it for long descriptions, logs or credits that don't fit on screen.
type TextArea struct {
	BaseWidget

	Theme *Theme
	Text  string
	// Height is the fixed viewport height. Defaults to 200.
	Height float64

	scroll float64
}

// NewTextArea creates a text area styled from theme.
func NewTextArea(content string, theme *Theme) *TextArea {
	return &TextArea{Text: content, Theme: theme, Height: 200}
}

func (t *TextArea) inner() *Text {
	return &Text{Label: t.Text, Face: t.Theme.Face, Color: t.Theme.TextColor}
}

func (t *TextArea) contentHeight(wrapWidth float64) float64 {
	_, h := t.inner().PreferredSize(wrapWidth)
	return h
}

func (t *TextArea) maxScroll() float64 {
	_, _, w, h := t.Bounds()
	m := t.contentHeight(w-t.Theme.Padding*2) - h
	if m < 0 {
		return 0
	}
	return m
}

func (t *TextArea) PreferredSize(maxWidth float64) (w, h float64) {
	if t.Height > 0 {
		return maxWidth, t.Height
	}
	return maxWidth, 200
}

func (t *TextArea) Update() {
	x, y, w, h := t.Bounds()
	cx, cy := ebiten.CursorPosition()
	fx, fy := float64(cx), float64(cy)
	if fx >= x && fx < x+w && fy >= y && fy < y+h {
		if _, dy := ebiten.Wheel(); dy != 0 {
			t.scroll -= dy * 40
		}
	}
	if t.scroll < 0 {
		t.scroll = 0
	}
	if max := t.maxScroll(); t.scroll > max {
		t.scroll = max
	}
}

func (t *TextArea) Draw(screen *ebiten.Image) {
	x, y, w, h := t.Bounds()
	if w <= 0 || h <= 0 {
		return
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), t.Theme.InputBackground, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, t.Theme.Border, false)

	clip, ok := screen.SubImage(image.Rect(int(x), int(y), int(x+w), int(y+h))).(*ebiten.Image)
	if !ok {
		return
	}

	pad := t.Theme.Padding
	content := t.inner()
	content.Layout(x+pad, y+pad-t.scroll, w-pad*2, t.contentHeight(w-pad*2))
	content.Draw(clip)

	if t.maxScroll() > 0 {
		trackW := 4.0
		total := t.contentHeight(w - pad*2)
		barH := h * (h / total)
		barY := y + (h-barH)*(t.scroll/t.maxScroll())
		vector.DrawFilledRect(screen, float32(x+w-trackW), float32(barY), float32(trackW), float32(barH), t.Theme.Accent, false)
	}
}
