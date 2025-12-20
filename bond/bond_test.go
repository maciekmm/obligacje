package bond

import (
	"testing"
	"time"

	"github.com/maciekmm/obligacje/tz"
)

func TestBond_Period(t *testing.T) {
	tests := []struct {
		name string

		saleStart      time.Time
		saleEnd        time.Time
		frequency      CouponPaymentsFrequency
		maturityMonths int

		// period
		i            uint
		purchasedDay uint

		periodStart time.Time
		periodEnd   time.Time
		wantErr     bool
	}{
		{
			name:           "first period, bought on the first day",
			saleStart:      time.Date(2024, time.August, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.August, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              0,
			purchasedDay:   1,
			periodStart:    time.Date(2024, time.August, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2024, time.September, 1, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "first period, bought in the middle of the month",
			saleStart:      time.Date(2024, time.August, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.August, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              0,
			purchasedDay:   15,
			periodStart:    time.Date(2024, time.August, 15, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2024, time.September, 15, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "first period, bought on the last day, clamps to the last day of the next month",
			saleStart:      time.Date(2024, time.August, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.August, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              0,
			purchasedDay:   31,
			periodStart:    time.Date(2024, time.August, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2024, time.September, 30, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "second period, bought on the last day, clamps to the last day of the next month",
			saleStart:      time.Date(2024, time.August, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.August, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              1,
			purchasedDay:   31,
			periodStart:    time.Date(2024, time.September, 30, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2024, time.October, 31, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "second period, bought on the last day, clamps to the last day of the Feb if it's a leap year",
			saleStart:      time.Date(2024, time.January, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              0,
			purchasedDay:   31,
			periodStart:    time.Date(2024, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2024, time.February, 29, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "second period, bought on the last day, clamps to the last day of the Feb if it's not a leap year",
			saleStart:      time.Date(2025, time.January, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2025, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              0,
			purchasedDay:   31,
			periodStart:    time.Date(2025, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2025, time.February, 28, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "second period, bought on the last day, clamps to the last day of the Feb if it's not a leap year",
			saleStart:      time.Date(2025, time.January, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2025, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              0,
			purchasedDay:   31,
			periodStart:    time.Date(2025, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2025, time.February, 28, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "invalid period index",
			saleStart:      time.Date(2025, time.January, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2025, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyMonthly,
			maturityMonths: 12,
			i:              12,
			purchasedDay:   31,
			wantErr:        true,
		},
		{
			name:           "yearly bond, purchased on the last day of the Feb when leap year",
			saleStart:      time.Date(2024, time.February, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.February, 29, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyYearly,
			maturityMonths: 120,
			i:              0,
			purchasedDay:   29,
			periodStart:    time.Date(2024, time.February, 29, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2025, time.February, 28, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "yearly bond, second period",
			saleStart:      time.Date(2024, time.January, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyYearly,
			maturityMonths: 120,
			i:              1,
			purchasedDay:   31,
			periodStart:    time.Date(2025, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
			periodEnd:      time.Date(2026, time.January, 31, 0, 0, 0, 0, tz.WarsawTimezone),
		},
		{
			name:           "invalid purchase Day - 30th of Feb",
			saleStart:      time.Date(2024, time.February, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.February, 29, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyYearly,
			maturityMonths: 120,
			i:              1,
			purchasedDay:   30,
			wantErr:        true,
		},
		{
			name:           "invalid purchase Day - 0",
			saleStart:      time.Date(2024, time.February, 1, 0, 0, 0, 0, tz.WarsawTimezone),
			saleEnd:        time.Date(2024, time.February, 29, 0, 0, 0, 0, tz.WarsawTimezone),
			frequency:      CouponPaymentsFrequencyYearly,
			maturityMonths: 120,
			i:              1,
			purchasedDay:   0,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Bond{
				SaleStart:               tt.saleStart,
				SaleEnd:                 tt.saleEnd,
				CouponPaymentsFrequency: tt.frequency,
				MonthsToMaturity:        tt.maturityMonths,
			}
			start, end, err := b.Period(tt.i, tt.purchasedDay)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("Period() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !start.Equal(tt.periodStart) {
				t.Errorf("Period() = %v, want %v", start, tt.periodStart)
			}
			if !end.Equal(tt.periodEnd) {
				t.Errorf("Period() = %v, want %v", end, tt.periodEnd)
			}
		})
	}
}
