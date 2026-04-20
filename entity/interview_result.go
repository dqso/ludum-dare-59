package entity

type InterviewResult interface {
	Outcome() Outcome
	Salary() int
}

type Outcome string

const (
	OutcomeRejection   Outcome = "rejection"
	OutcomeRoundPassed Outcome = "round passed"
	OutcomeOffer       Outcome = "offer"
)
