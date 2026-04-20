package character

import (
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/shopspring/decimal"
)

type Rejection struct{}

func NewRejection() Rejection {
	return Rejection{}
}

func (r Rejection) Outcome() entity.Outcome { return entity.OutcomeRejection }
func (r Rejection) Salary() decimal.Decimal { return decimal.Zero }

type RoundPassed struct{}

func NewRoundPassed() RoundPassed {
	return RoundPassed{}
}

func (r RoundPassed) Outcome() entity.Outcome { return entity.OutcomeRoundPassed }
func (r RoundPassed) Salary() decimal.Decimal { return decimal.Zero }

type Offer struct {
	salary decimal.Decimal
}

func NewOffer(salary decimal.Decimal) Offer {
	return Offer{
		salary: salary,
	}
}

func (o Offer) Outcome() entity.Outcome { return entity.OutcomeOffer }
func (o Offer) Salary() decimal.Decimal { return o.salary }
