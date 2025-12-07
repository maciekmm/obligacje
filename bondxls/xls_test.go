package bondxls

import (
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/internal/testutil"
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
				Series:                "OTS0118",
				ISIN:                  "PL0000110292",
				FaceValue:             10000,
				ExchangePrice:         0,
				Margin:                0,
				MonthsToMaturity:      3,
				InterestRecalculation: bond.InterestRecalculationNone,
				InterestPeriods:       []bond.Percentage{150},
			},
		},
		{
			series: "ROR0623",
			want: bond.Bond{
				Series:                "ROR0623",
				ISIN:                  "PL0000114716",
				FaceValue:             10000,
				ExchangePrice:         9990,
				Margin:                0,
				MonthsToMaturity:      12,
				InterestRecalculation: bond.InterestRecalculationMonthly,
				InterestPeriods:       []bond.Percentage{525, 600, 650, 650, 675, 675, 675, 675, 675, 675, 675, 675},
			},
		},
		{
			series: "ROR1226",
			want: bond.Bond{
				Series:                "ROR1226",
				ISIN:                  "PL0000118626",
				FaceValue:             10000,
				ExchangePrice:         9990,
				Margin:                0,
				MonthsToMaturity:      12,
				InterestRecalculation: bond.InterestRecalculationMonthly,
				InterestPeriods:       []bond.Percentage{425},
			},
		},
		{
			series: "TOS0825",
			want: bond.Bond{
				Series:                "TOS0825",
				ISIN:                  "PL0000113890",
				FaceValue:             10000,
				ExchangePrice:         9990,
				Margin:                0,
				MonthsToMaturity:      36,
				InterestRecalculation: bond.InterestRecalculationNone,
				InterestPeriods:       []bond.Percentage{650},
			},
		},
		{
			series: "TOS0728",
			want: bond.Bond{
				Series:                "TOS0728",
				ISIN:                  "PL0000118220",
				FaceValue:             10000,
				ExchangePrice:         9990,
				InterestPeriods:       []bond.Percentage{565},
				MonthsToMaturity:      36,
				InterestRecalculation: bond.InterestRecalculationNone,
				Margin:                0,
			},
		},
		{
			series: "EDO0531",
			want: bond.Bond{
				Series:                "EDO0531",
				ISIN:                  "PL0000113684",
				FaceValue:             10000,
				ExchangePrice:         9990,
				Margin:                100,
				MonthsToMaturity:      120,
				InterestRecalculation: bond.InterestRecalculationYearly,
				InterestPeriods:       []bond.Percentage{170, 1200, 1710, 300, 590},
			},
		},
		{
			series: "ROR0126",
			want: bond.Bond{
				Series:                "ROR0126",
				ISIN:                  "PL0000117552",
				FaceValue:             10000,
				ExchangePrice:         9990,
				MonthsToMaturity:      12,
				InterestRecalculation: bond.InterestRecalculationMonthly,
				InterestPeriods:       []bond.Percentage{575, 575, 575, 575, 575, 525, 525, 500, 500, 475, 450, 425},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lookup() got = %v, want %v", got, tt.want)
			}
		})
	}
}
