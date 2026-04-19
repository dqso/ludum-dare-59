package entity

import "time"

type Token interface {
	Positionable
	TopLeftGetter
	SizeGetter
	Drawable
	PositionSetter

	SetFocus(focus bool)
	IsFocused() bool
	Collect() CollectedToken
	Answer() Answer
	Deadline() time.Time
}

type CollectedToken interface {
	SizeGetter
	Drawable
	Drop(x, y float64) Token
}
