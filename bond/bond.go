package bond

type Price int64      // In 0.01 units
type Percentage int64 // In 0.01%

type InterestRecalculation string

var (
	InterestRecalculationMonthly InterestRecalculation = "monthly"
	InterestRecalculationYearly  InterestRecalculation = "yearly"
	InterestRecalculationNone    InterestRecalculation = "none"
	InterestRecalculationUnknown InterestRecalculation = "unknown"
)

type Bond struct {
	Series                string
	ISIN                  string
	Price                 Price
	ExchangePrice         Price
	MarginPercentage      Percentage
	InterestPeriods       []Percentage
	BuyoutMonths          int
	InterestRecalculation InterestRecalculation
}
