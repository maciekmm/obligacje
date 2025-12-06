package calculator

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/bondxls"
	"github.com/maciekmm/obligacje/internal/testutil"
)

var (
	warsawTimezone = MustLoadLocation("Europe/Warsaw")
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
	repo, err := bondxls.LoadFromXLSX(xlsxFile)
	if err != nil {
		panic(err)
	}
	return repo
}

func TestCalculator_Calculate(t *testing.T) {
	type args struct {
		series      string
		purchasedAt time.Time
		valuatedAt  time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    bond.Price
		wantErr bool
	}{
		{
			name: "Known bond with known valuation",
			args: args{
				series:      "EDO0834",
				purchasedAt: time.Date(2024, time.August, 8, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			want:    10887,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation, two interest periods",
			args: args{
				series:      "EDO0832",
				purchasedAt: time.Date(2022, time.August, 20, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2024, time.August, 12, 0, 0, 0, 0, warsawTimezone),
			},
			want:    11995,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				series:      "EDO0935",
				purchasedAt: time.Date(2025, time.September, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			want:    10156,
			wantErr: false,
		},
		{
			name: "Bond purchased in a different month than the series suggests",
			args: args{
				series:      "EDO0935",
				purchasedAt: time.Date(2025, time.October, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Bond purchased in the future",
			args: args{
				series:      "EDO0935",
				purchasedAt: time.Date(2026, time.September, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Bond series does not exist (invalid month)",
			args: args{
				series:      "EDO1335",
				purchasedAt: time.Date(2024, time.September, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Bond series does not exist (invalid year)",
			args: args{
				series:      "EDO0965",
				purchasedAt: time.Date(2024, time.September, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Bond series does not exist (unknown series)",
			args: args{
				series:      "SEP1335",
				purchasedAt: time.Date(2024, time.September, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, warsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Valuation date is before purchase date",
			args: args{
				series:      "SEP1335",
				purchasedAt: time.Date(2024, time.September, 2, 0, 0, 0, 0, warsawTimezone),
				valuatedAt:  time.Date(2024, time.August, 2, 0, 0, 0, 0, warsawTimezone),
			},
			wantErr: true,
		},
	}

	c := NewCalculator(LoadBondRepository())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Calculate(tt.args.series, tt.args.purchasedAt, tt.args.valuatedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Calculate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
