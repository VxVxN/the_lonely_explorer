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
	// Category groups records under a header in the list, e.g. "Флора" or
	// "Исследования станции". Records are grouped in the order their
	// category is first seen; an empty Category renders without a header.
	Category string
	Action   func()
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

	rows     []*simpleui.Clickable
	selected int

	running bool
}

const (
	itemHeight       = 50.0
	imageWidth       = 40.0
	descriptionWidth = 380.0
)

// NewJournal builds a Journal styled from theme, laid out against a
// screenWidth x screenHeight screen.
func NewJournal(theme *simpleui.Theme, screenWidth, screenHeight float64) *Journal {
	empty := simpleui.NewText("Журнал пуст", theme.Face, theme.MutedColor)
	empty.Align = simpleui.AlignCenter

	title := simpleui.NewText("Бортовой журнал", theme.TitleFace, theme.Accent)
	title.Align = simpleui.AlignCenter

	titleRule := simpleui.NewContainer(simpleui.Vertical)
	titleRule.Background = theme.Border
	titleRule.Padding = simpleui.Insets{Top: 1}

	header := simpleui.NewContainer(simpleui.Vertical)
	header.Spacing = 10
	header.Padding = simpleui.Insets{Bottom: 15}
	header.Add(title)
	header.Add(titleRule)

	list := simpleui.NewContainer(simpleui.Vertical)
	list.Spacing = 10

	panel := simpleui.NewContainer(simpleui.Vertical)
	panel.Background = simpleui.WithAlpha(theme.Panel, 235)
	panel.Border = theme.Border
	panel.BorderWidth = 1
	panel.Padding = simpleui.NewInsets(20)
	panel.Add(header)
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
	j.rows = nil
	j.selected = -1

	if len(records) == 0 {
		j.list.Add(j.empty)
		return
	}

	var categories []string
	grouped := make(map[string][]RecordJournal)
	for _, record := range records {
		if _, ok := grouped[record.Category]; !ok {
			categories = append(categories, record.Category)
		}
		grouped[record.Category] = append(grouped[record.Category], record)
	}

	for categoryIndex, category := range categories {
		if category != "" {
			header := simpleui.NewContainer(simpleui.Vertical)
			header.Spacing = 6
			if categoryIndex > 0 {
				header.Padding = simpleui.Insets{Top: 10}
			}
			header.Add(simpleui.NewText(category, j.theme.Face, j.theme.Accent))

			rule := simpleui.NewContainer(simpleui.Vertical)
			rule.Background = simpleui.WithAlpha(j.theme.Border, 150)
			rule.Padding = simpleui.Insets{Top: 1}
			header.Add(rule)

			j.list.Add(header)
		}
		for _, record := range grouped[category] {
			record := record
			description := strings.Split(record.Description, "\n")[0]

			descriptionText := simpleui.NewText(description, j.theme.Face, j.theme.TextColor)
			descriptionText.MaxWidth = descriptionWidth

			row := simpleui.NewContainer(simpleui.Horizontal)
			row.Spacing = 10
			row.CrossAlign = simpleui.AlignCenter
			row.Padding = simpleui.Insets{Top: 4, Bottom: 4, Left: 6, Right: 6}
			row.Add(simpleui.NewIcon(record.Image, imageWidth, itemHeight))
			row.Add(descriptionText)

			clickable := simpleui.NewClickable(row, j.theme, record.Action)
			j.rows = append(j.rows, clickable)
			j.list.Add(clickable)
		}
	}

	j.selected = 0
	j.highlightSelected()
}

// MoveSelection shifts the highlighted record by delta (e.g. -1 for the
// arrow-up key, +1 for arrow-down), wrapping around the ends of the list.
func (j *Journal) MoveSelection(delta int) {
	if len(j.rows) == 0 {
		return
	}
	j.selected = (j.selected + delta + len(j.rows)) % len(j.rows)
	j.highlightSelected()
}

// ActivateSelection runs the currently highlighted record's Action, e.g. in
// response to the Enter key.
func (j *Journal) ActivateSelection() {
	if j.selected < 0 || j.selected >= len(j.rows) {
		return
	}
	j.rows[j.selected].Activate()
}

func (j *Journal) highlightSelected() {
	for i, row := range j.rows {
		row.Selected = i == j.selected
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
