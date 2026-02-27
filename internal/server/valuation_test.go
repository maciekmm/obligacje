package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/maciekmm/obligacje/bondxls"
	"github.com/maciekmm/obligacje/internal/testutil"
)

var (
	testServerOnce sync.Once
	testServer     *Server
)

func loadTestServer(t *testing.T) *Server {
	t.Helper()
	testServerOnce.Do(func() {
		xlsxFile := filepath.Join(testutil.TestDataDirectory(), "data.xlsx")
		repo, err := bondxls.LoadFromXLSX(slog.Default(), xlsxFile)
		if err != nil {
			t.Fatalf("failed to load bond repository: %v", err)
		}
		testServer = NewServer(repo, slog.Default())
	})
	return testServer
}

func TestHandleValuation_PlainText(t *testing.T) {
	server := loadTestServer(t)

	tests := []struct {
		name      string
		bondName  string
		valuateAt string
		wantPrice float64
		wantCode  int
	}{
		{
			name:      "TOS bond before DST change",
			bondName:  "TOS112501",
			valuateAt: "2023-03-26",
			wantPrice: 102.72,
			wantCode:  http.StatusOK,
		},
		{
			name:      "EDO bond with purchase day 12",
			bondName:  "EDO083412",
			valuateAt: "2025-12-06",
			wantPrice: 108.87,
			wantCode:  http.StatusOK,
		},
		{
			name:      "EDO bond two interest periods",
			bondName:  "EDO083220",
			valuateAt: "2024-08-09",
			wantPrice: 119.95,
			wantCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/v1/bond/%s/valuation?valuated_at=%s", tt.bondName, tt.valuateAt)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Accept", "text/plain")
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("got status %d, want %d; body: %s", w.Code, tt.wantCode, w.Body.String())
			}

			wantBody := fmt.Sprintf("%.2f", tt.wantPrice)
			if w.Body.String() != wantBody {
				t.Errorf("got body %q, want %q", w.Body.String(), wantBody)
			}

			if ct := w.Header().Get("Content-Type"); ct != "text/plain" {
				t.Errorf("got Content-Type %q, want text/plain", ct)
			}
		})
	}
}

func TestHandleValuation_JSON(t *testing.T) {
	server := loadTestServer(t)

	tests := []struct {
		name      string
		bondName  string
		valuateAt string
		wantPrice float64
		wantCode  int
	}{
		{
			name:      "TOS bond after DST change",
			bondName:  "TOS112501",
			valuateAt: "2023-03-27",
			wantPrice: 102.74,
			wantCode:  http.StatusOK,
		},
		{
			name:      "EDO bond first day of month",
			bondName:  "EDO083401",
			valuateAt: "2025-08-01",
			wantPrice: 106.80,
			wantCode:  http.StatusOK,
		},
		{
			name:      "EDO0935 known valuation",
			bondName:  "EDO093502",
			valuateAt: "2025-12-06",
			wantPrice: 101.56,
			wantCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/v1/bond/%s/valuation?valuated_at=%s", tt.bondName, tt.valuateAt)
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

			var resp ValuationResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode JSON response: %v", err)
			}

			if resp.Name != tt.bondName {
				t.Errorf("got name %q, want %q", resp.Name, tt.bondName)
			}
			if resp.ValuatedAt != tt.valuateAt {
				t.Errorf("got valuated_at %q, want %q", resp.ValuatedAt, tt.valuateAt)
			}
			if math.Abs(resp.Price-tt.wantPrice) > 1e-9 {
				t.Errorf("got price %v, want %v", resp.Price, tt.wantPrice)
			}
			if resp.Currency != "PLN" {
				t.Errorf("got currency %q, want PLN", resp.Currency)
			}
			if resp.ISIN == "" {
				t.Error("expected non-empty ISIN")
			}
		})
	}
}

func TestHandleValuation_Errors(t *testing.T) {
	server := loadTestServer(t)

	tests := []struct {
		name     string
		bondName string
		query    string
		accept   string
		wantCode int
	}{
		{
			name:     "invalid purchase day - non-numeric suffix",
			bondName: "EDO0834XX",
			query:    "valuated_at=2025-12-06",
			accept:   "text/plain",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid purchase day - day 0",
			bondName: "EDO083400",
			query:    "valuated_at=2025-12-06",
			accept:   "application/json",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid purchase day - day 32",
			bondName: "EDO083432",
			query:    "valuated_at=2025-12-06",
			accept:   "text/plain",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid name too short",
			bondName: "AB01",
			query:    "valuated_at=2025-12-06",
			accept:   "text/plain",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid valuated_at format",
			bondName: "EDO083412",
			query:    "valuated_at=not-a-date",
			accept:   "application/json",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "bond not found",
			bondName: "NONEXIST01",
			query:    "valuated_at=2025-12-06",
			accept:   "text/plain",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "valuation date before purchase date",
			bondName: "EDO093502",
			query:    "valuated_at=2025-09-01",
			accept:   "application/json",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/v1/bond/%s/valuation?%s", tt.bondName, tt.query)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Accept", tt.accept)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantCode, w.Body.String())
			}
		})
	}
}

func TestHandleValuation_AfterMaturity(t *testing.T) {
	server := loadTestServer(t)

	t.Run("plain text", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/bond/TOS112501/valuation?valuated_at=2025-12-01", nil)
		req.Header.Set("Accept", "text/plain")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200; body: %s", w.Code, w.Body.String())
		}
		if got := w.Body.String(); got != "121.99" {
			t.Errorf("got body %q, want %q", got, "121.99")
		}
	})

	t.Run("json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/bond/TOS112501/valuation?valuated_at=2025-12-01", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200; body: %s", w.Code, w.Body.String())
		}

		var resp ValuationResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}

		if math.Abs(resp.Price-121.99) > 1e-9 {
			t.Errorf("got price %v, want 121.99", resp.Price)
		}
		if resp.ValuatedAt != "2025-12-01" {
			t.Errorf("got valuated_at %q, want 2025-12-01", resp.ValuatedAt)
		}
	})
}

func TestHandleValuation_DefaultDate(t *testing.T) {
	server := loadTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/bond/TOS112501/valuation", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp ValuationResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.ValuatedAt == "" {
		t.Error("expected non-empty valuated_at when no date provided")
	}
	if resp.Price <= 0 {
		t.Errorf("expected positive price, got %v", resp.Price)
	}
}
