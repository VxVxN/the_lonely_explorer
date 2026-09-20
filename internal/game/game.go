package game

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/VxVxN/gamedevlib/animation"
	keyeventmanager "github.com/VxVxN/gamedevlib/eventmanager"
	"github.com/VxVxN/gamedevlib/rectangle"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/vector"

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

	// spongeTicks хранит счётчик мигания для губок рядом с игроком по координатам тайла.
	pendingRespawn bool                // робот погиб, ждём закрытия сообщения о гибели
	acidReported   bool                // описание опасного растения уже показано и записано в журнал
	acidSponges    map[[2]int]struct{} // тайлы topSpongeID растений, которые распыляют кислоту
	deathTicks     int                 // кадров с начала фазы гибели (0, если робот в безопасности)
	robotNumber    int                 // номер текущего робота: RX-<robotNumber>, растёт после каждой смерти
	spongeTicks    map[[2]int]int
	// spongeAnimTicks хранит счётчик анимации topSpongeID для тайлов вплотную к игроку.
	spongeAnimTicks map[[2]int]int

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
	topSponge2ID     = 19
	downSponge2ID    = 20
	topSpongeAnim1ID = 21
	topSpongeAnim2ID = 22
	topSpongeAnim3ID = 23
	topSpongeAnim4ID = 24
	spongeExtraID    = 25
	deadRobotID      = 26

	visibilityLimit = 11

	spongeActivationTiles = 3
	spongeCloseTiles      = 1.5 // расстояние, на котором срабатывает анимация topSpongeID
	spongeIntroFrames     = 48  // кадров на один кадр первого проигрывания анимации topSpongeID
	spongeDeathFrames     = 90  // кадров (1,5 сек) после начала цикла 4–5, через которые робот погибает
	spongeLoopFrames      = 24  // кадров на один кадр зацикленной части (4-й и 5-й)
	acidPlantShare        = 0.2 // доля растений, распыляющих кислоту
	spongeBlinkFrames     = 40  // кадров на одно состояние при мигании

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
		robotNumber:      1,
		spongeTicks:      make(map[[2]int]int),
		spongeAnimTicks:  make(map[[2]int]int),
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
		topSponge2ID,
		downSponge2ID,
		topSpongeAnim1ID,
		topSpongeAnim2ID,
		topSpongeAnim3ID,
		topSpongeAnim4ID,
		spongeExtraID,
		deadRobotID,
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
		eventmanager.NewMeetEventWhere([]int{topSpongeID, downSpongeID}, game.isSafeSponge, func() {
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
	game.pickAcidSponges()

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

		game.updateSponges()

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

				if ticks, ok := game.spongeTicks[[2]int{x, y}]; ok && (ticks/spongeBlinkFrames)%2 == 1 {
					if img2, ok := game.imagesByObjID[spongeSecondState(tile)]; ok {
						img = img2
					}
				}

				if ticks, ok := game.spongeAnimTicks[[2]int{x, y}]; ok && tile == topSpongeID {
					img = game.imagesByObjID[topSpongeAnimID(ticks)]
				}

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(xPixel, yPixel)
				op.GeoM.Scale(game.mapScale, game.mapScale)
				screen.DrawImage(img, op)
			}
		}
	}
	if !game.pendingRespawn {
		game.player.Draw(screen, centerWindowX, centerWindowY)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Player %.0fx%.0f", game.player.X, game.player.Y))
	game.drawDeathPulse(screen)
	game.dialog.Draw(screen)
	game.journal.Draw(screen)
}

// drawDeathPulse заливает экран пульсирующим красным: чем ближе гибель, тем сильнее и чаще пульс.
func (game *Game) drawDeathPulse(screen *ebiten.Image) {
	if game.deathTicks <= 0 {
		return
	}
	progress := float64(game.deathTicks) / spongeDeathFrames
	phase := float64(game.deathTicks) * (0.08 + 0.15*progress)
	pulse := (1 + math.Sin(phase)) / 2
	alpha := (0.1 + 0.4*pulse) * (0.4 + 0.6*progress)
	a := uint8(alpha * 255) // цвет с предумноженной альфой
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.RGBA{R: a, A: a}, false)
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
			game.dialog.TurnOff()
			if game.pendingRespawn {
				game.respawn()
				return
			}
			game.stager.SetStage(stager.GameStage)
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

