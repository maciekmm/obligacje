package calculator

import (
	"log/slog"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/bondxls"
	"github.com/maciekmm/obligacje/internal/testutil"
	"github.com/maciekmm/obligacje/tz"
)

func MustLoadLocation(name string) *time.Location {
	if loc, err := time.LoadLocation(name); err != nil {
		panic(err)
	} else {
		return loc
	}
}

func LoadBondRepository() bond.Repository {
	xlsxFile := filepath.Join(testutil.TestDataDirectory(), "data.xlsx")
	repo, err := bondxls.LoadFromXLSX(slog.Default(), xlsxFile)
	if err != nil {
		panic(err)
	}
	return repo
}

func TestCalculator_Calculate(t *testing.T) {
	type args struct {
		name        string
		purchaseDay int
		valuatedAt  time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    bond.Price
		wantErr bool
	}{
		// OTS uses different calculation logic and returns slightly incorrect data
		// API won't support OTS unless fixed
		// {
		// 	name: "OTS bond at the end of valuation",
		// 	args: args{
		// 		name:      "OTS0825",
		// 		purchasedAt: time.Date(2025, time.May, 1, 0, 0, 0, 0, tz.WarsawTimezone),
		// 		valuatedAt:  time.Date(2025, time.August, 1, 0, 0, 0, 0, tz.WarsawTimezone),
		// 	},
		// 	want:    100.76,
		// 	wantErr: false,
		// },
		{
			name: "TOS bond before DST change",
			args: args{
				name:        "TOS1125",
				purchaseDay: 1,
				valuatedAt:  time.Date(2023, time.March, 26, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    102.72,
			wantErr: false,
		},
		{
			name: "TOS bond after DST change",
			args: args{
				name:        "TOS1125",
				purchaseDay: 1,
				valuatedAt:  time.Date(2023, time.March, 27, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    102.74,
			wantErr: false,
		},
		{
			name: "TOS bond end of valuation",
			args: args{
				name:        "TOS1125",
				purchaseDay: 1,
				valuatedAt:  time.Date(2025, time.November, 1, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    121.99,
			wantErr: false,
		},
		{
			name: "TOS bond towards end of valuation",
			args: args{
				name:        "TOS1125",
				purchaseDay: 1,
				valuatedAt:  time.Date(2025, time.April, 13, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    117.66,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				name:        "EDO0834",
				purchaseDay: 12,
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    108.87,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				name:        "EDO0834",
				purchaseDay: 12,
				valuatedAt:  time.Date(2025, time.December, 20, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    109.12,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation - first day of mth",
			args: args{
				name:        "EDO0834",
				purchaseDay: 1,
				valuatedAt:  time.Date(2025, time.August, 1, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    106.80,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				name:        "EDO0834",
				purchaseDay: 1,
				valuatedAt:  time.Date(2025, time.August, 2, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    106.82,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				name:        "EDO0834",
				purchaseDay: 1,
				valuatedAt:  time.Date(2025, time.August, 22, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    107.17,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				name:        "EDO0834",
				purchaseDay: 2,
				valuatedAt:  time.Date(2025, time.August, 23, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    107.17,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation, two interest periods",
			args: args{
				name:        "EDO0832",
				purchaseDay: 20,
				valuatedAt:  time.Date(2024, time.August, 9, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    119.95,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				name:        "EDO0935",
				purchaseDay: 2,
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			want:    101.56,
			wantErr: false,
		},
		{
			name: "Valuation date is before purchase date",
			args: args{
				name:        "EDO0935",
				purchaseDay: 2,
				valuatedAt:  time.Date(2025, time.September, 1, 0, 0, 0, 0, tz.UnifiedTimezone),
			},
			wantErr: true,
		},
	}

	c := NewCalculator()
	repo := LoadBondRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bond, err := repo.Lookup(tt.args.name)
			if err != nil {
				t.Errorf("Lookup() error = %v", err)
				return
			}
			got, err := c.Calculate(bond, tt.args.purchaseDay, tt.args.valuatedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if math.Abs(float64(tt.want)-float64(got)) > 1e-9 {
				t.Errorf("Calculate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
