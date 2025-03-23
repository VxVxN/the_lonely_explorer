package player

import (
	"github.com/VxVxN/gamedevlib/rectangle"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	name string
	*rectangle.Rectangle
	speed float64
	image *ebiten.Image
	dead  bool
}

func NewPlayer(image *ebiten.Image, speed float64) *Player {
	return &Player{
		speed:     speed,
		Rectangle: rectangle.New(0, 0, float64(image.Bounds().Dx()), float64(image.Bounds().Dy())),
		image:     image,
	}
}

func (player *Player) Update() {
}

func (player *Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(player.X, player.Y)
	screen.DrawImage(player.image, op)
}

func (player *Player) SetPosition(x, y float64) {
	player.X = x
	player.Y = y
}

func (player *Player) Move(key ebiten.Key) {
	switch key {
	case ebiten.KeyLeft:
		player.X -= player.speed
	case ebiten.KeyRight:
		player.X += player.speed
	case ebiten.KeyUp:
		player.Y -= player.speed
	case ebiten.KeyDown:
		player.Y += player.speed
	default:
	}
}

func (player *Player) Reset() {
	player.dead = false
}

func (player *Player) SetName(name string) {
	player.name = name
}

func (player *Player) Name() string {
	return player.name
}

func (player *Player) SetDead(dead bool) {
	player.dead = dead
}

func (player *Player) Dead() bool {
	return player.dead
}

func (player *Player) SetSpeed(speed float64) {
	player.speed = speed
}

func (player *Player) Speed() float64 {
	return player.speed
}
