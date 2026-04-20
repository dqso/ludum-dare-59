package entity

import "github.com/shopspring/decimal"

type InterviewResult interface {
	Outcome() Outcome
	Salary() decimal.Decimal
}

type Outcome string

const (
	OutcomeRejection   Outcome = "rejection"
	OutcomeRoundPassed Outcome = "round passed"
	OutcomeOffer       Outcome = "offer"
)
