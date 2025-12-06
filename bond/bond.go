package bond

type Price int64      // In 0.01 units
type Percentage int64 // In 0.01%

type InterestRecalculation string

const (
	InterestRecalculationMonthly InterestRecalculation = "monthly"
	InterestRecalculationYearly  InterestRecalculation = "yearly"
	InterestRecalculationNone    InterestRecalculation = "none"

	InterestRecalculationUnknown InterestRecalculation = "unknown"
)

type Bond struct {
	Series string
	ISIN   string

	FaceValue        Price
	MonthsToMaturity int
	ExchangePrice    Price

	Margin                Percentage
	InterestPeriods       []Percentage
	InterestRecalculation InterestRecalculation
}
