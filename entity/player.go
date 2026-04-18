package entity

import "github.com/hajimehoshi/ebiten/v2"

type Positionable interface {
	X() float64
	Y() float64
}

type TopLeftGetter interface {
	TopLeftX() float64
	TopLeftY() float64
}

type SpeedGetter interface {
	Speed() float64
}

type Movable interface {
	Move(dx, dy float64)
}

type Drawable interface {
	Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions)
}

type Player interface {
	Positionable
	TopLeftGetter
	SpeedGetter
	Movable
	Drawable
}
