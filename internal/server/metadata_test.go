package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleMetadata(t *testing.T) {
	server := loadTestServer(t)

	tests := []struct {
		name     string
		bondName string
		wantCode int
		wantISIN string
	}{
		{
			name:     "TOS bond without purchase day",
			bondName: "TOS1125",
			wantCode: http.StatusOK,
			wantISIN: "PL0000115143",
		},
		{
			name:     "TOS bond with purchase day",
			bondName: "TOS112501",
			wantCode: http.StatusOK,
			wantISIN: "PL0000115143",
		},
		{
			name:     "EDO bond without purchase day",
			bondName: "EDO0834",
			wantCode: http.StatusOK,
			wantISIN: "PL0000117164",
		},
		{
			name:     "EDO bond with purchase day",
			bondName: "EDO083412",
			wantCode: http.StatusOK,
			wantISIN: "PL0000117164",
		},
		{
			name:     "bond not found",
			bondName: "NONEXST",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/v1/bond/%s", tt.bondName)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("got status %d, want %d; body: %s", w.Code, tt.wantCode, w.Body.String())
			}

			if tt.wantCode == http.StatusOK {
				var resp MetadataResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if resp.ISIN != tt.wantISIN {
					t.Errorf("got ISIN %q, want %q", resp.ISIN, tt.wantISIN)
				}
				if resp.Name == "" {
					t.Error("expected non-empty name")
				}
				if resp.FaceValue == 0 {
					t.Error("expected non-zero face value")
				}
			}
		})
	}
}
