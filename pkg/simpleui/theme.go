package simpleui

import (
	"image/color"

	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Theme is the one place styling lives. Widgets read colors/fonts straight
// from it instead of each owning its own set of nine-slice images, which is
// what made ebitenui's resources.go so large. Backgrounds are drawn as
// plain vector rectangles, so no image assets are required at all; set
// Panel/ButtonIdle/etc. to nil to fall back to a flat fill.
type Theme struct {
	Face      text.Face // default body font
	TitleFace text.Face // used by widgets that want a bigger face (falls back to Face if nil)

	TextColor    color.Color
	MutedColor   color.Color
	DisabledText color.Color

	Panel  color.Color // container/dialog background
	Border color.Color

	ButtonIdle     color.Color
	ButtonHover    color.Color
	ButtonPressed  color.Color
	ButtonDisabled color.Color
	ButtonText     color.Color

	InputBackground color.Color
	Accent          color.Color // checkbox mark, selection highlight, caret

	Padding float64 // default inner padding used by Button/Checkbox/Dialog
}

// DefaultTheme returns a dark theme close to the one the project already
// used with ebitenui, using face for everything unless overridden.
func DefaultTheme(face text.Face) *Theme {
	return &Theme{
		Face:      face,
		TitleFace: face,

		TextColor:    hexColor("dff4ff"),
		MutedColor:   hexColor("9fb7c6"),
		DisabledText: hexColor("5a7a91"),

		Panel:  hexColor("131a22"),
		Border: hexColor("2a3944"),

		ButtonIdle:     hexColor("22303c"),
		ButtonHover:    hexColor("2a3944"),
		ButtonPressed:  hexColor("4b687a"),
		ButtonDisabled: hexColor("1a232c"),
		ButtonText:     hexColor("dff4ff"),

		InputBackground: hexColor("1a232c"),
		Accent:          hexColor("e7c34b"),

		Padding: 12,
	}
}

// WithAlpha returns c with its alpha channel replaced by a (0 = fully
// transparent, 255 = fully opaque). Handy for translucent panel
// backgrounds, e.g. panel.Background = simpleui.WithAlpha(theme.Panel, 200).
func WithAlpha(c color.Color, a uint8) color.Color {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	n.A = a
	return n
}

func hexColor(hex string) color.Color {
	var r, g, b uint8
	for i := 0; i < 3; i++ {
		var v uint32
		for _, c := range hex[i*2 : i*2+2] {
			v <<= 4
			switch {
			case c >= '0' && c <= '9':
				v |= uint32(c - '0')
			case c >= 'a' && c <= 'f':
				v |= uint32(c-'a') + 10
			case c >= 'A' && c <= 'F':
				v |= uint32(c-'A') + 10
			}
		}
		switch i {
		case 0:
			r = uint8(v)
		case 1:
			g = uint8(v)
		case 2:
			b = uint8(v)
		}
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}
