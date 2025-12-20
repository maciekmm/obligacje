package bondxls

import (
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/internal/testutil"
	"github.com/maciekmm/obligacje/tz"
)

func TestXLSRepository_Lookup(t *testing.T) {
	tests := []struct {
		series  string
		want    bond.Bond
		wantErr bool
	}{
		{
			series:  "NONEXISTENT",
			want:    bond.Bond{},
			wantErr: true,
		},
		{
			series: "OTS0118",
			want: bond.Bond{
				Series:                  "OTS0118",
				ISIN:                    "PL0000110292",
				FaceValue:               100.00,
				ExchangePrice:           0.00,
				Margin:                  0.00,
				MonthsToMaturity:        3,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyNone,
				InterestPeriods:         []bond.Percentage{0.0150},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2017-10-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2017-10-31", tz.WarsawTimezone)),
			},
		},
		{
			series: "ROR0623",
			want: bond.Bond{
				Series:                  "ROR0623",
				ISIN:                    "PL0000114716",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				Margin:                  0.00,
				MonthsToMaturity:        12,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyMonthly,
				InterestPeriods:         []bond.Percentage{0.0525, 0.0600, 0.0650, 0.0650, 0.0675, 0.0675, 0.0675, 0.0675, 0.0675, 0.0675, 0.0675, 0.0675},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2022-06-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2022-06-30", tz.WarsawTimezone)),
			},
		},
		{
			series: "ROR1226",
			want: bond.Bond{
				Series:                  "ROR1226",
				ISIN:                    "PL0000118626",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				Margin:                  0.00,
				MonthsToMaturity:        12,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyMonthly,
				InterestPeriods:         []bond.Percentage{0.0425},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2025-12-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2025-12-31", tz.WarsawTimezone)),
			},
		},
		{
			series: "TOS0825",
			want: bond.Bond{
				Series:                  "TOS0825",
				ISIN:                    "PL0000113890",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				Margin:                  0.00,
				MonthsToMaturity:        36,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyNone,
				InterestPeriods:         []bond.Percentage{0.0650},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2022-08-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2022-08-31", tz.WarsawTimezone)),
			},
		},
		{
			series: "TOS0728",
			want: bond.Bond{
				Series:                  "TOS0728",
				ISIN:                    "PL0000118220",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				InterestPeriods:         []bond.Percentage{0.0565},
				MonthsToMaturity:        36,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyNone,
				Margin:                  0.00,

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2025-07-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2025-07-31", tz.WarsawTimezone)),
			},
		},
		{
			series: "EDO0531",
			want: bond.Bond{
				Series:                  "EDO0531",
				ISIN:                    "PL0000113684",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				Margin:                  0.01,
				MonthsToMaturity:        120,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyYearly,
				InterestPeriods:         []bond.Percentage{0.0170, 0.1200, 0.1710, 0.0300, 0.0590},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2021-05-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2021-05-31", tz.WarsawTimezone)),
			},
		},
		{
			series: "EDO1235",
			want: bond.Bond{
				Series:                  "EDO1235",
				ISIN:                    "PL0000118667",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				Margin:                  0.02,
				MonthsToMaturity:        120,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyYearly,
				InterestPeriods:         []bond.Percentage{0.056},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2025-12-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2025-12-31", tz.WarsawTimezone)),
			},
		},
		{
			series: "ROR0126",
			want: bond.Bond{
				Series:                  "ROR0126",
				ISIN:                    "PL0000117552",
				FaceValue:               100.00,
				ExchangePrice:           99.90,
				MonthsToMaturity:        12,
				CouponPaymentsFrequency: bond.CouponPaymentsFrequencyMonthly,
				InterestPeriods:         []bond.Percentage{0.0575, 0.0575, 0.0575, 0.0575, 0.0575, 0.0525, 0.0525, 0.0500, 0.0500, 0.0475, 0.0450, 0.0425},

				SaleStart: testutil.Must(time.ParseInLocation(time.DateOnly, "2025-01-01", tz.WarsawTimezone)),
				SaleEnd:   testutil.Must(time.ParseInLocation(time.DateOnly, "2025-01-31", tz.WarsawTimezone)),
			},
		},
	}
	r, err := LoadFromXLSX(slog.New(slog.NewTextHandler(os.Stdout, nil)), filepath.Join(testutil.TestDataDirectory(), "data.xlsx"))
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
		return
	}
	for _, tt := range tests {
		t.Run(tt.series, func(t *testing.T) {
			got, err := r.Lookup(tt.series)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Lookup() got = \n%+v, want = \n%+v", got, tt.want)
			}
		})
	}
}

func equal(a, b bond.Bond) bool {
	if a.Series != b.Series {
		return false
	}
	if a.ISIN != b.ISIN {
		return false
	}
	if !floatEqual(float64(a.FaceValue), float64(b.FaceValue)) {
		return false
	}
	if a.MonthsToMaturity != b.MonthsToMaturity {
		return false
	}
	if !floatEqual(float64(a.ExchangePrice), float64(b.ExchangePrice)) {
		return false
	}
	if !floatEqual(float64(a.Margin), float64(b.Margin)) {
		return false
	}
	if a.CouponPaymentsFrequency != b.CouponPaymentsFrequency {
		return false
	}
	if len(a.InterestPeriods) != len(b.InterestPeriods) {
		return false
	}
	for i := range a.InterestPeriods {
		if !floatEqual(float64(a.InterestPeriods[i]), float64(b.InterestPeriods[i])) {
			return false
		}
	}
	if !a.SaleStart.Equal(b.SaleStart) {
		return false
	}
	if !a.SaleEnd.Equal(b.SaleEnd) {
		return false
	}
	return true
}

func floatEqual(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) <= epsilon
}
