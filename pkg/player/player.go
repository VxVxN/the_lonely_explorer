package player

import (
	"github.com/VxVxN/gamedevlib/animation"
	"github.com/VxVxN/gamedevlib/rectangle"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	name string
	*rectangle.Rectangle
	speed                  float64
	animationSpeed         float64
	image                  *ebiten.Image
	playerForwardAnimation *animation.Animation
	playerBackAnimation    *animation.Animation
	playerLeftAnimation    *animation.Animation
	playerRightAnimation   *animation.Animation
	lastKey                ebiten.Key
	scale                  float64
	dead                   bool
}

func NewPlayer(image *ebiten.Image, playerForwardAnimation, playerBackAnimation, playerLeftAnimation, playerRightAnimation *animation.Animation, speed float64) *Player {
	return &Player{
		speed:                  speed,
		Rectangle:              rectangle.New(0, 0, float64(image.Bounds().Dx()), float64(image.Bounds().Dy())),
		image:                  image,
		playerForwardAnimation: playerForwardAnimation,
		playerBackAnimation:    playerBackAnimation,
		playerLeftAnimation:    playerLeftAnimation,
		playerRightAnimation:   playerRightAnimation,
		scale:                  1.0,
		animationSpeed:         0.1,
	}
}

func (player *Player) Update() {
	player.playerForwardAnimation.Update(player.animationSpeed)
	player.playerBackAnimation.Update(player.animationSpeed)
	player.playerLeftAnimation.Update(player.animationSpeed)
	player.playerRightAnimation.Update(player.animationSpeed)
}

func (player *Player) Draw(screen *ebiten.Image, x, y float64) {
	switch player.lastKey {
	case ebiten.KeyDown:
		player.playerForwardAnimation.Start()
		player.playerForwardAnimation.SetPosition(x, y)
		player.playerForwardAnimation.Draw(screen)
	case ebiten.KeyUp:
		player.playerBackAnimation.Start()
		player.playerBackAnimation.SetPosition(x, y)
		player.playerBackAnimation.Draw(screen)
	case ebiten.KeyLeft:
		player.playerLeftAnimation.Start()
		player.playerLeftAnimation.SetPosition(x, y)
		player.playerLeftAnimation.Draw(screen)
	case ebiten.KeyRight:
		player.playerRightAnimation.Start()
		player.playerRightAnimation.SetPosition(x, y)
		player.playerRightAnimation.Draw(screen)
	default:
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		op.GeoM.Scale(player.scale, player.scale)
		screen.DrawImage(player.image, op)
	}
}

func (player *Player) SetPosition(x, y float64) {
	player.X = x
	player.Y = y
}

func (player *Player) Move(key ebiten.Key) {
	player.lastKey = key
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

func (player *Player) SetScale(scale float64) {
	player.scale = scale
}
