package character

import "github.com/dqso/ludum-dare-59/entity"

type Rejection struct{}

func NewRejection() Rejection {
	return Rejection{}
}

func (r Rejection) Outcome() entity.Outcome { return entity.OutcomeRejection }
func (r Rejection) Salary() int             { return 0 }

type RoundPassed struct{}

func NewRoundPassed() RoundPassed {
	return RoundPassed{}
}

func (r RoundPassed) Outcome() entity.Outcome { return entity.OutcomeRoundPassed }
func (r RoundPassed) Salary() int             { return 0 }

type Offer struct {
	salary int
}

func NewOffer(salary int) Offer {
	return Offer{
		salary: salary,
	}
}

func (o Offer) Outcome() entity.Outcome { return entity.OutcomeOffer }
func (o Offer) Salary() int             { return o.salary }
