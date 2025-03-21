package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/VxVxN/the_lonely_explorer/internal/game"
)

func main() {
	game, err := game.NewGame()
	if err != nil {
		log.Fatalf("Failed to init game: %v", err)
	}
	defer game.Close()

	ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("The lonely explorer")

	if err = ebiten.RunGame(game); err != nil {
		log.Fatalf("Failed to run game: %v", err)
	}
}
