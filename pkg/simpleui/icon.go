package simpleui

import "github.com/hajimehoshi/ebiten/v2"

// Icon draws an image scaled to a fixed size. Use it for buttons/list rows
// that need a picture next to text.
type Icon struct {
	BaseWidget

	Image         *ebiten.Image
	Width, Height float64
}

// NewIcon creates an icon that scales img to (w, h).
func NewIcon(img *ebiten.Image, w, h float64) *Icon {
	return &Icon{Image: img, Width: w, Height: h}
}

func (i *Icon) PreferredSize(maxWidth float64) (w, h float64) {
	return i.Width, i.Height
}

func (i *Icon) Update() {}

func (i *Icon) Draw(screen *ebiten.Image) {
	if i.Image == nil {
		return
	}
	x, y, w, h := i.Bounds()
	if w <= 0 || h <= 0 {
		return
	}
	b := i.Image.Bounds()
	sw, sh := float64(b.Dx()), float64(b.Dy())
	if sw == 0 || sh == 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/sw, h/sh)
	op.GeoM.Translate(x, y)
	screen.DrawImage(i.Image, op)
}
