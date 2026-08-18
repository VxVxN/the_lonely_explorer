package simpleui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// List shows a scrollable, selectable list of strings. Scroll with the
// mouse wheel, click an entry to select it.
type List struct {
	BaseWidget

	Theme      *Theme
	Entries    []string
	Selected   int // -1 for no selection
	OnSelect   func(index int)
	ItemHeight float64
	// Height is the fixed viewport height. Defaults to 5*ItemHeight.
	Height float64

	scroll       float64
	hoveredIndex int
}

// NewList creates a list styled from theme.
func NewList(entries []string, theme *Theme, onSelect func(index int)) *List {
	return &List{
		Entries:      entries,
		Theme:        theme,
		OnSelect:     onSelect,
		Selected:     -1,
		ItemHeight:   32,
		hoveredIndex: -1,
	}
}

func (l *List) viewportHeight() float64 {
	if l.Height > 0 {
		return l.Height
	}
	return l.ItemHeight * 5
}

func (l *List) contentHeight() float64 {
	return float64(len(l.Entries)) * l.ItemHeight
}

func (l *List) maxScroll() float64 {
	m := l.contentHeight() - l.viewportHeight()
	if m < 0 {
		return 0
	}
	return m
}

func (l *List) PreferredSize(maxWidth float64) (w, h float64) {
	return maxWidth, l.viewportHeight()
}

func (l *List) Update() {
	x, y, w, h := l.Bounds()
	cx, cy := ebiten.CursorPosition()
	fx, fy := float64(cx), float64(cy)

	inBounds := fx >= x && fx < x+w && fy >= y && fy < y+h
	l.hoveredIndex = -1
	if inBounds {
		if _, dy := ebiten.Wheel(); dy != 0 {
			l.scroll -= dy * l.ItemHeight
		}
		idx := int((fy - y + l.scroll) / l.ItemHeight)
		if idx >= 0 && idx < len(l.Entries) {
			l.hoveredIndex = idx
		}
	}
	if l.scroll < 0 {
		l.scroll = 0
	}
	if l.scroll > l.maxScroll() {
		l.scroll = l.maxScroll()
	}

	if inBounds && l.hoveredIndex >= 0 && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		l.Selected = l.hoveredIndex
		if l.OnSelect != nil {
			l.OnSelect(l.Selected)
		}
	}
}

func (l *List) Draw(screen *ebiten.Image) {
	x, y, w, h := l.Bounds()
	if w <= 0 || h <= 0 {
		return
	}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), l.Theme.InputBackground, false)

	clip, ok := screen.SubImage(image.Rect(int(x), int(y), int(x+w), int(y+h))).(*ebiten.Image)
	if !ok {
		return
	}

	first := int(l.scroll / l.ItemHeight)
	if first < 0 {
		first = 0
	}
	last := int((l.scroll+h)/l.ItemHeight) + 1
	if last > len(l.Entries) {
		last = len(l.Entries)
	}

	textColor := l.Theme.TextColor
	m := l.Theme.Face.Metrics()
	textH := m.HAscent + m.HDescent

	for i := first; i < last; i++ {
		itemY := y - l.scroll + float64(i)*l.ItemHeight

		var bg color.Color
		switch {
		case i == l.Selected:
			bg = l.Theme.ButtonPressed
		case i == l.hoveredIndex:
			bg = l.Theme.ButtonHover
		}
		if bg != nil {
			vector.DrawFilledRect(clip, float32(x), float32(itemY), float32(w), float32(l.ItemHeight), bg, false)
		}

		t := &Text{Label: l.Entries[i], Face: l.Theme.Face, Color: textColor}
		t.Layout(x+l.Theme.Padding, itemY+(l.ItemHeight-textH)/2, w-l.Theme.Padding*2, textH)
		t.Draw(clip)
	}

	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, l.Theme.Border, false)

	if l.maxScroll() > 0 {
		trackW := 4.0
		barH := h * (h / l.contentHeight())
		barY := y + (h-barH)*(l.scroll/l.maxScroll())
		vector.DrawFilledRect(screen, float32(x+w-trackW), float32(barY), float32(trackW), float32(barH), l.Theme.Accent, false)
	}
}
