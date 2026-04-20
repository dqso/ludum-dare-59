package entity

import (
	"iter"
	"time"
)

type Character interface {
	Name() string
	Positionable
	TopLeftGetter
	PivotGetter
	SpeedGetter
	SizeGetter
	Movable
	Drawable
	Company() string
	SetCompany(company string)
	Role() CharacterRole
	SetFocus(focus bool)
	IsFocused() bool
	GetQuestion() Question
	AnswerTheQuestion(questionsDatabase QuestionsDatabase, answer Answer)
	Deadline() time.Time
	UpdateDeadline(delay time.Duration)
	InterviewResult() InterviewResult
	SetInterviewResult(offer InterviewResult)
	PlayerPoints() int
	SetPlayerPoints(playerPoints int)
}

type CharacterList interface {
	Add(c Character)
	All() iter.Seq[Character]
	FilterByRoles(roles ...CharacterRole) iter.Seq[Character]
	FilterByCompany(company string) iter.Seq[Character]
	FilterFunc(fn func(c Character) bool) iter.Seq[Character]
	DeleteFunc(fn func(c Character) bool) iter.Seq2[Character, bool]
}

type CharacterRole string

const (
	CharacterRoleRecruiter = CharacterRole("recruiter")
	CharacterRoleEngineer  = CharacterRole("engineer")
	CharacterRoleFounder   = CharacterRole("founder")
	//CharacterRoleWolf      = CharacterRole("wolf")
)
