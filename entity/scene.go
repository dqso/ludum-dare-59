package entity

import "github.com/hajimehoshi/ebiten/v2"

type Scene interface {
	Name() string
	Update() (Scene, error)
	Draw(screen *ebiten.Image)
}
