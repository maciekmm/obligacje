package server

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/calculator"
	"github.com/maciekmm/obligacje/tz"
)

const maxHistoricalDays = 366

type HistoricalResponse struct {
	Valuations map[string]float64 `json:"valuations"`
}

func (s *Server) handleHistorical(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		http.Error(w, "from and to parameters are required", http.StatusBadRequest)
		return
	}

	from, err := time.ParseInLocation("2006-01-02", fromStr, tz.UnifiedTimezone)
	if err != nil {
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return
	}

	to, err := time.ParseInLocation("2006-01-02", toStr, tz.UnifiedTimezone)
	if err != nil {
		http.Error(w, "invalid to date", http.StatusBadRequest)
		return
	}

	if to.Before(from) {
		http.Error(w, "to must not be before from", http.StatusBadRequest)
		return
	}

	if int(math.Round(to.Sub(from).Hours()/24)) > maxHistoricalDays {
		http.Error(w, "date range must not exceed 366 days", http.StatusBadRequest)
		return
	}

	nameWithPurchaseDay := r.PathValue("name")
	purchaseDay, err := extractPurchaseDayFromName(nameWithPurchaseDay)
	if err != nil {
		s.log.Info("invalid name", "name", nameWithPurchaseDay, "err", err)
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	name := nameWithPurchaseDay[:len(nameWithPurchaseDay)-2]

	bnd, err := s.repo.Lookup(name)
	if errors.Is(err, bond.ErrNameNotFound) {
		s.log.Info("bond not found", "name", name)
		http.Error(w, "invalid name", http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Info("error looking up bond", "name", name, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	valuations := make(map[string]float64)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		price, err := s.calc.Calculate(bnd, purchaseDay, d)
		if errors.Is(err, calculator.ErrValuationDateBeforePurchaseDate) {
			continue
		}
		if err != nil && !errors.Is(err, calculator.ErrValuationDateAfterMaturity) {
			s.log.Warn("error calculating price", "name", name, "purchase_day", purchaseDay, "date", d, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		valuations[d.Format("2006-01-02")] = float64(price)
	}

	s.log.Info("historical valuation", "name", name, "purchase_day", purchaseDay, "from", from, "to", to, "days", len(valuations))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HistoricalResponse{
		Valuations: valuations,
	})
}
