package entity

import (
	"image/color"

	"golang.org/x/image/colornames"
)

type QuestionsDatabase interface {
	GetRandomQuestions(num int) []Question
	GetRandomAnswers(num int) []Answer
	Match(question Question, answer Answer) (int, error)
}

type Question interface {
	Question() string
}

type Answer interface {
	Answer() string
	Category() AnswerCategory
}

type AnswerCategory string

const (
	AnswerCategoryEvasive   AnswerCategory = "evasive"   // уклончивые
	AnswerCategoryBinary    AnswerCategory = "binary"    // бинарные
	AnswerCategoryFrequency AnswerCategory = "frequency" // частотные
	AnswerCategoryEmotional AnswerCategory = "emotional" // эмоциональные
)

func AnswerCategoryToColor(_type AnswerCategory) color.Color {
	switch _type {
	case AnswerCategoryEvasive:
		return colornames.Ivory
	case AnswerCategoryBinary:
		return colornames.Springgreen
	case AnswerCategoryFrequency:
		return colornames.Honeydew
	case AnswerCategoryEmotional:
		return colornames.Purple
	default:
		return colornames.Red
	}
}

func (t AnswerCategory) IsValid() bool {
	switch t {
	case AnswerCategoryEvasive:
	case AnswerCategoryBinary:
	case AnswerCategoryFrequency:
	case AnswerCategoryEmotional:
	default:
		return false
	}
	return true
}