var topSpongeAnimIDs = []int{topSpongeAnim1ID, topSpongeAnim2ID, topSpongeAnim3ID, topSpongeAnim4ID}

// topSpongeAnimID: первые кадры проигрываются один раз (медленно), затем 4-й и 5-й чередуются по кругу.
func topSpongeAnimID(ticks int) int {
	last := len(topSpongeAnimIDs) - 1
	if ticks < last*spongeIntroFrames {
		return topSpongeAnimIDs[ticks/spongeIntroFrames]
	}
	if ((ticks-last*spongeIntroFrames)/spongeLoopFrames)%2 == 0 {
		return topSpongeAnimIDs[last]
	}
	return spongeExtraID
}

func spongeSecondState(tile int) int {
	switch tile {
	case topSpongeID:
		return topSponge2ID
	case downSpongeID:
		return downSponge2ID
	}
	return 0
}

// killPlayer оставляет мёртвого робота там, где он погиб, и показывает сообщение о гибели; после него срабатывает respawn.
func (game *Game) killPlayer() {
	tileX := int(math.Round(game.player.X / float64(game.tileSize)))
	tileY := int(math.Round(game.player.Y / float64(game.tileSize)))
	for _, layer := range game.gameMap.Layers[1:] {
		if tileX < 0 || tileX >= len(layer) || tileY < 0 || tileY >= len(layer[tileX]) {
			break
		}
		if layer[tileX][tileY] == 0 {
			layer[tileX][tileY] = deadRobotID
			ts := float64(game.tileSize)
			game.collisionObjs = append(game.collisionObjs, rectangle.New(float64(tileX)*ts, float64(tileY)*ts, ts, ts))
			break
		}
	}

	clear(game.spongeTicks)
	clear(game.spongeAnimTicks)
	game.deathTicks = 0

	game.pendingRespawn = true
	game.stager.SetStage(stager.DialogStage)
	game.dialog.TurnOn("Растение оказалось распыляющим кислоту. Робот RX-" + strconv.Itoa(game.robotNumber) + " был разрушен.")
}

// respawn возвращает нового робота на старт и показывает начальную сцену.
func (game *Game) respawn() {
	game.pendingRespawn = false
	game.player.SetPosition(game.startPlayerX, game.startPlayerY)
	game.player.Move(ebiten.Key0)
	game.robotNumber++
	game.scene1UI.SetRobotNumber(game.robotNumber)
	game.stager.SetStage(stager.SceneStage)
}

// pickAcidSponges выбирает случайные acidPlantShare растений (связных групп клеток губки), которые распыляют кислоту.
// Хотя бы одно растение выбирается всегда, если они есть на карте.
func (game *Game) pickAcidSponges() {
	game.acidSponges = make(map[[2]int]struct{})

	seen := make(map[[2]int]struct{})
	var plants [][][2]int
	for x := range game.gameMap.Layers[0] {
		for y := range game.gameMap.Layers[0][x] {
			start := [2]int{x, y}
			if _, ok := seen[start]; ok || !game.isSponge(x, y) {
				continue
			}
			plant := [][2]int{start}
			seen[start] = struct{}{}
			for i := 0; i < len(plant); i++ {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						next := [2]int{plant[i][0] + dx, plant[i][1] + dy}
						if _, ok := seen[next]; ok || !game.isSponge(next[0], next[1]) {
							continue
						}
						seen[next] = struct{}{}
						plant = append(plant, next)
					}
				}
			}
			plants = append(plants, plant)
		}
	}

	rand.Shuffle(len(plants), func(i, j int) { plants[i], plants[j] = plants[j], plants[i] })
	count := int(math.Round(float64(len(plants)) * acidPlantShare))
	if len(plants) > 0 {
		count = max(count, 1)
	}
	for _, plant := range plants[:count] {
		for _, tile := range plant {
			game.acidSponges[tile] = struct{}{}
		}
	}
	game.logger.Info("Acid plants picked", "plants", len(plants), "acid", count)
}

func (game *Game) isSafeSponge(x, y int) bool {
	_, acid := game.acidSponges[[2]int{x, y}]
	return !acid
}

