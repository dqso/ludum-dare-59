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

type SizeGetter interface {
	Width() float64
	Height() float64
}

type PositionSetter interface {
	SetPosition(x, y float64)
}

type PivotGetter interface {
	PivotX() float64
	PivotY() float64
	PivotRadius() float64
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
	PivotGetter
	SpeedGetter
	Movable
	Drawable
}
