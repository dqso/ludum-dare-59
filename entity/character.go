package entity

import "time"

type Character interface {
	Positionable
	TopLeftGetter
	PivotGetter
	SpeedGetter
	SizeGetter
	Movable
	Drawable
	Role() CharacterRole
	SetFocus(focus bool)
	IsFocused() bool
	GetQuestion() Question
	AnswerTheQuestion(questionsDatabase QuestionsDatabase, answer Answer)
	Deadline() time.Time
}

type CharacterRole string

const (
	CharacterRoleRecruiter = CharacterRole("recruiter")
	CharacterRoleEngineer  = CharacterRole("engineer")
	CharacterRoleFounder   = CharacterRole("founder")
	//CharacterRoleWolf      = CharacterRole("wolf")
)
