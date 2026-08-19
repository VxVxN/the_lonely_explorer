package game

import (
	"github.com/VxVxN/the_lonely_explorer/pkg/simpleui"
)

type scene1UI struct {
	ui *simpleui.UI
}

func newScene1UI(theme *simpleui.Theme, screenWidth, screenHeight float64) *scene1UI {
	body := simpleui.NewText("Внимание, исследовательский модуль RX-7. Говорит Центр управления миссией, Земля. Ты успешно доставлен на поверхность Kepler-452b — последнего кандидата в списке пригодных для колонизации планет. Задача: определить, можно ли здесь жить. Исследуй грунт, атмосферу, флору и фауну, фиксируй находки в бортовом журнале. Оружия у тебя нет и не будет — ты наблюдатель, не боец. Обнаружишь угрозу — не вступай в контакт: отступи, обойди, запиши и двигайся дальше. Сигнал с Земли идёт до тебя больше десяти минут в одну сторону, так что дальше рассчитывай только на себя. Передавай данные, береги себя. Земля ждёт твоего отчёта. Конец связи.", theme.Face, theme.TextColor)
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
