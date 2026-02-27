package calculator

import (
	"errors"
	"math"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/tz"
)

var (
	ErrValuationDateBeforePurchaseDate = errors.New("valuation date is before purchase date")
	ErrValuationDateAfterMaturity      = errors.New("valuation date is after last known interest period")
)

type Calculator struct {
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(bnd bond.Bond, purchaseDay int, valuatedAt time.Time) (bond.Price, error) {
	purchaseDate := time.Date(bnd.SaleStart.Year(), bnd.SaleStart.Month(), purchaseDay, 0, 0, 0, 0, tz.UnifiedTimezone)
	if valuatedAt.Before(purchaseDate) {
		return 0, ErrValuationDateBeforePurchaseDate
	}

	price := float64(bnd.FaceValue)
	for i, perc := range bnd.InterestPeriods {
		start, end, err := bnd.Period(i, purchaseDay)
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
			return bond.Price(math.Round(price*100.0) / 100.0), ErrValuationDateAfterMaturity
		}
	}

	return bond.Price(math.Round(price*100.0) / 100.0), nil
}
