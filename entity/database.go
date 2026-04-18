package entity

type QuestionsDatabase interface {
	GetRandomQuestions(num int) []string
}
