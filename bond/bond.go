package bond

import "time"

type Price float64
type Percentage float64

type CouponPaymentsFrequency int

const (
	CouponPaymentsFrequencyMonthly CouponPaymentsFrequency = 12
	CouponPaymentsFrequencyYearly  CouponPaymentsFrequency = 1
	CouponPaymentsFrequencyNone    CouponPaymentsFrequency = 1
	CouponPaymentsFrequencyUnknown CouponPaymentsFrequency = 0
)

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

func (b Bond) Period(i uint, purchasedDay uint) (time.Time, time.Time, error) {
	return time.Time{}, time.Time{}, nil
}
