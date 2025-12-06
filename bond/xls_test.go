package bond

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/maciekmm/obligacje/internal/testutil"
)

func TestXLSRepository_Lookup(t *testing.T) {
	tests := []struct {
		series  string
		want    Bond
		wantErr bool
	}{
		{
			series:  "NONEXISTENT",
			want:    Bond{},
			wantErr: true,
		},
		{
			series: "OTS0118",
			want: Bond{
				Series:                "OTS0118",
				ISIN:                  "PL0000110292",
				Price:                 10000,
				ExchangePrice:         0,
				MarginPercentage:      0,
				BuyoutInMonths:        3,
				InterestRecalculation: InterestRecalculationNone,
				InterestPeriods:       []Percentage{150},
			},
		},
		{
			series: "ROR0623",
			want: Bond{
				Series:                "ROR0623",
				ISIN:                  "PL0000114716",
				Price:                 10000,
				ExchangePrice:         9990,
				MarginPercentage:      0,
				BuyoutInMonths:        12,
				InterestRecalculation: InterestRecalculationMonthly,
				InterestPeriods:       []Percentage{525, 600, 650, 650, 675, 675, 675, 675, 675, 675, 675, 675},
			},
		},
		{
			series: "ROR1226",
			want: Bond{
				Series:                "ROR1226",
				ISIN:                  "PL0000118626",
				Price:                 10000,
				ExchangePrice:         9990,
				MarginPercentage:      0,
				BuyoutInMonths:        12,
				InterestRecalculation: InterestRecalculationMonthly,
				InterestPeriods:       []Percentage{425},
			},
		},
		{
			series: "TOS0825",
			want: Bond{
				Series:                "TOS0825",
				ISIN:                  "PL0000113890",
				Price:                 10000,
				ExchangePrice:         9990,
				MarginPercentage:      0,
				BuyoutInMonths:        36,
				InterestRecalculation: InterestRecalculationNone,
				InterestPeriods:       []Percentage{650},
			},
		},
		{
			series: "TOS0728",
			want: Bond{
				Series:                "TOS0728",
				ISIN:                  "PL0000118220",
				Price:                 10000,
				ExchangePrice:         9990,
				InterestPeriods:       []Percentage{565},
				BuyoutInMonths:        36,
				InterestRecalculation: InterestRecalculationNone,
				MarginPercentage:      0,
			},
		},
		{
			series: "EDO0531",
			want: Bond{
				Series:                "EDO0531",
				ISIN:                  "PL0000113684",
				Price:                 10000,
				ExchangePrice:         9990,
				MarginPercentage:      100,
				BuyoutInMonths:        120,
				InterestRecalculation: InterestRecalculationYearly,
				InterestPeriods:       []Percentage{170, 1200, 1710, 300, 590},
			},
		},
		{
			series: "ROR0126",
			want: Bond{
				Series:                "ROR0126",
				ISIN:                  "PL0000117552",
				Price:                 10000,
				ExchangePrice:         9990,
				BuyoutInMonths:        12,
				InterestRecalculation: InterestRecalculationMonthly,
				InterestPeriods:       []Percentage{575, 575, 575, 575, 575, 525, 525, 500, 500, 475, 450, 425},
			},
		},
	}
	r, err := LoadFromXLSX(filepath.Join(testutil.TestDataDirectory(), "data.xlsx"))
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lookup() got = %v, want %v", got, tt.want)
			}
		})
	}
}
