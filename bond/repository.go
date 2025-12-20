package bond

import (
	"errors"
)

var (
	ErrNameNotFound = errors.New("name not found")
)

type Repository interface {
	Lookup(name string) (Bond, error)
}
