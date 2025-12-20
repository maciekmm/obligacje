package server

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

func extractPurchasedDayFromSeries(series string) (int, error) {
	if len(series) < 6 {
		return 0, errors.New("invalid series")
	}

	purchasedDayStr := series[len(series)-2:]
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
	series := r.PathValue("series")

	purchasedDay, err := extractPurchasedDayFromSeries(series)
	if err != nil {
		http.Error(w, "invalid series", http.StatusBadRequest)
		return
	}

	s.calc.Calculate(series, purchasedAt, time.Now())

}
