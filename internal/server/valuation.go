package server

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

func extractPurchasedDayFromName(name string) (int, error) {
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

func (s *Server) handleValuation(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	purchasedDay, err := extractPurchasedDayFromName(name)
	if err != nil {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}

	s.calc.Calculate(name, purchasedDay, time.Now())

}
