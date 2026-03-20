package server

import (
	"errors"
	"strconv"
)

func extractPurchaseDayFromName(name string) (int, error) {
	if len(name) < 6 {
		return 0, errors.New("invalid name")
	}

	purchasedDayStr := name[len(name)-2:]
	purchasedDay, err := strconv.Atoi(purchasedDayStr)
	if err != nil {
		return 0, errors.New("not a number")
	}

	if purchasedDay < 1 || purchasedDay > 31 {
		return 0, errors.New("invalid day")
	}

	return purchasedDay, nil
}
