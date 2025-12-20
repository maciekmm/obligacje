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
				purchasedAt: time.Date(2024, time.August, 8, 0, 0, 0, 0, tz.WarsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, tz.WarsawTimezone),
			},
			want:    108.87,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation, two interest periods",
			args: args{
				series:      "EDO0832",
				purchasedAt: time.Date(2022, time.August, 20, 0, 0, 0, 0, tz.WarsawTimezone),
				valuatedAt:  time.Date(2024, time.August, 12, 0, 0, 0, 0, tz.WarsawTimezone),
			},
			want:    119.95,
			wantErr: false,
		},
		{
			name: "Known bond with known valuation",
			args: args{
				series:      "EDO0935",
				purchasedAt: time.Date(2025, time.September, 2, 0, 0, 0, 0, tz.WarsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, tz.WarsawTimezone),
			},
			want:    101.56,
			wantErr: false,
		},
		{
			name: "Bond purchased in a different month than the series suggests",
			args: args{
				series:      "EDO0935",
				purchasedAt: time.Date(2025, time.October, 2, 0, 0, 0, 0, tz.WarsawTimezone),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, tz.WarsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Bond purchased in the future",
			args: args{
				series:      "EDO0935",
				purchasedAt: time.Now().AddDate(1, 0, 0),
				valuatedAt:  time.Date(2025, time.December, 6, 0, 0, 0, 0, tz.WarsawTimezone),
			},
			wantErr: true,
		},
		{
			name: "Valuation date is before purchase date",
			args: args{
				series:      "SEP1335",
				purchasedAt: time.Date(2024, time.September, 2, 0, 0, 0, 0, tz.WarsawTimezone),
				valuatedAt:  time.Date(2024, time.August, 2, 0, 0, 0, 0, tz.WarsawTimezone),
			},
			wantErr: true,
		},
	}

	c := NewCalculator()
	repo := LoadBondRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bond, err := repo.Lookup(tt.args.series)
			if err != nil {
				t.Errorf("Lookup() error = %v", err)
				return
			}
			got, err := c.Calculate(bond, tt.args.purchasedAt, tt.args.valuatedAt)
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
