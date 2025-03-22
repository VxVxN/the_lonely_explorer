package game

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"path"

	"github.com/VxVxN/gamedevlib/eventmanager"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	_map "github.com/VxVxN/the_lonely_explorer/internal/map"
	"github.com/VxVxN/the_lonely_explorer/internal/stager"
	"github.com/VxVxN/the_lonely_explorer/internal/ui"
	player2 "github.com/VxVxN/the_lonely_explorer/pkg/player"
)

type Game struct {
	windowWidth, windowHeight float64
	tileSize                  int

	scene1UI *scene1UI

	groundImage *ebiten.Image
	plantImage  *ebiten.Image
	robotImage  *ebiten.Image
	spongeImage *ebiten.Image

	gameMap      *_map.Map
	eventManager *eventmanager.EventManager
	player       *player2.Player
	stager       *stager.Stager

	logger *slog.Logger
}

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
	}

	res, err := ui.NewUIResources()
	if err != nil {
		return nil, err
	}

	game := &Game{
		windowWidth:  float64(w),
		windowHeight: float64(h),
		tileSize:     tileSize,

		scene1UI: newScene1UI(res),

		groundImage: tilesetImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image),
		plantImage:  tilesetImage.SubImage(image.Rect(tileSize, 0, tileSize*2, tileSize)).(*ebiten.Image),
		spongeImage: tilesetImage.SubImage(image.Rect(tileSize*2, 0, tileSize*3, tileSize)).(*ebiten.Image),
		robotImage:  tilesetImage.SubImage(image.Rect(tileSize*3, 0, tileSize*4, tileSize)).(*ebiten.Image),

		gameMap:      gameMap,
		eventManager: eventmanager.NewEventManager(supportedKeys),
		stager:       stager.New(),

		logger: logger,
	}
	game.stager.SetStage(stager.GameStage)

	player := player2.NewPlayer(game.robotImage, 3)
	game.player = player

	game.addEvents()

	return game, nil
}

func (game *Game) Update() error {
	game.eventManager.Update()

	switch game.stager.Stage() {
	case stager.MainMenuStage:
	case stager.Scene1Stage:
		game.scene1UI.ui.Update()
		return nil
	case stager.GameStage:
		game.player.Update()
		return nil
	}
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	switch game.stager.Stage() {
	case stager.Scene1Stage:
		game.scene1UI.ui.Draw(screen)
		return
	case stager.GameStage:
	}
	scale := 1.5
	for _, layer := range game.gameMap.Data.Layers {
		for i, datum := range layer.Data {
			var img *ebiten.Image
			switch datum {
			case 0:
				// empty tile
				continue
			case 1:
				img = game.groundImage
			case 2:
				img = game.plantImage
			case 3:
				img = game.spongeImage
			case 4:
				img = game.robotImage
			default:
				//game.logger.Error("Unknown layer", "image", datum)
				continue
			}

			shiftX := game.player.X
			shiftY := game.player.Y
			if datum != 4 {
				shiftX = 0
				shiftY = 0
			}

			x := float64(i%game.gameMap.Data.Width*game.tileSize) + shiftX
			y := float64(i/game.gameMap.Data.Height*game.tileSize) + shiftY

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, y)
			op.GeoM.Scale(scale, scale)
			screen.DrawImage(img, op)
		}
	}
}

func (game *Game) Layout(screenWidthPx, screenHeightPx int) (int, int) {
	return screenWidthPx, screenHeightPx
}

func (game *Game) addEvents() {
	//game.eventManager.AddPressEvent(ebiten.KeyRight, func() {
	//	if !game.player.Dead() && game.player.X < game.windowWidth {
	//		game.player.Move(ebiten.KeyRight)
	//	}
	//})
	//game.eventManager.AddPressEvent(ebiten.KeyLeft, func() {
	//	if !game.player.Dead() && game.player.X < game.windowWidth {
	//		game.player.Move(ebiten.KeyLeft)
	//	}
	//})
	//game.eventManager.AddPressEvent(ebiten.KeyUp, func() {
	//	if !game.player.Dead() && game.player.Y > 0 {
	//		game.player.Move(ebiten.KeyUp)
	//	}
	//})
	//game.eventManager.AddPressEvent(ebiten.KeyDown, func() {
	//	if !game.player.Dead() && game.player.Y < game.windowHeight {
	//		game.player.Move(ebiten.KeyDown)
	//	}
	//})
	game.eventManager.AddPressedEvent(ebiten.KeyEnter, func() {
		switch game.stager.Stage() {
		case stager.Scene1Stage:
			game.stager.SetStage(stager.GameStage)
		}
	})
	game.eventManager.AddPressedEvent(ebiten.KeyEscape, func() {
		os.Exit(0)
	})
}

func (game *Game) Close() {}
