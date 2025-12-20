package calculator

import (
	"errors"
	"time"

	"github.com/maciekmm/obligacje/bond"
)

type Calculator struct {
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(bondSeries bond.Bond, purchasedAt time.Time, valuatedAt time.Time) (bond.Price, error) {
	if purchasedAt.Before(bondSeries.SaleStart) || purchasedAt.After(bondSeries.SaleEnd) {
		return 0, errors.New("bond must be purchased during sale period")
	}
	return 0, nil
}
