package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
}

func NewGame() (*Game, error) {
	return &Game{}, nil
}

func (game *Game) Update() error {
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
}

func (game *Game) Layout(screenWidthPx, screenHeightPx int) (int, int) {
	return screenWidthPx, screenHeightPx
}

func (game *Game) Close() {}
