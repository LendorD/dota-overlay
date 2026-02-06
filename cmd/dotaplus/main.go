package main

import (
	"log"

	"overlay/internal/dotaplus"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	app := dotaplus.New()

	ebiten.SetWindowSize(dotaplus.ViewWidth, dotaplus.ViewHeight)
	ebiten.SetWindowTitle("Dota Plus")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetScreenTransparent(true)

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
