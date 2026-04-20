package database

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"

	"github.com/dqso/ludum-dare-59/entity"
)

type QuestionsDatabase struct {
	byQuestion map[string]*QuestionWithAnswers
	questions  []Question
	answers    []Answer
}

// --------------------------------------------------

type Question struct {
	question string
	// TODO list type_answer
}

func (q Question) Question() string { return q.question }

// --------------------------------------------------

type Answer struct {
	answer   string
	category entity.AnswerCategory
}

func (a Answer) Answer() string                  { return a.answer }
func (a Answer) Category() entity.AnswerCategory { return a.category }

// --------------------------------------------------

type QuestionWithAnswers struct {
	question  Question
	byAnswers map[string]*QuestionAnswer
}

// --------------------------------------------------

type QuestionAnswer struct {
	question Question
	answer   Answer
	points   int8
}

func (a QuestionAnswer) Question() entity.Question { return a.question }
func (a QuestionAnswer) Answer() entity.Answer     { return a.answer }
func (a QuestionAnswer) Points() int8              { return a.points }

func NewQuestionsDatabase(csvData []byte) (*QuestionsDatabase, error) {
	db := &QuestionsDatabase{
		byQuestion: make(map[string]*QuestionWithAnswers),
		questions:  make([]Question, 0),
		answers:    make([]Answer, 0),
	}

	r := csv.NewReader(bytes.NewReader(csvData))
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to parse header of CSV: %w", err)
	}
	questions := header[2:]
	for _, q := range questions {
		question := Question{
			question: q,
		}
		db.byQuestion[question.question] = &QuestionWithAnswers{
			question:  question,
			byAnswers: make(map[string]*QuestionAnswer),
		}
		db.questions = append(db.questions, question)
	}
	var idxAnswer int
	for values, err := r.Read(); !errors.Is(err, io.EOF); values, err = r.Read() {
		if err != nil {
			return nil, fmt.Errorf("failed to parse answer %d of CSV: %w", idxAnswer, err)
		}
		if len(values) != len(header) {
			return nil, fmt.Errorf("failed to parse answer %d of CSV: size is not %d", idxAnswer, len(questions))
		}
		answer := Answer{
			answer:   values[0],
			category: entity.AnswerCategory(values[1]),
		}
		if !answer.category.IsValid() {
			return nil, fmt.Errorf("category of answer %d is not valid", idxAnswer)
		}
		for i := 2; i < len(values); i++ {
			question := questions[i-2]
			points, err := strconv.ParseInt(values[i], 10, 8)
			if err != nil {
				return nil, fmt.Errorf("failed to parse points for question %q and for answer %q: %w", question, answer.Answer(), err)
			}
			if points == 0 {
				continue
			}
			qa := &QuestionAnswer{
				question: Question{
					question: question,
				},
				answer: answer,
				points: int8(points),
			}
			db.byQuestion[question].byAnswers[answer.Answer()] = qa
		}
		db.answers = append(db.answers, answer)
	}

	return db, nil
}

func (d *QuestionsDatabase) GetRandomQuestions(num int) ([]entity.Question, entity.AnswersGrouped) {
	answers := entity.NewAnswersGrouped()
	if num == 0 {
		return make([]entity.Question, 0), answers
	}
	list := make([]entity.Question, len(d.questions))
	for i, q := range d.questions {
		list[i] = q
	}
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	if len(list) > num {
		list = list[:num]
	}
	for _, e := range list {
		q := d.byQuestion[e.Question()]
		for _, a := range q.byAnswers {
			if a.points > 0 {
				answers.Positive[a.answer] = struct{}{}
			} else if a.points == 0 {
				answers.Neutral[a.answer] = struct{}{}
			} else {
				answers.Negative[a.answer] = struct{}{}
			}
		}
	}
	return list, answers
}

func (d *QuestionsDatabase) GetRandomAnswers(num int) []entity.Answer {
	if num == 0 {
		return make([]entity.Answer, 0)
	}
	list := make([]entity.Answer, len(d.answers))
	for i, a := range d.answers {
		list[i] = a
	}
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	if len(list) > num {
		return list[:num]
	}
	return list
}

func (d *QuestionsDatabase) Match(question entity.Question, answer entity.Answer) (int, error) {
	qa, ok := d.byQuestion[question.Question()]
	if !ok {
		return 0, fmt.Errorf("question not found")
	}
	a, ok := qa.byAnswers[answer.Answer()]
	if !ok {
		return 0, nil
	}
	return int(a.points), nil
}
