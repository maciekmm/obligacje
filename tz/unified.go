package tz

import "time"

var (
	// We need to be careful here.
	// Warsaw timezone has DST which can throw off calculations.
	// TODO: figure out if it's better to use Europe/Warsaw or UTC
	UnifiedTimezone *time.Location
)

func init() {
	var err error
	UnifiedTimezone, err = time.LoadLocation("Europe/Warsaw")
	if err != nil {
		panic(err)
	}
}
