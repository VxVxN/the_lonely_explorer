package dialog

import (
	"image/color"

	"github.com/VxVxN/the_lonely_explorer/internal/ui"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type Dialog struct {
	ui        *ebitenui.UI
	textPanel *widget.Text
	isRunning bool
}

func NewDialog(res *ui.UiResources) *Dialog {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	textPanel := widget.NewText(
		widget.TextOpts.Text("",
			res.Text.SmallFace,
			res.Text.IdleColor),
		widget.TextOpts.MaxWidth(800),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceSimple(createPanelImage(), 1, 1)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(50)))),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	panel.AddChild(textPanel)
	rootContainer.AddChild(panel)
	rootContainer.AddChild(widget.NewText(
		widget.TextOpts.Text("Нажмите Enter для продолжения", res.Text.Face, res.Text.DisabledColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionEnd),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
				Padding:            widget.NewInsetsSimple(30),
			}),
		),
	))

	ui := &ebitenui.UI{
		Container: rootContainer,
	}

	return &Dialog{
		ui:        ui,
		textPanel: textPanel,
	}
}

func (d *Dialog) Draw(screen *ebiten.Image) {
	if !d.isRunning {
		return
	}
	d.ui.Draw(screen)
}
func (d *Dialog) Update() {
	if !d.isRunning {
		return
	}
	d.ui.Update()
}

func (d *Dialog) TurnOn(text string) {
	d.textPanel.Label = text
	d.isRunning = true
}

func (d *Dialog) TurnOff() {
	d.isRunning = false
}

func createPanelImage() *ebiten.Image {
	img := ebiten.NewImage(500, 500)
	img.Fill(color.RGBA{0, 0, 0, 180})
	return img
}
