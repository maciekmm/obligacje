package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maciekmm/obligacje/bond"
)

type MetadataResponse struct {
	Name                    string    `json:"name"`
	ISIN                    string    `json:"isin"`
	FaceValue               float64   `json:"face_value"`
	MonthsToMaturity        int       `json:"months_to_maturity"`
	ExchangePrice           float64   `json:"exchange_price"`
	Margin                  float64   `json:"margin"`
	InterestPeriods         []float64 `json:"interest_periods"`
	CouponPaymentsFrequency int       `json:"coupon_payments_frequency"`
	SaleStart               string    `json:"sale_start"`
	SaleEnd                 string    `json:"sale_end"`
	MaturityDate            string    `json:"maturity_date,omitempty"`
}

func (s *Server) handleMetadata(w http.ResponseWriter, r *http.Request) {
	var (
		purchaseDay int
		name        string
		err         error
	)
	name = r.PathValue("name")
	if len(name) > 7 {
		purchaseDay, err = extractPurchaseDayFromName(name)
		if err != nil {
			s.log.Info("invalid name", "name", name, "err", err)
			http.Error(w, "invalid name", http.StatusBadRequest)
			return
		}
		name = name[:len(name)-2]
	}

	bnd, err := s.repo.Lookup(name)
	if errors.Is(err, bond.ErrNameNotFound) {
		s.log.Info("bond not found", "name", name)
		http.Error(w, "bond not found", http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Info("error looking up bond", "name", name, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	interestPeriods := make([]float64, len(bnd.InterestPeriods))
	for i, p := range bnd.InterestPeriods {
		interestPeriods[i] = float64(p)
	}

	resp := MetadataResponse{
		Name:                    bnd.Name,
		ISIN:                    bnd.ISIN,
		FaceValue:               float64(bnd.FaceValue),
		MonthsToMaturity:        bnd.MonthsToMaturity,
		ExchangePrice:           float64(bnd.ExchangePrice),
		Margin:                  float64(bnd.Margin),
		InterestPeriods:         interestPeriods,
		CouponPaymentsFrequency: int(bnd.CouponPaymentsFrequency),
		SaleStart:               bnd.SaleStart.Format("2006-01-02"),
		SaleEnd:                 bnd.SaleEnd.Format("2006-01-02"),
	}

	if purchaseDay > 0 {
		_, endAt, err := bnd.Period(bnd.InterestPeriodCount()-1, purchaseDay)
		if err == nil {
			resp.MaturityDate = endAt.Format("2006-01-02")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
