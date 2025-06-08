package game

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/VxVxN/gamedevlib/animation"
	keyeventmanager "github.com/VxVxN/gamedevlib/eventmanager"
	"github.com/VxVxN/gamedevlib/rectangle"
	"github.com/VxVxN/the_lonely_explorer/internal/eventmanager"
	"github.com/VxVxN/the_lonely_explorer/internal/journal"
	"github.com/VxVxN/the_lonely_explorer/pkg/dialog"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"

	_map "github.com/VxVxN/the_lonely_explorer/internal/map"
	"github.com/VxVxN/the_lonely_explorer/internal/stager"
	"github.com/VxVxN/the_lonely_explorer/internal/ui"
	player2 "github.com/VxVxN/the_lonely_explorer/pkg/player"
)

type Game struct {
	windowWidth, windowHeight float64
	tileSize                  int

	scene1UI *scene1UI

	imagesByObjID              map[int]*ebiten.Image
	animationByObjID           map[int]*animation.Animation
	gameMap                    *_map.Map
	mapScale                   float64
	collisionObjs              []*rectangle.Rectangle
	keyEventManager            *keyeventmanager.EventManager
	eventManager               *eventmanager.EventManager
	player                     *player2.Player
	journal                    *journal.Journal
	startPlayerX, startPlayerY float64
	stager                     *stager.Stager
	dialog                     *dialog.Dialog
	journalRecords             []journal.RecordJournal

	logger *slog.Logger
}

var backgroundColor = color.RGBA{0xf7, 0xf9, 0xb9, 0xff}

const (
	groundID         = 1
	parachute1       = 2
	parachute2       = 3
	plant1ID         = 4
	plant12D         = 5
	plant13D         = 6
	plant14D         = 7
	playerBack1ID    = 8
	playerBack2ID    = 9
	playerForward1ID = 10
	playerForward2ID = 11
	playerLeft1ID    = 12
	playerLeft2ID    = 13
	playerRight1ID   = 14
	playerRight2ID   = 15
	topSpongeID      = 16
	downSpongeID     = 17

	visibilityLimit = 11
)

