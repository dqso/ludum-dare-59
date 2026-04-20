package entity

import (
	"image/color"

	"golang.org/x/image/colornames"
)

type QuestionsForInterview struct {
	Recruiter QuestionsDatabase
	Engineer  QuestionsDatabase
}

type FirstNamesDatabase interface {
	GetRandomFemaleFirstName() string
	GetRandomMaleFirstName() string
}

func (q QuestionsForInterview) Choose(roleGetter interface {
	Role() CharacterRole
}) QuestionsDatabase {
	switch roleGetter.Role() {
	case CharacterRoleRecruiter:
		return q.Recruiter
	case CharacterRoleEngineer:
		return q.Engineer
	case CharacterRoleFounder:
		return q.Recruiter // TODO
	default:
		return q.Recruiter // TODO
	}
}

type AnswersGrouped struct {
	Positive map[Answer]struct{}
	Neutral  map[Answer]struct{}
	Negative map[Answer]struct{}
}

func NewAnswersGrouped() AnswersGrouped {
	return AnswersGrouped{
		Positive: make(map[Answer]struct{}),
		Neutral:  make(map[Answer]struct{}),
		Negative: make(map[Answer]struct{}),
	}
}

type QuestionsDatabase interface {
	GetRandomQuestions(num int) ([]Question, AnswersGrouped)
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
	AnswerCategoryEvasive     AnswerCategory = "evasive"   // уклончивые
	AnswerCategoryBinaryYes   AnswerCategory = "yes"       // бинарные: да
	AnswerCategoryBinaryNo    AnswerCategory = "no"        // бинарные: нет
	AnswerCategoryFrequency   AnswerCategory = "frequency" // частотные
	AnswerCategoryEmotional   AnswerCategory = "emotional" // эмоциональные
	AnswerCategoryQuantity    AnswerCategory = "quantity"  // количество или время
	AnswerCategoryProgramming AnswerCategory = "programming"
)

func AnswerCategoryToColor(_type AnswerCategory) color.Color {
	switch _type {
	case AnswerCategoryEvasive:
		return colornames.Ivory
	case AnswerCategoryBinaryYes, AnswerCategoryBinaryNo:
		return colornames.Springgreen
	case AnswerCategoryFrequency:
		return colornames.Honeydew
	case AnswerCategoryEmotional:
		return colornames.Purple
	case AnswerCategoryQuantity:
		return colornames.Gold
	case AnswerCategoryProgramming:
		return colornames.Lightblue
	default:
		return colornames.Red
	}
}

func (t AnswerCategory) IsValid() bool {
	switch t {
	case AnswerCategoryEvasive:
	case AnswerCategoryBinaryYes, AnswerCategoryBinaryNo:
	case AnswerCategoryFrequency:
	case AnswerCategoryEmotional:
	case AnswerCategoryQuantity:
	case AnswerCategoryProgramming:
	default:
		return false
	}
	return true
}
