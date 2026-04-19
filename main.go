package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dqso/ludum-dare-59/game"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	windowWidth  = 1024
	windowHeight = 768
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	g := game.New(ctx)

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeOnlyFullscreenEnabled)
	ebiten.SetWindowTitle("Game for Ludum Dare 59")

	if err := ebiten.RunGame(g); err != nil {
		log.Print(err)
		return 1
	}

	return 0
}
