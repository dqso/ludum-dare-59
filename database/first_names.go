package database

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
)

type FirstNameDatabase struct {
	female            []string
	male              []string
	companies         []string
	shuffledCompanies []string
}

func NewFirstNameDatabase(femaleData, maleData, companiesData []byte) (*FirstNameDatabase, error) {
	db := &FirstNameDatabase{
		female:    make([]string, 0, 5001),
		male:      make([]string, 0, 2943),
		companies: make([]string, 0, 75),
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

	r = bufio.NewReader(bytes.NewReader(companiesData))
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		db.companies = append(db.companies, string(line))
	}

	return db, nil
}

func (db *FirstNameDatabase) GetRandomFemaleFirstName() string {
	return db.female[rand.IntN(len(db.female))]
}

func (db *FirstNameDatabase) GetRandomMaleFirstName() string {
	return db.male[rand.IntN(len(db.male))]
}

func (db *FirstNameDatabase) GetRandomCompanyName() string {
	if len(db.shuffledCompanies) == 0 {
		db.shuffledCompanies = make([]string, len(db.companies))
		copy(db.shuffledCompanies, db.companies)
	}
	rand.Shuffle(len(db.shuffledCompanies), func(i, j int) {
		db.shuffledCompanies[i], db.shuffledCompanies[j] = db.shuffledCompanies[j], db.shuffledCompanies[i]
	})

	var company string
	company, db.shuffledCompanies = db.shuffledCompanies[0], db.shuffledCompanies[1:]
	return company
}
