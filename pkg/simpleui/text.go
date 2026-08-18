package simpleui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Text draws a (optionally wrapping) line or paragraph of text.
type Text struct {
	BaseWidget

	Label string
	Face  text.Face
	Color color.Color
	// Align is the horizontal alignment of the text within its box.
	Align Align
	// MaxWidth forces wrapping at a fixed width regardless of the space the
	// parent container offers. Leave zero to wrap at the container's width.
	MaxWidth float64
}

// NewText creates a Text widget. Long lines wrap to whatever width the
// parent Container gives them; set t.MaxWidth to force a fixed wrap width.
func NewText(label string, face text.Face, col color.Color) *Text {
	return &Text{Label: label, Face: face, Color: col, Align: AlignStart}
}

func (t *Text) lineSpacing() float64 {
	m := t.Face.Metrics()
	return m.HAscent + m.HDescent + m.HLineGap
}

func (t *Text) wrapWidth(maxWidth float64) float64 {
	if t.MaxWidth > 0 {
		return t.MaxWidth
	}
	return maxWidth
}

func (t *Text) wrapped(maxWidth float64) string {
	width := t.wrapWidth(maxWidth)
	if width <= 0 {
		return t.Label
	}

	var out []string
	for _, paragraph := range strings.Split(t.Label, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := words[0]
		for _, word := range words[1:] {
			trial := line + " " + word
			if text.Advance(trial, t.Face) > width {
				out = append(out, line)
				line = word
				continue
			}
			line = trial
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func (t *Text) PreferredSize(maxWidth float64) (w, h float64) {
	if t.Face == nil || t.Label == "" {
		return 0, 0
	}
	return text.Measure(t.wrapped(maxWidth), t.Face, t.lineSpacing())
}

func (t *Text) Update() {}

func (t *Text) Draw(screen *ebiten.Image) {
	if t.Face == nil || t.Label == "" {
		return
	}
	x, y, w, _ := t.Bounds()

	op := &text.DrawOptions{}
	op.LineSpacing = t.lineSpacing()
	op.PrimaryAlign = alignToText(t.Align)

	tx := x
	switch t.Align {
	case AlignCenter:
		tx = x + w/2
	case AlignEnd:
		tx = x + w
	}
	op.GeoM.Translate(tx, y)
	if t.Color != nil {
		op.ColorScale.ScaleWithColor(t.Color)
	}

	text.Draw(screen, t.wrapped(w), t.Face, op)
}

func alignToText(a Align) text.Align {
	switch a {
	case AlignCenter:
		return text.AlignCenter
	case AlignEnd:
		return text.AlignEnd
	default:
		return text.AlignStart
	}
}
