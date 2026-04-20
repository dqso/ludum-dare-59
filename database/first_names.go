package database

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
)

type FirstNameDatabase struct {
	female []string
	male   []string
}

func NewFirstNameDatabase(femaleData, maleData []byte) (*FirstNameDatabase, error) {
	db := &FirstNameDatabase{
		female: make([]string, 0, 5001),
		male:   make([]string, 0, 2943),
	}

	r := bufio.NewReader(bytes.NewReader(femaleData))
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		db.female = append(db.female, string(line))
	}

	r = bufio.NewReader(bytes.NewReader(maleData))
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		db.male = append(db.male, string(line))
	}

	return db, nil
}

func (db *FirstNameDatabase) GetRandomFemaleFirstName() string {
	return db.female[rand.IntN(len(db.female))]
}

func (db *FirstNameDatabase) GetRandomMaleFirstName() string {
	return db.male[rand.IntN(len(db.male))]
}
