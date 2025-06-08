package journal

import (
	"image/color"
	"strings"

	"golang.org/x/image/font"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Journal struct {
	isRunning bool
	font      font.Face
	position  struct {
		x, y float64
	}
	bgColor       color.RGBA
	knowRecords   []RecordJournal
	hoveredIndex  int // Index of the hovered point (-1 if nothing is hovered)
	pressedIndex  int
	itemHeight    float64
	padding       float64
	imageWidth    float64
	textOffsetX   float64
	cornerRadius  float64
	screenWidth   int
	hoverColor    color.RGBA
	selectedColor color.RGBA
}

type RecordJournal struct {
	Image       *ebiten.Image
	Description string
	Action      func()
}

func NewJournal(font font.Face) *Journal {
	return &Journal{
		font:    font,
		bgColor: color.RGBA{0, 0, 0, 200},
		position: struct{ x, y float64 }{
			x: 50,
			y: 50,
		},
		hoveredIndex:  -1,
		pressedIndex:  -1,
		itemHeight:    50,
		padding:       10,
		imageWidth:    40,
		textOffsetX:   50,
		cornerRadius:  5,
		hoverColor:    color.RGBA{50, 50, 50, 255},
		selectedColor: color.RGBA{100, 100, 100, 255},
	}
}

func (j *Journal) Draw(screen *ebiten.Image) {
	if !j.isRunning {
		return
	}

	j.screenWidth = screen.Bounds().Dx()
	width := j.screenWidth - int(j.position.x) - 50

	itemCount := len(j.knowRecords)
	if itemCount == 0 {
		itemCount = 1
	}
	bg := ebiten.NewImage(width, int(float64(itemCount)*(j.itemHeight+j.padding)+j.padding))
	bg.Fill(j.bgColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(j.position.x, j.position.y)
	screen.DrawImage(bg, op)

	for i := 0; i < len(j.knowRecords); i++ {
		yPos := j.position.y + j.padding + float64(i)*(j.itemHeight+j.padding)

		// Draw the background for the item if it is hovered over or selected
		if i == j.hoveredIndex || i == j.pressedIndex {
			itemBg := ebiten.NewImage(width-int(j.padding*2), int(j.itemHeight))
			if i == j.pressedIndex {
				itemBg.Fill(j.selectedColor)
			} else {
				itemBg.Fill(j.hoverColor)
			}
			itemOp := &ebiten.DrawImageOptions{}
			itemOp.GeoM.Translate(j.position.x+j.padding, yPos)
			screen.DrawImage(itemBg, itemOp)
		}

		imgOp := &ebiten.DrawImageOptions{}
		imgOp.GeoM.Scale(j.imageWidth/float64(j.knowRecords[i].Image.Bounds().Dx()), j.itemHeight/float64(j.knowRecords[i].Image.Bounds().Dy()))
		imgOp.GeoM.Translate(j.position.x+j.padding, yPos)
		screen.DrawImage(j.knowRecords[i].Image, imgOp)

		description := strings.Split(j.knowRecords[i].Description, "\n")[0]
		text.Draw(screen, description, j.font, int(j.position.x+j.padding+j.textOffsetX), int(yPos+j.itemHeight/2+5), color.White)
	}

	if len(j.knowRecords) == 0 {
		yPos := j.position.y + j.padding
		text.Draw(screen, "Журнал пуст", j.font, int(j.position.x+j.padding+j.textOffsetX), int(yPos+j.itemHeight/2+5), color.White)
	}
}

func (j *Journal) Update() {
	if !j.isRunning {
		return
	}

	j.hoveredIndex = -1

	cursorX, cursorY := ebiten.CursorPosition()
	cursorFX, cursorFY := float64(cursorX), float64(cursorY)

	// Check if the cursor is inside the journal
	if cursorFX >= j.position.x && cursorFX <= j.position.x+float64(j.screenWidth)-50 &&
		cursorFY >= j.position.y && cursorFY <= j.position.y+float64(len(j.knowRecords))*(j.itemHeight+j.padding)+j.padding {

		// Determining which point the cursor is on
		for i := 0; i < len(j.knowRecords); i++ {
			yPos := j.position.y + j.padding + float64(i)*(j.itemHeight+j.padding)
			if cursorFY >= yPos && cursorFY <= yPos+j.itemHeight {
				j.hoveredIndex = i
				break
			}
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if j.hoveredIndex != -1 {
				j.pressedIndex = j.hoveredIndex
				j.knowRecords[j.hoveredIndex].Action()
			}
		} else {
			j.pressedIndex = -1
		}
	}
}

func (j *Journal) TurnOnOff() {
	j.isRunning = !j.isRunning
	if !j.isRunning {
		j.hoveredIndex = -1
		j.pressedIndex = -1
	}
}

func (j *Journal) TurnOff() {
	j.isRunning = false
	j.hoveredIndex = -1
	j.pressedIndex = -1
}

func (j *Journal) SetPosition(x, y float64) {
	j.position.x = x
	j.position.y = y
}

func (j *Journal) SetBackgroundColor(c color.RGBA) {
	j.bgColor = c
}

func (j *Journal) SetKnowRecords(records []RecordJournal) {
	j.knowRecords = records
}

func (j *Journal) SetHoverColor(c color.RGBA) {
	j.hoverColor = c
}

func (j *Journal) SetSelectedColor(c color.RGBA) {
	j.selectedColor = c
}
