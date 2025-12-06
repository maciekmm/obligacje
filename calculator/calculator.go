package calculator

import (
	"errors"
	"time"

	"github.com/maciekmm/obligacje/bond"
)

type Calculator struct {
	repo bond.Repository
}

func NewCalculator(repo bond.Repository) *Calculator {
	return &Calculator{
		repo: repo,
	}
}

func (c *Calculator) Calculate(series string, purchasedAt time.Time, valuatedAt time.Time) (bond.Price, error) {
	return 0, errors.New("not implemented")
}
