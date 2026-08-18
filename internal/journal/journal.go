package journal

import (
	"image/color"
	"strings"

	"github.com/VxVxN/the_lonely_explorer/pkg/simpleui"
	"github.com/hajimehoshi/ebiten/v2"
)

// RecordJournal is a single collected entry: an icon, its description (only
// the first line is shown in the list) and what to do when it's clicked
// (typically reopen the lore dialog).
type RecordJournal struct {
	Image       *ebiten.Image
	Description string
	Action      func()
}

// Journal is a toggleable panel listing collected records, anchored at a
// fixed screen position and auto-sized to its content.
type Journal struct {
	theme *simpleui.Theme

	ui    *simpleui.UI
	root  *simpleui.Container
	panel *simpleui.Container
	list  *simpleui.Container
	empty *simpleui.Text

	running bool
}

const (
	itemHeight = 50.0
	imageWidth = 40.0
)

// NewJournal builds a Journal styled from theme, laid out against a
// screenWidth x screenHeight screen.
func NewJournal(theme *simpleui.Theme, screenWidth, screenHeight float64) *Journal {
	empty := simpleui.NewText("Журнал пуст", theme.Face, theme.TextColor)

	list := simpleui.NewContainer(simpleui.Vertical)
	list.Spacing = 10

	panel := simpleui.NewContainer(simpleui.Vertical)
	panel.Background = color.RGBA{0, 0, 0, 200}
	panel.Padding = simpleui.NewInsets(15)
	panel.Add(list)

	root := simpleui.NewContainer(simpleui.Vertical)
	root.Padding = simpleui.Insets{Top: 100, Left: 100, Right: 50}
	root.Add(panel)

	j := &Journal{
		theme: theme,
		root:  root,
		panel: panel,
		list:  list,
		empty: empty,
	}
	j.ui = simpleui.New(root)
	j.ui.SetSize(screenWidth, screenHeight)
	j.SetKnowRecords(nil)
	return j
}

// SetKnowRecords replaces the list contents.
func (j *Journal) SetKnowRecords(records []RecordJournal) {
	j.list.Clear()

	if len(records) == 0 {
		j.list.Add(j.empty)
		return
	}

	for _, record := range records {
		record := record
		description := strings.Split(record.Description, "\n")[0]

		row := simpleui.NewContainer(simpleui.Horizontal)
		row.Spacing = 10
		row.CrossAlign = simpleui.AlignCenter
		row.Add(simpleui.NewIcon(record.Image, imageWidth, itemHeight))
		row.Add(simpleui.NewText(description, j.theme.Face, j.theme.TextColor))

		j.list.Add(simpleui.NewClickable(row, j.theme, record.Action))
	}
}

// SetPosition moves the panel's top-left corner.
func (j *Journal) SetPosition(x, y float64) {
	j.root.Padding.Left = x
	j.root.Padding.Top = y
}

// SetBackgroundColor sets the panel's fill color.
func (j *Journal) SetBackgroundColor(c color.Color) {
	j.panel.Background = c
}

func (j *Journal) TurnOnOff() {
	j.running = !j.running
}

func (j *Journal) TurnOff() {
	j.running = false
}

func (j *Journal) Update() {
	if !j.running {
		return
	}
	j.ui.Update()
}

func (j *Journal) Draw(screen *ebiten.Image) {
	if !j.running {
		return
	}
	j.ui.Draw(screen)
}
