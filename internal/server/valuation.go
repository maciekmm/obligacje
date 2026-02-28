package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/calculator"
	"github.com/maciekmm/obligacje/tz"
)

type ValuationResponse struct {
	Name       string  `json:"name"`
	ISIN       string  `json:"isin"`
	ValuatedAt string  `json:"valuated_at"`
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
}

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

func (s *Server) handleValuation(w http.ResponseWriter, r *http.Request) {
	var valuatedAt time.Time
	var err error
	if valAtQ := r.URL.Query().Get("valuated_at"); valAtQ != "" {
		valuatedAt, err = time.ParseInLocation("2006-01-02", valAtQ, tz.UnifiedTimezone)
		if err != nil {
			http.Error(w, "invalid valuated_at", http.StatusBadRequest)
			return
		}
	} else {
		valuatedAt = time.Now().In(tz.UnifiedTimezone)
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

	price, err := s.calc.Calculate(bnd, purchaseDay, valuatedAt)
	if errors.Is(err, calculator.ErrValuationDateBeforePurchaseDate) {
		s.log.Info("valuation date before purchase date", "name", name, "purchase_day", purchaseDay, "valuated_at", valuatedAt)
		http.Error(w, "valuation date is before purchase date", http.StatusBadRequest)
		return
	}
	if err != nil && !errors.Is(err, calculator.ErrValuationDateAfterMaturity) {
		s.log.Warn("error calculating price", "name", name, "purchase_day", purchaseDay, "valuated_at", valuatedAt, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.log.Info("valuated bond", "name", name, "purchase_day", purchaseDay, "valuated_at", valuatedAt, "price", price)

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValuationResponse{
			Name:       nameWithPurchaseDay,
			ISIN:       bnd.ISIN,
			ValuatedAt: valuatedAt.Format("2006-01-02"),
			Price:      float64(price),
			Currency:   "PLN",
		})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%.2f", price)
}
