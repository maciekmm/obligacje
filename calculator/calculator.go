package calculator

import (
	"errors"
	"math"
	"time"

	"github.com/maciekmm/obligacje/bond"
)

type Calculator struct {
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(bnd bond.Bond, purchasedAt time.Time, valuatedAt time.Time) (bond.Price, error) {
	if purchasedAt.Before(bnd.SaleStart) || purchasedAt.After(bnd.SaleEnd) {
		return 0, errors.New("bond must be purchased during sale period")
	}
	if valuatedAt.Before(purchasedAt) {
		return 0, errors.New("bond must be valuated after purchase")
	}

	price := float64(bnd.FaceValue)
	for i, perc := range bnd.InterestPeriods {
		start, end, err := bnd.Period(i, purchasedAt.Day())
		if err != nil {
			return 0, err
		}

		if valuatedAt.Before(start) {
			break
		}

		if valuatedAt.After(end) || valuatedAt.Equal(end) {
			price = float64(price) * (1.0 + float64(perc)/float64(bnd.CouponPaymentsFrequency))
		} else {
			periodDays := int(math.Round(end.Sub(start).Hours() / 24))
			heldDays := int(math.Round(valuatedAt.Sub(start).Hours() / 24))
			price = float64(price) * (1.0 + float64(perc)/float64(bnd.CouponPaymentsFrequency)*float64(heldDays)/float64(periodDays))
		}

		if len(bnd.InterestPeriods) == i+1 && valuatedAt.After(end) {
			return bond.Price(math.Round(price*100.0) / 100.0), errors.New("valuation date is after last known interest period")
		}
	}

	return bond.Price(math.Round(price*100.0) / 100.0), nil
}
