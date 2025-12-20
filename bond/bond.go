package bond

import (
	"fmt"
	"time"

	"github.com/maciekmm/obligacje/tz"
)

type Price float64
type Percentage float64

type CouponPaymentsFrequency int

const (
	CouponPaymentsFrequencyMonthly CouponPaymentsFrequency = 12
	CouponPaymentsFrequencyYearly  CouponPaymentsFrequency = 1
	CouponPaymentsFrequencyNone    CouponPaymentsFrequency = 1
	CouponPaymentsFrequencyUnknown CouponPaymentsFrequency = -1
)

func (cpf CouponPaymentsFrequency) Months() int {
	return int(12 / cpf)
}

type Bond struct {
	Series string
	ISIN   string

	FaceValue        Price
	MonthsToMaturity int
	ExchangePrice    Price

	Margin                  Percentage
	InterestPeriods         []Percentage
	CouponPaymentsFrequency CouponPaymentsFrequency

	SaleStart time.Time
	SaleEnd   time.Time
}

func (b Bond) Period(i int, purchaseDay int) (time.Time, time.Time, error) {
	if b.CouponPaymentsFrequency == CouponPaymentsFrequencyUnknown {
		return time.Time{}, time.Time{}, fmt.Errorf("unknown coupon payments frequency")
	}

	if i < 0 || i >= b.InterestPeriodCount() {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period index: %d", i)
	}

	if purchaseDay < 1 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid purchase day: %d", purchaseDay)
	}

	lastDayOfPurchaseMonth := lastDayOfMonth(b.SaleStart.Year(), b.SaleStart.Month())
	if purchaseDay > lastDayOfPurchaseMonth {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid purchase day: %d", purchaseDay)
	}

	purchasedAt := time.Date(b.SaleStart.Year(), b.SaleStart.Month(), purchaseDay, 0, 0, 0, 0, tz.WarsawTimezone)

	startAt := time.Date(purchasedAt.Year(), purchasedAt.Month()+time.Month(i*b.CouponPaymentsFrequency.Months()), 1, 0, 0, 0, 0, tz.WarsawTimezone)
	lastDayOfPeriodStartDay := lastDayOfMonth(startAt.Year(), startAt.Month())
	if purchaseDay > lastDayOfPeriodStartDay {
		startAt = startAt.AddDate(0, 0, lastDayOfPeriodStartDay-1)
	} else {
		startAt = startAt.AddDate(0, 0, purchaseDay-1)
	}

	endAt := time.Date(purchasedAt.Year(), purchasedAt.Month()+time.Month((i+1)*b.CouponPaymentsFrequency.Months()), 1, 0, 0, 0, 0, tz.WarsawTimezone)
	lastDayOfPeriodEndDay := lastDayOfMonth(endAt.Year(), endAt.Month())
	if purchaseDay > lastDayOfPeriodEndDay {
		endAt = endAt.AddDate(0, 0, lastDayOfPeriodEndDay-1)
	} else {
		endAt = endAt.AddDate(0, 0, purchaseDay-1)
	}

	return startAt, endAt, nil
}

func (b Bond) InterestPeriodCount() int {
	return b.MonthsToMaturity / b.CouponPaymentsFrequency.Months()
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day()
}
