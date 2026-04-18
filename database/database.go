package database

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"
)

type QuestionsDatabase struct {
	Database map[string]map[string]int8
}

func NewQuestionsDatabase(csvData []byte) (*QuestionsDatabase, error) {
	db := &QuestionsDatabase{
		Database: make(map[string]map[string]int8),
	}

	r := csv.NewReader(bytes.NewReader(csvData))
	questions, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to parse header of CSV: %w", err)
	}
	questions = questions[1:]
	for _, question := range questions {
		db.Database[question] = make(map[string]int8)
	}
	var idxAnswer int
	for values, err := r.Read(); !errors.Is(err, io.EOF); values, err = r.Read() {
		if err != nil {
			return nil, fmt.Errorf("failed to parse answer %d of CSV: %w", idxAnswer, err)
		}
		if len(values) != len(questions)+1 {
			return nil, fmt.Errorf("failed to parse answer %d of CSV: size is not %d", idxAnswer, len(questions)+1)
		}
		answer := values[0]
		for i := 1; i < len(values); i++ {
			question := questions[i-1]
			points, err := strconv.ParseInt(values[i], 10, 8)
			if err != nil {
				return nil, fmt.Errorf("failed to parse points for question %q and for answer %q: %w", question, answer, err)
			}
			if points == 0 {
				continue
			}
			db.Database[question][answer] = int8(points)
		}
	}

	return db, nil
}

func (d *QuestionsDatabase) GetRandomQuestions(num int) []string {
	list := make([]string, len(d.Database))
	for k := range d.Database {
		list = append(list, k)
	}
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	if len(list) > num {
		return list[:num]
	}
	return list
}
