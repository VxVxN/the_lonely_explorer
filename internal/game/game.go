package game

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path"

	"github.com/VxVxN/gamedevlib/animation"
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

	imagesByObjID              map[int]*ebiten.Image
	animationByObjID           map[int]*animation.Animation
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

		imagesByObjID:    make(map[int]*ebiten.Image),
		animationByObjID: make(map[int]*animation.Animation),

		gameMap:      gameMap,
		mapScale:     1.5,
		eventManager: eventmanager.NewEventManager(supportedKeys),
		stager:       stager.New(),

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

	plantAnimation := animation.NewAnimation([]*ebiten.Image{game.imagesByObjID[plant1ID], game.imagesByObjID[plant12D], game.imagesByObjID[plant13D], game.imagesByObjID[plant14D]})
	plantAnimation.SetScale(game.mapScale, game.mapScale)
	plantAnimation.SetReverse(true)
	plantAnimation.SetRepeatable(true)

	game.animationByObjID[plant1ID] = plantAnimation

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
		if tile != playerForward1ID {
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

		for _, animation := range game.animationByObjID {
			animation.Update(0.05)
		}
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
			img, ok := game.imagesByObjID[datum]
			if !ok {
				//game.logger.Error("Unknown layer", "image", datum)
				continue
			}

			centerWindowX := (game.windowWidth/2 - float64(game.tileSize)/2) / game.mapScale
			centerWindowY := (game.windowHeight/2 - float64(game.tileSize)/2) / game.mapScale
			var x, y float64
			if datum == playerForward1ID {
				x = centerWindowX
				y = centerWindowY
				game.player.Draw(screen, x, y)
				continue
			} else {
				x = (float64(i%game.gameMap.Data.Width*game.tileSize) - game.player.X) + centerWindowX
				y = (float64(i/game.gameMap.Data.Height*game.tileSize) - game.player.Y) + centerWindowY
				animation, ok := game.animationByObjID[datum]
				if ok {
					animation.Start()
					animation.SetPosition(x, y)
					animation.Draw(screen)	
					continue
				}
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
	game.eventManager.SetDefaultEvent(func() {
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
