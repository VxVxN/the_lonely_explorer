package simpleui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Direction is the main axis a Container lays its children out along.
type Direction int

const (
	Vertical Direction = iota
	Horizontal
)

// LayoutData controls how a single child behaves inside its Container.
type LayoutData struct {
	// Stretch makes the child grow to fill the remaining space on the
	// container's main axis, shared evenly with other stretching siblings.
	Stretch bool
	// CrossAlign overrides the container's default CrossAlign for this
	// child. Leave zero (AlignInherit) to use the container's default.
	CrossAlign Align
}

type containerChild struct {
	widget Widget
	data   LayoutData
}

// Container lays out its children in a single row or column, similar to a
// CSS flexbox with one axis. It is the only layout primitive in simpleui:
// build grids or centered dialogs by nesting containers and using Spacer to
// push content around.
type Container struct {
	BaseWidget

	Direction Direction
	Spacing   float64
	Padding   Insets
	// CrossAlign is the default alignment of children on the cross axis
	// (horizontal for a Vertical container, vertical for a Horizontal one).
	CrossAlign Align
	// Background, if non-nil, is painted behind the children.
	Background color.Color
	// Border, if non-nil, is stroked around the container's edge.
	Border      color.Color
	BorderWidth float64

	children []containerChild
}

// NewContainer creates an empty container with sensible defaults
// (children stretch to fill the cross axis).
func NewContainer(direction Direction) *Container {
	return &Container{
		Direction:  direction,
		CrossAlign: AlignStretch,
	}
}

// Add appends a child widget, optionally with LayoutData. It returns the
// container so calls can be chained.
func (c *Container) Add(w Widget, data ...LayoutData) *Container {
	var d LayoutData
	if len(data) > 0 {
		d = data[0]
	}
	c.children = append(c.children, containerChild{widget: w, data: d})
	return c
}

// Clear removes all children.
func (c *Container) Clear() {
	c.children = nil
}

func (c *Container) Layout(x, y, w, h float64) {
	c.BaseWidget.Layout(x, y, w, h)

	innerX := x + c.Padding.Left
	innerY := y + c.Padding.Top
	innerW := w - c.Padding.Left - c.Padding.Right
	innerH := h - c.Padding.Top - c.Padding.Bottom
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	if len(c.children) == 0 {
		return
	}

	horizontal := c.Direction == Horizontal

	mainLen, crossLen := innerH, innerW
	if horizontal {
		mainLen, crossLen = innerW, innerH
	}

	// Widths are the only dimension text wraps against, so only pass a
	// wrap hint to children of a Vertical container.
	wrapHint := 0.0
	if !horizontal {
		wrapHint = crossLen
	}

	prefMain := make([]float64, len(c.children))
	prefCross := make([]float64, len(c.children))
	totalFixed := 0.0
	stretchCount := 0
	for i, ch := range c.children {
		pw, ph := ch.widget.PreferredSize(wrapHint)
		if horizontal {
			prefMain[i], prefCross[i] = pw, ph
		} else {
			prefMain[i], prefCross[i] = ph, pw
		}
		if ch.data.Stretch {
			stretchCount++
			continue
		}
		totalFixed += prefMain[i]
	}

	totalSpacing := c.Spacing * float64(len(c.children)-1)
	remaining := mainLen - totalFixed - totalSpacing
	stretchSize := 0.0
	if stretchCount > 0 && remaining > 0 {
		stretchSize = remaining / float64(stretchCount)
	}

	pos := 0.0
	if horizontal {
		pos = innerX
	} else {
		pos = innerY
	}

	for i, ch := range c.children {
		mainSize := prefMain[i]
		if ch.data.Stretch {
			mainSize = stretchSize
		}
		if mainSize < 0 {
			mainSize = 0
		}

		align := ch.data.CrossAlign
		if align == AlignInherit {
			align = c.CrossAlign
		}

		crossSize := prefCross[i]
		if align == AlignStretch || crossSize > crossLen {
			crossSize = crossLen
		}
		crossOffset := 0.0
		switch align {
		case AlignCenter:
			crossOffset = (crossLen - crossSize) / 2
		case AlignEnd:
			crossOffset = crossLen - crossSize
		}
		if crossOffset < 0 {
			crossOffset = 0
		}

		var cx, cy, cw, ch2 float64
		if horizontal {
			cx, cy, cw, ch2 = pos, innerY+crossOffset, mainSize, crossSize
		} else {
			cx, cy, cw, ch2 = innerX+crossOffset, pos, crossSize, mainSize
		}
		ch.widget.Layout(cx, cy, cw, ch2)

		pos += mainSize + c.Spacing
	}
}

func (c *Container) PreferredSize(maxWidth float64) (w, h float64) {
	if len(c.children) == 0 {
		return c.Padding.Left + c.Padding.Right, c.Padding.Top + c.Padding.Bottom
	}

	horizontal := c.Direction == Horizontal
	innerMaxWidth := maxWidth
	if innerMaxWidth > 0 {
		innerMaxWidth -= c.Padding.Left + c.Padding.Right
		if innerMaxWidth < 0 {
			innerMaxWidth = 0
		}
	}
	wrapHint := 0.0
	if !horizontal {
		wrapHint = innerMaxWidth
	}

	mainTotal := 0.0
	crossMax := 0.0
	for i, ch := range c.children {
		pw, ph := ch.widget.PreferredSize(wrapHint)
		main, cross := ph, pw
		if horizontal {
			main, cross = pw, ph
		}
		mainTotal += main
		if cross > crossMax {
			crossMax = cross
		}
		if i > 0 {
			mainTotal += c.Spacing
		}
	}

	if horizontal {
		return mainTotal + c.Padding.Left + c.Padding.Right, crossMax + c.Padding.Top + c.Padding.Bottom
	}
	return crossMax + c.Padding.Left + c.Padding.Right, mainTotal + c.Padding.Top + c.Padding.Bottom
}

func (c *Container) Update() {
	for _, ch := range c.children {
		ch.widget.Update()
	}
}

func (c *Container) Draw(screen *ebiten.Image) {
	x, y, w, h := c.Bounds()
	if c.Background != nil && w > 0 && h > 0 {
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), c.Background, false)
	}
	if c.Border != nil && w > 0 && h > 0 {
		bw := c.BorderWidth
		if bw <= 0 {
			bw = 1
		}
		vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(bw), c.Border, false)
	}
	for _, ch := range c.children {
		ch.widget.Draw(screen)
	}
}

// Spacer is an invisible widget that only exists to soak up stretch space,
// e.g. to center a widget by placing a Spacer before and after it in a
// container.
type Spacer struct {
	BaseWidget
}

func NewSpacer() *Spacer { return &Spacer{} }

func (s *Spacer) PreferredSize(maxWidth float64) (w, h float64) { return 0, 0 }
func (s *Spacer) Update()                                       {}
func (s *Spacer) Draw(screen *ebiten.Image)                     {}
