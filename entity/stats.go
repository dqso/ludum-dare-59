package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Stats struct {
	DurationInGame    time.Duration
	RealTimePlayed    time.Duration
	Earned            decimal.Decimal
	Spent             decimal.Decimal
	CurrentBalance    decimal.Decimal
	MonthlySalary     decimal.Decimal
	OffersAccepted    int
	OffersRejected    int
	OffersReceived    int
	QuestionsAnswered int
}
