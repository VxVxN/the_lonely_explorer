package simpleui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Dropdown is a button that opens a popup List to pick one entry from,
// i.e. what ebitenui calls a ListComboButton.
type Dropdown struct {
	BaseWidget

	Theme    *Theme
	Entries  []string
	Selected int
	OnSelect func(index int)

	open   bool
	button *Button
	list   *List
}

// NewDropdown creates a dropdown styled from theme, initially showing
// entries[selected].
func NewDropdown(entries []string, selected int, theme *Theme, onSelect func(index int)) *Dropdown {
	d := &Dropdown{
		Theme:    theme,
		Entries:  entries,
		Selected: selected,
		OnSelect: onSelect,
	}
	d.button = NewButton("", theme, func() { d.open = !d.open })
	d.list = NewList(entries, theme, nil)
	d.list.Selected = selected
	d.list.OnSelect = func(index int) {
		d.Selected = index
		d.open = false
		if d.OnSelect != nil {
			d.OnSelect(index)
		}
	}
	return d
}

func (d *Dropdown) label() string {
	if d.Selected >= 0 && d.Selected < len(d.Entries) {
		return d.Entries[d.Selected] + "  ▾"
	}
	return "▾"
}

func (d *Dropdown) PreferredSize(maxWidth float64) (w, h float64) {
	d.button.Label = d.label()
	return d.button.PreferredSize(maxWidth)
}

func (d *Dropdown) Layout(x, y, w, h float64) {
	d.BaseWidget.Layout(x, y, w, h)
	d.button.Layout(x, y, w, h)

	listH := d.list.ItemHeight * float64(min(len(d.Entries), 6))
	d.list.Height = listH
	d.list.Layout(x, y+h, w, listH)
}

func (d *Dropdown) Update() {
	d.button.Label = d.label()
	d.button.Update()

	if !d.open {
		return
	}
	d.list.Update() // may close d.open via d.list.OnSelect above

	if d.open && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		fx, fy := float64(cx), float64(cy)
		if !d.button.Contains(fx, fy) && !d.list.Contains(fx, fy) {
			d.open = false
		}
	}
}

func (d *Dropdown) Draw(screen *ebiten.Image) {
	d.button.Draw(screen)
	if d.open {
		d.list.Draw(screen)
	}
}
