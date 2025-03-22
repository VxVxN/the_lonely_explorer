package game

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/VxVxN/the_lonely_explorer/internal/ui"
)

type scene1UI struct {
	widget widget.PreferredSizeLocateableWidget
	ui     *ebitenui.UI
}

func newScene1UI(res *ui.UiResources) *scene1UI {
	container := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(100)),
		)),
	)

	container.AddChild(widget.NewText(
		widget.TextOpts.Text("Внимание, исследовательский модуль RX-7. Это Центр управления миссией на Земле. Вы успешно доставлены на поверхность планеты Kepler-452b. Ваша основная задача — исследование и анализ окружающей среды. Соберите данные о геологии, атмосфере и возможных признаках жизни. Будьте осторожны: планета мало изучена, и мы не можем предсказать все угрозы. Поддерживайте связь, передавайте информацию и следуйте протоколам безопасности. Удачи, RX-7. Земля с вами. Конец связи.", res.Text.Face, res.Text.IdleColor),
		//widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.MaxWidth(800),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	container.AddChild(widget.NewText(
		widget.TextOpts.Text("Нажмите Enter для продолжения", res.Text.Face, res.Text.DisabledColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
			}),
		),
	))

	return &scene1UI{
		widget: container,
		ui:     &ebitenui.UI{Container: container},
	}
}
