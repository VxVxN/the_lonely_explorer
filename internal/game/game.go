package game

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/VxVxN/gamedevlib/animation"
	keyeventmanager "github.com/VxVxN/gamedevlib/eventmanager"
	"github.com/VxVxN/gamedevlib/rectangle"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"

	"github.com/VxVxN/the_lonely_explorer/internal/eventmanager"
	"github.com/VxVxN/the_lonely_explorer/internal/journal"
	"github.com/VxVxN/the_lonely_explorer/pkg/dialog"

	_map "github.com/VxVxN/the_lonely_explorer/internal/map"
	"github.com/VxVxN/the_lonely_explorer/internal/stager"
	player2 "github.com/VxVxN/the_lonely_explorer/pkg/player"
	"github.com/VxVxN/the_lonely_explorer/pkg/simpleui"
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

	stationPlaced   bool
	stationReady    bool
	stationTileX    int
	stationTileY    int
	stationSoilID   int
	stationPlacedAt time.Time
	researchedSoils map[int]struct{}

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
	stationID        = 18

	visibilityLimit = 11

	stationResearchDuration = time.Second * 10
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
		ebiten.KeyB,
	}

	bodyFace, err := simpleui.LoadFont(fonts.MPlus1pRegular_ttf, 28)
	if err != nil {
		return nil, fmt.Errorf("can't load UI font: %v", err)
	}
	dialogFace, err := simpleui.LoadFont(fonts.MPlus1pRegular_ttf, 18)
	if err != nil {
		return nil, fmt.Errorf("can't load UI font: %v", err)
	}

	theme := simpleui.DefaultTheme(bodyFace)
	dialogTheme := *theme
	dialogTheme.Face = dialogFace

	dialog := dialog.NewDialog(&dialogTheme)

	game := &Game{
		windowWidth:  float64(w),
		windowHeight: float64(h),
		tileSize:     tileSize,

		scene1UI: newScene1UI(theme, float64(w), float64(h)),

		imagesByObjID:    make(map[int]*ebiten.Image),
		animationByObjID: make(map[int]*animation.Animation),

		gameMap:         gameMap,
		mapScale:        1.5,
		keyEventManager: keyeventmanager.NewEventManager(supportedKeys),
		stager:          stager.New(),
		dialog:          dialog,

		researchedSoils: make(map[int]struct{}),

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
		stationID,
	}
	for _, id := range objIDs {
		game.imagesByObjID[id] = getSubImage(id, tilesetImage, tileSize)
	}

	game.journal = journal.NewJournal(theme, float64(w), float64(h))
	game.journal.SetPosition(100, 100)

	plantAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[plant1ID], game.imagesByObjID[plant12D], game.imagesByObjID[plant13D], game.imagesByObjID[plant14D]})
	plantAnimation.SetScale(game.mapScale, game.mapScale)
	plantAnimation.SetReverse(true)
	plantAnimation.SetRepeatable(true)

	game.animationByObjID[plant1ID] = plantAnimation

	game.stager.SetStage(stager.SceneStage)
	//game.stager.SetStage(stager.GameStage)

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
			text := "FLORA-2284-Y (\"Солнечный шёпот\")  \n\nЖёлтый организм простой формы: два толстых корня, вросших в грунт Kepler-442b, соединены единственным гибким стеблем. Стебель не замирает ни на секунду — он непрерывно покачивается из стороны в сторону, будто его треплет ветер, даже когда воздух вокруг неподвижен.\n\nСканирование не выявило ни токсинов, ни защитных механизмов: колебание стебля — это, судя по всему, просто способ организма улавливать свет или влагу, а не реакция на угрозу. Контакт не вызывает никакого отклика, кроме лёгкой вибрации. Организм статичен, неподвижен корнями и не проявляет агрессии. Классифицировано как безопасное."
			turnOnDialog := func() {
				game.journal.TurnOff()
				game.stager.SetStage(stager.DialogStage)
				game.dialog.TurnOn(text)
			}
			turnOnDialog()
			game.journalRecords = append(game.journalRecords, journal.RecordJournal{
				Image:       game.imagesByObjID[plant1ID],
				Description: text,
				Category:    "Флора",
				Action:      turnOnDialog,
			})
		}),
		eventmanager.NewMeetEvent([]int{topSpongeID, downSpongeID}, func() {
			text := "FLORA-4712-P (\"Розовый Пульсар\")\n\nМягкий, пористый организм, по виду и текстуре напоминающий воздушную губку — вся его розовая поверхность испещрена мелкими порами, через которые, судя по показаниям сенсоров, идёт медленный газообмен с атмосферой Kepler-442b. Тело полностью неподвижно: ни движения, ни пульсации не зафиксировано за всё время наблюдения.\n\nПри контакте отклика не последовало — ни изменения формы, ни звука. Анализ подтверждает: ни ядовитых спор, ни раздражающих веществ не обнаружено. Организм статичен и не представляет угрозы — классифицировано как безопасное."
			turnOnDialog := func() {
				game.journal.TurnOff()
				game.stager.SetStage(stager.DialogStage)
				game.dialog.TurnOn(text)
			}
			turnOnDialog()
			game.journalRecords = append(game.journalRecords, journal.RecordJournal{
				Image:       game.imagesByObjID[topSpongeID],
				Description: text,
				Category:    "Флора",
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

		if game.stationPlaced && !game.stationReady && time.Since(game.stationPlacedAt) >= stationResearchDuration {
			game.stationReady = true
			text := "СТАНЦИЯ RX-7: обследование почвы и воздуха завершено. Вернитесь к месту установки и заберите собранные данные."
			game.stager.SetStage(stager.DialogStage)
			game.dialog.TurnOn(text)
		}

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
			game.stager.SetStage(stager.GameStage)
			game.dialog.TurnOff()
		case stager.JournalStage:
			game.journal.ActivateSelection()
		}
	})
	game.keyEventManager.AddPressedEvent(ebiten.KeyUp, func() {
		switch game.stager.Stage() {
		case stager.JournalStage:
			game.journal.MoveSelection(-1)
		}
	})
	game.keyEventManager.AddPressedEvent(ebiten.KeyDown, func() {
		switch game.stager.Stage() {
		case stager.JournalStage:
			game.journal.MoveSelection(1)
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
	game.keyEventManager.AddPressedEvent(ebiten.KeyB, func() {
		switch game.stager.Stage() {
		case stager.GameStage:
			if game.stationPlaced {
				game.collectStation()
			} else {
				game.placeStation()
			}
		}
	})
	game.keyEventManager.AddPressedEvent(ebiten.KeyEscape, func() {
		os.Exit(0)
	})
	game.keyEventManager.SetDefaultEvent(func() {
		game.player.Move(ebiten.Key0) // not move player
	})
}

func (game *Game) placeStation() {
	tileX := int(game.player.X) / game.tileSize
	tileY := int(game.player.Y) / game.tileSize
	soilID := game.gameMap.Layers[0][tileX][tileY]

	if _, ok := game.researchedSoils[soilID]; ok {
		game.showStationMessage("СТАНЦИЯ RX-7: проба грунта на этом участке совпадает с уже исследованным образцом. Здесь нечего исследовать.")
		return
	}

	game.gameMap.Layers[1][tileX][tileY] = stationID
	game.stationPlaced = true
	game.stationReady = false
	game.stationTileX = tileX
	game.stationTileY = tileY
	game.stationSoilID = soilID
	game.stationPlacedAt = time.Now()

	game.showStationMessage("СТАНЦИЯ RX-7: развёрнута. Начато обследование почвы и состава воздуха — потребуется некоторое время. Вернитесь позже, чтобы забрать собранные данные.")
}

func (game *Game) collectStation() {
	dx := int(game.player.X)/game.tileSize - game.stationTileX
	dy := int(game.player.Y)/game.tileSize - game.stationTileY
	if abs(dx) > 1 || abs(dy) > 1 {
		game.showStationMessage("СТАНЦИЯ RX-7: станция находится в другом месте. Сначала нужно вернуться к ней и забрать данные.")
		return
	}

	if !game.stationReady {
		game.showStationMessage("СТАНЦИЯ RX-7: обследование ещё продолжается. Вернитесь позже.")
		return
	}

	game.researchedSoils[game.stationSoilID] = struct{}{}
	game.gameMap.Layers[1][game.stationTileX][game.stationTileY] = 0
	game.stationPlaced = false
	game.stationReady = false

	text := stationReportText(game.stationSoilID)
	turnOnDialog := func() {
		game.journal.TurnOff()
		game.stager.SetStage(stager.DialogStage)
		game.dialog.TurnOn(text)
	}
	turnOnDialog()
	game.journalRecords = append(game.journalRecords, journal.RecordJournal{
		Image:       game.imagesByObjID[stationID],
		Description: text,
		Category:    "Исследования станции",
		Action:      turnOnDialog,
	})
}

func stationReportText(soilID int) string {
	switch soilID {
	case groundID:
		return "ГРУНТ #1 — Песчаная равнина\n\nПервый образец, взятый роботом на Kepler-442b: мелкозернистый песок охристого оттенка, состоящий преимущественно из силикатных частиц с высокой отражающей способностью — вероятно, именно поэтому равнина вокруг места посадки светится под местным солнцем ярче, чем показывали орбитальные снимки. Органических соединений не обнаружено, зато зафиксирован лёгкий электростатический заряд: частицы трутся друг о друга на ветру и слабо потрескивают под манипулятором станции.\n\nДатчики зафиксировали температуру поверхности +" +
			"76 °C в момент замера и падение до +46 °C уже через несколько часов после захода светила — суточный перепад около тридцати градусов, характерный для сухого грунта без растительного покрова, который не удерживает тепло. Радиационный фон в норме, токсичных примесей нет. Грунт стабилен и не представляет угрозы, но беден питательными веществами — для земледелия потребуется обработка. Первый кирпичик в общей картине пригодности планеты для жизни."
	default:
		return fmt.Sprintf("ДАННЫЕ СТАНЦИИ RX-7 (образец грунта #%d)\n\nАнализ образца грунта и локальной атмосферы завершён. Опасных веществ, токсичных примесей и активных биологических агентов не обнаружено. Результат внесён в бортовой архив.", soilID)
	}
}

func (game *Game) showStationMessage(text string) {
	game.stager.SetStage(stager.DialogStage)
	game.dialog.TurnOn(text)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (game *Game) Close() {}

func getSubImage(id int, tilesetImage *ebiten.Image, tileSize int) *ebiten.Image {
	row := (id - 1) / 10
	col := (id - 1) % 10
	x := col * tileSize
	y := row * tileSize

	return tilesetImage.SubImage(image.Rect(x, y, x+tileSize, y+tileSize)).(*ebiten.Image)
}
