package entity

type Token interface {
	Positionable
	TopLeftGetter
	SizeGetter
	Drawable

	SetFocus(focus bool)
	IsFocused() bool
}
