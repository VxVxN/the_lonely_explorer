package simpleui

import "github.com/hajimehoshi/ebiten/v2"

// Dialog is a centered modal text panel with a "press enter to continue"
// hint pinned to the bottom of the screen. It is a ready-made screen, not
// just a widget: create one, then TurnOn/TurnOff/Draw it like any other
// screen-level component. It sizes itself to the screen passed to Draw.
type Dialog struct {
	ui      *UI
	text    *Text
	running bool
}

// NewDialog builds a Dialog styled from theme.
func NewDialog(theme *Theme) *Dialog {
	body := NewText("", theme.Face, theme.TextColor)
	body.Align = AlignCenter
	body.MaxWidth = 800

	panel := NewContainer(Vertical)
	panel.Background = WithAlpha(theme.Panel, 200)
	panel.Border = theme.Border
	panel.BorderWidth = 1
	panel.Padding = NewInsets(50)
	panel.CrossAlign = AlignCenter
	panel.Add(body)

	panelRow := NewContainer(Horizontal)
	panelRow.Add(NewSpacer(), LayoutData{Stretch: true})
	panelRow.Add(panel)
	panelRow.Add(NewSpacer(), LayoutData{Stretch: true})

	hintWrap := NewContainer(Vertical)
	hintWrap.Padding = Insets{Bottom: 30}
	hint := NewText("Нажмите Enter для продолжения", theme.Face, theme.MutedColor)
	hint.Align = AlignCenter
	hintWrap.Add(hint)

	root := NewContainer(Vertical)
	root.Add(NewSpacer(), LayoutData{Stretch: true})
	root.Add(panelRow)
	root.Add(NewSpacer(), LayoutData{Stretch: true})
	root.Add(hintWrap)

	return &Dialog{ui: New(root), text: body}
}

func (d *Dialog) Update() {
	if !d.running {
		return
	}
	d.ui.Update()
}

func (d *Dialog) Draw(screen *ebiten.Image) {
	if !d.running {
		return
	}
	b := screen.Bounds()
	d.ui.SetSize(float64(b.Dx()), float64(b.Dy()))
	d.ui.Draw(screen)
}

// TurnOn shows the dialog with the given text.
func (d *Dialog) TurnOn(text string) {
	d.text.Label = text
	d.running = true
}

// TurnOff hides the dialog.
func (d *Dialog) TurnOff() {
	d.running = false
}
