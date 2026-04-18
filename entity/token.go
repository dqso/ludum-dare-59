package entity

type Token interface {
	Positionable
	TopLeftGetter
	Drawable
}
