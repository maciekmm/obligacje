package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHistorical_HappyPath(t *testing.T) {
	server := loadTestServer(t)

	tests := []struct {
		name       string
		bondName   string
		from       string
		to         string
		wantCode   int
		wantDays   int
		spotChecks map[string]float64
	}{
		{
			name:     "TOS bond short range",
			bondName: "TOS112501",
			from:     "2023-03-25",
			to:       "2023-03-27",
			wantCode: http.StatusOK,
			wantDays: 3,
			spotChecks: map[string]float64{
				"2023-03-26": 102.72,
				"2023-03-27": 102.74,
			},
		},
		{
			name:     "EDO bond single day",
			bondName: "EDO083412",
			from:     "2025-12-06",
			to:       "2025-12-06",
			wantCode: http.StatusOK,
			wantDays: 1,
			spotChecks: map[string]float64{
				"2025-12-06": 108.87,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/v1/bond/%s/historical?from=%s&to=%s", tt.bondName, tt.from, tt.to)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("got status %d, want %d; body: %s", w.Code, tt.wantCode, w.Body.String())
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("got Content-Type %q, want application/json", ct)
			}

			var resp HistoricalResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode JSON: %v", err)
			}

			if len(resp.Valuations) != tt.wantDays {
				t.Errorf("got %d valuation days, want %d", len(resp.Valuations), tt.wantDays)
			}

			for date, wantPrice := range tt.spotChecks {
				gotPrice, ok := resp.Valuations[date]
				if !ok {
					t.Errorf("missing valuation for date %s", date)
					continue
				}
				if math.Abs(gotPrice-wantPrice) > 1e-9 {
					t.Errorf("date %s: got price %v, want %v", date, gotPrice, wantPrice)
				}
			}
		})
	}
}

func TestHandleHistorical_Errors(t *testing.T) {
	server := loadTestServer(t)

	tests := []struct {
		name     string
		bondName string
		query    string
		wantCode int
	}{
		{
			name:     "missing from",
			bondName: "TOS112501",
			query:    "to=2023-03-27",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing to",
			bondName: "TOS112501",
			query:    "from=2023-03-25",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid from date",
			bondName: "TOS112501",
			query:    "from=not-a-date&to=2023-03-27",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid to date",
			bondName: "TOS112501",
			query:    "from=2023-03-25&to=not-a-date",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "to before from",
			bondName: "TOS112501",
			query:    "from=2023-03-27&to=2023-03-25",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "span exceeds 366 days",
			bondName: "TOS112501",
			query:    "from=2023-01-01&to=2024-01-03",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "bond not found",
			bondName: "NONEXIST01",
			query:    "from=2023-03-25&to=2023-03-27",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "invalid name - non-numeric suffix",
			bondName: "EDO0834XX",
			query:    "from=2023-03-25&to=2023-03-27",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid name - day 0",
			bondName: "EDO083400",
			query:    "from=2023-03-25&to=2023-03-27",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid name - too short",
			bondName: "AB01",
			query:    "from=2023-03-25&to=2023-03-27",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/v1/bond/%s/historical?%s", tt.bondName, tt.query)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantCode, w.Body.String())
			}
		})
	}
}

func TestHandleHistorical_FromBeforePurchaseDate(t *testing.T) {
	server := loadTestServer(t)

	// EDO0935 sale starts in Sept 2025, purchase day 02 → purchase date 2025-09-02
	// Request from before purchase date to after it
	url := "/v1/bond/EDO093502/historical?from=2025-09-01&to=2025-09-05"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp HistoricalResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	// 2025-09-01 should be omitted (before purchase date)
	if _, ok := resp.Valuations["2025-09-01"]; ok {
		t.Error("expected date before purchase to be omitted, but it was present")
	}

	// 2025-09-02 onwards should be present
	if _, ok := resp.Valuations["2025-09-02"]; !ok {
		t.Error("expected purchase date to be present in valuations")
	}
}

func TestHandleHistorical_AfterMaturity(t *testing.T) {
	server := loadTestServer(t)

	// TOS1125 matures after ~3 years from Nov 2022, purchase day 01
	// Request a range that spans beyond maturity
	url := "/v1/bond/TOS112501/historical?from=2025-11-30&to=2025-12-02"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp HistoricalResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	// After maturity days should still be present (same behavior as valuation endpoint)
	if len(resp.Valuations) == 0 {
		t.Error("expected valuations to be present even after maturity")
	}
}

func TestHandleHistorical_ExactlyMaxSpan(t *testing.T) {
	_ = slog.Default() // ensure the test server logger is available
	server := loadTestServer(t)

	// Exactly 366 days should succeed
	url := "/v1/bond/TOS112501/historical?from=2023-01-01&to=2024-01-02"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp HistoricalResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if len(resp.Valuations) == 0 {
		t.Error("expected non-empty valuations for max span")
	}
}
