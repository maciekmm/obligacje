package bond

import (
	"errors"
)

var (
	ErrSeriesNotFound = errors.New("series not found")
)

type Repository interface {
	Lookup(series string) (Bond, error)
}
