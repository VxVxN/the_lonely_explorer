package game

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path"

	"github.com/VxVxN/gamedevlib/eventmanager"
	"github.com/VxVxN/gamedevlib/rectangle"
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

	images                     map[int]*ebiten.Image
	gameMap                    *_map.Map
	mapScale                   float64
	collisionObjs              []*rectangle.Rectangle
	eventManager               *eventmanager.EventManager
	player                     *player2.Player
	startPlayerX, startPlayerY float64
	stager                     *stager.Stager

	logger *slog.Logger
}

var backgroundColor = color.RGBA{0xf7, 0xf9, 0xb9, 0xff}

const (
	groundID     = 1
	parachute1   = 2
	parachute2   = 3
	plantID      = 4
	playerID     = 10
	topSpongeID  = 16
	downSpongeID = 17
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

		images: map[int]*ebiten.Image{
			groundID:     getSubImage(groundID, tilesetImage, tileSize),
			parachute1:   getSubImage(parachute1, tilesetImage, tileSize),
			parachute2:   getSubImage(parachute2, tilesetImage, tileSize),
			plantID:      getSubImage(plantID, tilesetImage, tileSize),
			playerID:     getSubImage(playerID, tilesetImage, tileSize),
			topSpongeID:  getSubImage(topSpongeID, tilesetImage, tileSize),
			downSpongeID: getSubImage(downSpongeID, tilesetImage, tileSize),
		},

		gameMap:      gameMap,
		mapScale:     1.5,
		eventManager: eventmanager.NewEventManager(supportedKeys),
		stager:       stager.New(),

		logger: logger,
	}
	game.stager.SetStage(stager.GameStage)

	player := player2.NewPlayer(game.images[playerID], 6)
	game.player = player

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
	for i, tile := range gameMap.Data.Layers[1].Data {
		if _, ok := collisionPropertyByTIle[tile]; !ok {
			continue
		}
		x := float64(i % game.gameMap.Data.Width * game.tileSize)
		y := float64(i / game.gameMap.Data.Height * game.tileSize)
		game.collisionObjs = append(game.collisionObjs, rectangle.New(x, y, float64(game.tileSize), float64(game.tileSize)))
	}
	for i, tile := range gameMap.Data.Layers[2].Data {
		if tile != playerID {
			continue
		}
		x := float64(i % game.gameMap.Data.Width * game.tileSize)
		y := float64(i / game.gameMap.Data.Height * game.tileSize)
		game.player.SetPosition(x, y)
		game.startPlayerX, game.startPlayerY = x, y
		break
	}

	game.addEvents()

	return game, nil
}

func (game *Game) Update() error {
	game.eventManager.Update()

	switch game.stager.Stage() {
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
	screen.Fill(backgroundColor)
	switch game.stager.Stage() {
	case stager.Scene1Stage:
		game.scene1UI.ui.Draw(screen)
		return
	case stager.GameStage:
	}
	for _, layer := range game.gameMap.Data.Layers {
		for i, datum := range layer.Data {
			img, ok := game.images[datum]
			if !ok {
				//game.logger.Error("Unknown layer", "image", datum)
				continue
			}

			centerWindowX := (game.windowWidth/2 - float64(game.tileSize)/2) / game.mapScale
			centerWindowY := (game.windowHeight/2 - float64(game.tileSize)/2) / game.mapScale
			var x, y float64
			if datum == playerID {
				x = centerWindowX
				y = centerWindowY
			} else {
				x = (float64(i%game.gameMap.Data.Width*game.tileSize) - game.player.X) + centerWindowX
				y = (float64(i/game.gameMap.Data.Height*game.tileSize) - game.player.Y) + centerWindowY
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, y)
			op.GeoM.Scale(game.mapScale, game.mapScale)
			screen.DrawImage(img, op)
		}
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Player %.0fx%.0f", game.player.X, game.player.Y))
}

func (game *Game) Layout(screenWidthPx, screenHeightPx int) (int, int) {
	return screenWidthPx, screenHeightPx
}

func (game *Game) addEvents() {
	game.eventManager.AddPressEvent(ebiten.KeyRight, func() {
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
	game.eventManager.AddPressEvent(ebiten.KeyLeft, func() {
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
	game.eventManager.AddPressEvent(ebiten.KeyUp, func() {
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
	game.eventManager.AddPressEvent(ebiten.KeyDown, func() {
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

func getSubImage(id int, tilesetImage *ebiten.Image, tileSize int) *ebiten.Image {
	row := (id - 1) / 10
	col := (id - 1) % 10
	x := col * tileSize
	y := row * tileSize

	return tilesetImage.SubImage(image.Rect(x, y, x+tileSize, y+tileSize)).(*ebiten.Image)
}