func NewGame() (*Game, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	w, h := ebiten.Monitor().Size()
	logger.Info("Monitor size", "width", w, "height", h)
	//width, height := float64(w), float64(h)

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("can't get working dir: %s", err)
	}

	assetPath := path.Join(workingDir, "assets")

	tilesetImage, _, err := ebitenutil.NewImageFromFile(path.Join(assetPath, "tileset.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to init tileset image: %v", err)
	}

	gameMap, err := _map.NewMap(path.Join(workingDir, "map.json"))
	if err != nil {
		return nil, fmt.Errorf("can't init gameMap: %v", err)
	}

	tileSize := gameMap.Data.TileWidth

	logger.Info("Loading tileset",
		"tileSize", tileSize,
		"mapSize", fmt.Sprintf("(%dx%d)", gameMap.Data.Width, gameMap.Data.Height))

	supportedKeys := []ebiten.Key{
		ebiten.KeyUp,
		ebiten.KeyDown,
		ebiten.KeyLeft,
		ebiten.KeyRight,
		ebiten.KeyEscape,
		ebiten.KeyEnter,
		ebiten.KeyJ,
	}

	res, err := ui.NewUIResources()
	if err != nil {
		return nil, err
	}

	dialog := dialog.NewDialog(res)

	game := &Game{
		windowWidth:  float64(w),
		windowHeight: float64(h),
		tileSize:     tileSize,

		scene1UI: newScene1UI(res),

		imagesByObjID:    make(map[int]*ebiten.Image),
		animationByObjID: make(map[int]*animation.Animation),

		gameMap:         gameMap,
		mapScale:        1.5,
		keyEventManager: keyeventmanager.NewEventManager(supportedKeys),
		stager:          stager.New(),
		dialog:          dialog,

		logger: logger,
	}
	objIDs := []int{
		groundID,
		parachute1,
		parachute2,
		plant1ID,
		plant12D,
		plant13D,
		plant14D,
		playerBack1ID,
		playerBack2ID,
		playerForward1ID,
		playerForward2ID,
		playerLeft1ID,
		playerLeft2ID,
		playerRight1ID,
		playerRight2ID,
		topSpongeID,
		downSpongeID,
	}
	for _, id := range objIDs {
		game.imagesByObjID[id] = getSubImage(id, tilesetImage, tileSize)
	}

	font, err := loadDefaultFont()
	if err != nil {
		return nil, fmt.Errorf("can't load font: %v", err)
	}

	game.journal = journal.NewJournal(font)
	game.journal.SetPosition(100, 100)
	game.journal.SetBackgroundColor(color.RGBA{30, 30, 30, 200})

	plantAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[plant1ID], game.imagesByObjID[plant12D], game.imagesByObjID[plant13D], game.imagesByObjID[plant14D]})
	plantAnimation.SetScale(game.mapScale, game.mapScale)
	plantAnimation.SetReverse(true)
	plantAnimation.SetRepeatable(true)

	game.animationByObjID[plant1ID] = plantAnimation

	//game.stager.SetStage(stager.SceneStage)
	game.stager.SetStage(stager.GameStage)

	playerForwardAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[playerForward1ID], game.imagesByObjID[playerForward2ID]})
	playerForwardAnimation.SetScale(game.mapScale, game.mapScale)
	playerForwardAnimation.SetRepeatable(true)

	playerBackAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[playerBack1ID], game.imagesByObjID[playerBack2ID]})
	playerBackAnimation.SetScale(game.mapScale, game.mapScale)
	playerBackAnimation.SetRepeatable(true)

	playerLeftAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[playerLeft1ID], game.imagesByObjID[playerLeft2ID]})
	playerLeftAnimation.SetScale(game.mapScale, game.mapScale)
	playerLeftAnimation.SetRepeatable(true)

	playerRightAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[playerRight1ID], game.imagesByObjID[playerRight2ID]})
	playerRightAnimation.SetScale(game.mapScale, game.mapScale)
	playerRightAnimation.SetRepeatable(true)

	player := player2.NewPlayer(game.imagesByObjID[playerForward1ID], playerForwardAnimation, playerBackAnimation, playerLeftAnimation, playerRightAnimation, 4)
	player.SetScale(game.mapScale)
	game.player = player

	eventManager := eventmanager.NewEventManager(player, gameMap)
	eventManager.SetEvents([]eventmanager.Event{
		eventmanager.NewMeetEvent([]int{plant1ID}, func() {
			text := "FLORA-2284-Y (\"Солнечный шёпот\")  \n\nЖелтый, как сгусток инопланетного света, этот странный организм колышется в разреженном ветре Kepler-442b, будто пойманный в ловушку собственного сияния. Его лепестки, тонкие, как лезвия, мерцают неестественным золотом, словно впитали свет далекой звезды и теперь медленно излучают его обратно в сумрачный мир. При малейшем прикосновении растение звенит, будто стеклянная арфа, а его поверхность, покрытая серебристыми ворсинками, дрожит, словно живая ртуть. Оно не похоже на земные цветы — в нем нет ни мягкости, ни нежности, только холодная, почти механическая красота, словно сама планета вырастила его из металла и солнечного ветра. И когда ночь опускается на равнины, ксантоид начинает светиться изнутри, как забытый сигнальный маяк, будто пытается что-то сказать… или предупредить."
			turnOnDialog := func() {
				game.stager.SetStage(stager.DialogStage)
				game.dialog.TurnOn(text)
			}
			turnOnDialog()
			game.journalRecords = append(game.journalRecords, journal.RecordJournal{
				Image:       game.imagesByObjID[plant1ID],
				Description: text,
				Action:      turnOnDialog,
			})
		}),
		eventmanager.NewMeetEvent([]int{topSpongeID, downSpongeID}, func() {
			text := "FLORA-4712-P (\"Розовый Пульсар\")\n\nМягкий, почти неестественно пухлый, этот организм напоминает гигантскую каплю жевательной резинки, случайно упавшую на каменистую поверхность Kepler-442b. Его розовая, полупрозрачная поверхность переливается перламутровыми бликами, словно покрыта тонкой плёнкой слизи, но при этом выглядит сухой на ощупь. Цветок пульсирует едва заметно, как будто дышит, расширяясь и сжимаясь в медленном, гипнотическом ритме.\n\nПри приближении его бархатистая текстура внезапно меняется — поверхность вздымается крошечными пузырьками, словно кипящая жидкость, а затем снова опадает в гладкую массу. Если коснуться, он нежно дрожит, издавая слабый, похожий на бульканье звук, а затем медленно начинает менять оттенок — от нежно-розового до глубокого фуксии, будто реагируя на контакт."
			turnOnDialog := func() {
				game.stager.SetStage(stager.DialogStage)
				game.dialog.TurnOn(text)
			}
			turnOnDialog()
			game.journalRecords = append(game.journalRecords, journal.RecordJournal{
				Image:       game.imagesByObjID[topSpongeID],
				Description: text,
				Action:      turnOnDialog,
			})
		}),
	})

	game.eventManager = eventManager

	collisionPropertyByTIle := make(map[int]struct{})
	for _, tile := range gameMap.Data.Tilesets[0].Tiles {
		isCollision := false
		for _, property := range tile.Properties {
			if property.Name == "collision" {
				isCollision = true
				break
			}
		}
		if isCollision {
			collisionPropertyByTIle[tile.Id+1] = struct{}{}
		}
	}
	for x, column := range gameMap.Layers[1] {
		for y, tile := range column {
			if _, ok := collisionPropertyByTIle[tile]; !ok {
				continue
			}
			game.collisionObjs = append(game.collisionObjs, rectangle.New(float64(x*game.tileSize), float64(y*game.tileSize), float64(game.tileSize), float64(game.tileSize)))
		}
	}
	for x, column := range gameMap.Layers[2] {
		for y, tile := range column {
			if tile != playerForward1ID {
				continue
			}
			xPixel := float64(x * game.tileSize)
			yPixel := float64(y * game.tileSize)
			game.player.SetPosition(xPixel, yPixel)
			game.startPlayerX, game.startPlayerY = xPixel, yPixel
			break
		}
	}

	game.addEvents()

	return game, nil
}

