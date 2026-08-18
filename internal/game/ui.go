package game

import (
	"github.com/VxVxN/the_lonely_explorer/pkg/simpleui"
)

type scene1UI struct {
	ui *simpleui.UI
}

func newScene1UI(theme *simpleui.Theme, screenWidth, screenHeight float64) *scene1UI {
	body := simpleui.NewText("Внимание, исследовательский модуль RX-7. Это Центр управления миссией на Земле. Вы успешно доставлены на поверхность планеты Kepler-452b. Ваша основная задача — исследование и анализ окружающей среды. Соберите данные о геологии, атмосфере и возможных признаках жизни. Будьте осторожны: планета мало изучена, и мы не можем предсказать все угрозы. Поддерживайте связь, передавайте информацию и следуйте протоколам безопасности. Удачи, RX-7. Земля с вами. Конец связи.", theme.Face, theme.TextColor)
	body.Align = simpleui.AlignCenter
	body.MaxWidth = 800

	bodyRow := simpleui.NewContainer(simpleui.Horizontal)
	bodyRow.Add(simpleui.NewSpacer(), simpleui.LayoutData{Stretch: true})
	bodyRow.Add(body)
	bodyRow.Add(simpleui.NewSpacer(), simpleui.LayoutData{Stretch: true})

	hint := simpleui.NewText("Нажмите Enter для продолжения", theme.Face, theme.MutedColor)
	hint.Align = simpleui.AlignCenter

	root := simpleui.NewContainer(simpleui.Vertical)
	root.Padding = simpleui.NewInsets(100)
	root.Add(simpleui.NewSpacer(), simpleui.LayoutData{Stretch: true})
	root.Add(bodyRow)
	root.Add(simpleui.NewSpacer(), simpleui.LayoutData{Stretch: true})
	root.Add(hint)

	ui := simpleui.New(root)
	ui.SetSize(screenWidth, screenHeight)

	return &scene1UI{ui: ui}
}
