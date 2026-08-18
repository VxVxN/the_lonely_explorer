package dialog

import (
	"github.com/VxVxN/the_lonely_explorer/pkg/simpleui"
	"github.com/hajimehoshi/ebiten/v2"
)

// Dialog is a centered modal text box with a "press enter to continue"
// hint, used for in-game lore/description popups.
type Dialog struct {
	dialog *simpleui.Dialog
}

func NewDialog(theme *simpleui.Theme) *Dialog {
	return &Dialog{dialog: simpleui.NewDialog(theme)}
}

func (d *Dialog) Draw(screen *ebiten.Image) {
	d.dialog.Draw(screen)
}

func (d *Dialog) Update() {
	d.dialog.Update()
}

func (d *Dialog) TurnOn(text string) {
	d.dialog.TurnOn(text)
}

func (d *Dialog) TurnOff() {
	d.dialog.TurnOff()
}