// reportAcidPlant показывает описание опасного растения и записывает его в журнал, когда оно начинает брызгать кислотой.
func (game *Game) reportAcidPlant() {
	game.acidReported = true
	text := "FLORA-4712-A (\"Розовый Пульсар\", кислотная форма)\n\nВНИМАНИЕ: ОПАСНО. Внешне организм неотличим от безопасных особей своего вида, однако при приближении на расстояние менее полутора метров он раскрывает поры и выбрасывает плотную струю концентрированной кислоты. Сенсоры зафиксировали pH ниже 1 и стремительное разрушение внешних покровов робота.\n\nОрганизм активен, реагирует на присутствие в радиусе нескольких метров — его выдаёт короткое мерцание перед выбросом. Классифицировано как смертельно опасное. Рекомендация: не приближаться вплотную к растениям этого вида, обходить их на безопасном расстоянии."
	turnOnDialog := func() {
		game.journal.TurnOff()
		game.stager.SetStage(stager.DialogStage)
		game.dialog.TurnOn(text)
	}
	turnOnDialog()
	game.journalRecords = append(game.journalRecords, journal.RecordJournal{
		Image:       game.imagesByObjID[topSpongeAnim4ID],
		Description: text,
		Category:    "Флора",
		Action:      turnOnDialog,
	})
}

func (game *Game) isSponge(x, y int) bool {
	for _, layer := range game.gameMap.Layers {
		if x >= 0 && x < len(layer) && y >= 0 && y < len(layer[x]) && spongeSecondState(layer[x][y]) != 0 {
			return true
		}
	}
	return false
}

// updateSponges заставляет губки мигать между двумя состояниями, пока игрок ближе трёх клеток.
func (game *Game) updateSponges() {
	ts := float64(game.tileSize)
	playerCX, playerCY := game.player.X+ts/2, game.player.Y+ts/2
	limit := spongeActivationTiles * ts
	px, py := int(game.player.X)/game.tileSize, int(game.player.Y)/game.tileSize
	r := spongeActivationTiles + 1

	near := make(map[[2]int]struct{})
	close := make(map[[2]int]struct{})
	for _, layer := range game.gameMap.Layers {
		for x := max(px-r, 0); x <= px+r && x < len(layer); x++ {
			for y := max(py-r, 0); y <= py+r && y < len(layer[x]); y++ {
				if spongeSecondState(layer[x][y]) == 0 {
					continue
				}
				_, acid := game.acidSponges[[2]int{x, y}]
				if !acid {
					continue
				}
				dx := float64(x)*ts + ts/2 - playerCX
				dy := float64(y)*ts + ts/2 - playerCY
				if dx*dx+dy*dy < limit*limit {
					near[[2]int{x, y}] = struct{}{}
				}
				if layer[x][y] == topSpongeID && dx*dx+dy*dy < spongeCloseTiles*spongeCloseTiles*ts*ts {
					close[[2]int{x, y}] = struct{}{}
				}
			}
		}
	}

	// растение мигает целиком: заражаем соседние клетки с губками (включая диагонали)
	queue := make([][2]int, 0, len(near))
	for key := range near {
		queue = append(queue, key)
	}
	for len(queue) > 0 {
		cur := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				next := [2]int{cur[0] + dx, cur[1] + dy}
				if _, ok := near[next]; ok || !game.isSponge(next[0], next[1]) {
					continue
				}
				near[next] = struct{}{}
				queue = append(queue, next)
			}
		}
	}

	killed := false
	game.deathTicks = 0
	for key := range close {
		game.spongeAnimTicks[key]++
		dying := game.spongeAnimTicks[key] - (len(topSpongeAnimIDs)-1)*spongeIntroFrames
		game.deathTicks = max(game.deathTicks, dying)
		if dying >= 0 && !game.acidReported {
			game.reportAcidPlant()
		}
		if dying >= spongeDeathFrames {
			killed = true
		}
	}
	if killed {
		game.killPlayer()
		return
	}
	for key := range game.spongeAnimTicks {
		if _, ok := close[key]; !ok {
			delete(game.spongeAnimTicks, key)
		}
	}

	for key := range near {
		game.spongeTicks[key]++
	}
	for key := range game.spongeTicks {
		if _, ok := near[key]; !ok {
			delete(game.spongeTicks, key)
		}
	}
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