func (game *Game) Update() error {
	game.keyEventManager.Update()

	switch game.stager.Stage() {
	case stager.JournalStage:
		game.journal.Update()
	case stager.SceneStage:
		game.scene1UI.ui.Update()
		return nil
	case stager.DialogStage:
		return nil
	case stager.GameStage:
		game.player.Update()
		game.eventManager.Update()

		for _, animation := range game.animationByObjID {
			animation.Update(0.05)
		}
		return nil
	}
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	switch game.stager.Stage() {
	case stager.SceneStage:
		game.scene1UI.ui.Draw(screen)
		return
	case stager.GameStage:
	}
	screen.Fill(backgroundColor)
	centerWindowX := (game.windowWidth/2 - float64(game.tileSize)/2) / game.mapScale
	centerWindowY := (game.windowHeight/2 - float64(game.tileSize)/2) / game.mapScale

	for _, layer := range game.gameMap.Layers {
	nextX:
		for x, column := range layer {
			for y, tile := range column {
				if tile == 0 {
					continue // empty tile
				}
				if x+visibilityLimit < int(game.player.X)/game.tileSize || x-visibilityLimit > int(game.player.X)/game.tileSize {
					continue nextX
				}
				if y+visibilityLimit < int(game.player.Y)/game.tileSize || y-visibilityLimit > int(game.player.Y)/game.tileSize {
					continue
				}
				img, ok := game.imagesByObjID[tile]
				if !ok {
					game.logger.Error("Unknown tile", "tile", tile)
					continue
				}

				var xPixel, yPixel float64
				if tile == playerForward1ID {
					continue
				}
				xPixel = (float64(x*game.tileSize) - game.player.X) + centerWindowX
				yPixel = (float64(y*game.tileSize) - game.player.Y) + centerWindowY
				animation, ok := game.animationByObjID[tile]
				if ok {
					animation.Start()
					animation.SetPosition(xPixel, yPixel)
					animation.Draw(screen)
					continue
				}

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(xPixel, yPixel)
				op.GeoM.Scale(game.mapScale, game.mapScale)
				screen.DrawImage(img, op)
			}
		}
	}
	game.player.Draw(screen, centerWindowX, centerWindowY)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Player %.0fx%.0f", game.player.X, game.player.Y))
	game.dialog.Draw(screen)
	game.journal.Draw(screen)
}

func (game *Game) Layout(screenWidthPx, screenHeightPx int) (int, int) {
	return screenWidthPx, screenHeightPx
}

func (game *Game) addEvents() {
	game.keyEventManager.AddPressEvent(ebiten.KeyRight, func() {
		switch game.stager.Stage() {
		case stager.GameStage:
			if !game.player.Dead() && game.player.X+float64(game.tileSize) < float64(game.gameMap.Data.Width*game.tileSize) {
				game.player.Rectangle.X += game.player.Speed()
				for _, obj := range game.collisionObjs {
					if game.player.Rectangle.Collision(obj) {
						game.player.Rectangle.X -= game.player.Speed()
						return
					}
				}
				game.player.Rectangle.X -= game.player.Speed()
				game.player.Move(ebiten.KeyRight)
			}
		}
	})
	game.keyEventManager.AddPressEvent(ebiten.KeyLeft, func() {
		switch game.stager.Stage() {
		case stager.GameStage:
			if !game.player.Dead() && game.player.X > 0 {
				game.player.Rectangle.X -= game.player.Speed()
				for _, obj := range game.collisionObjs {
					if game.player.Rectangle.Collision(obj) {
						game.player.Rectangle.X += game.player.Speed()
						return
					}
				}
				game.player.Rectangle.X += game.player.Speed()
				game.player.Move(ebiten.KeyLeft)
			}
		}
	})
	game.keyEventManager.AddPressEvent(ebiten.KeyUp, func() {
		switch game.stager.Stage() {
		case stager.GameStage:
			if !game.player.Dead() && game.player.Y > 0 {
				game.player.Rectangle.Y -= game.player.Speed()
				for _, obj := range game.collisionObjs {
					if game.player.Rectangle.Collision(obj) {
						game.player.Rectangle.Y += game.player.Speed()
						return
					}
				}
				game.player.Rectangle.Y += game.player.Speed()
				game.player.Move(ebiten.KeyUp)
			}
		}
	})
	game.keyEventManager.AddPressEvent(ebiten.KeyDown, func() {
		switch game.stager.Stage() {
		case stager.GameStage:
			if !game.player.Dead() && game.player.Y+float64(game.tileSize) < float64(game.gameMap.Data.Height*game.tileSize) {
				game.player.Rectangle.Y += game.player.Speed()
				for _, obj := range game.collisionObjs {
					if game.player.Rectangle.Collision(obj) {
						game.player.Rectangle.Y -= game.player.Speed()
						return
					}
				}
				game.player.Rectangle.Y -= game.player.Speed()
				game.player.Move(ebiten.KeyDown)
			}
		}
	})
	game.keyEventManager.AddPressedEvent(ebiten.KeyEnter, func() {
		switch game.stager.Stage() {
		case stager.SceneStage:
			game.stager.SetStage(stager.GameStage)
		case stager.DialogStage:
			game.stager.RecoveryLastStage()
			game.dialog.TurnOff()
		}
	})
	game.keyEventManager.AddPressedEvent(ebiten.KeyJ, func() {
		switch game.stager.Stage() {
		case stager.GameStage:
			game.journal.TurnOnOff()
			game.journal.SetKnowRecords(game.journalRecords)
			game.stager.SetStage(stager.JournalStage)
		case stager.JournalStage:
			game.journal.TurnOnOff()
			game.stager.SetStage(stager.GameStage)
		}
	})
	game.keyEventManager.AddPressedEvent(ebiten.KeyEscape, func() {
		os.Exit(0)
	})
	game.keyEventManager.SetDefaultEvent(func() {
		game.player.Move(ebiten.Key0) // not move player
	})
}

func (game *Game) Close() {}

func getSubImage(id int, tilesetImage *ebiten.Image, tileSize int) *ebiten.Image {
	row := (id - 1) / 10
	col := (id - 1) % 10
	x := col * tileSize
	y := row * tileSize

	return tilesetImage.SubImage(image.Rect(x, y, x+tileSize, y+tileSize)).(*ebiten.Image)
}

func loadDefaultFont() (font.Face, error) {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		return nil, err
	}

	const dpi = 72
	fontFace, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	return fontFace, nil
}
