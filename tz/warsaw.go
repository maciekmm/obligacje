package tz

import "time"

var (
	WarsawTimezone *time.Location
)

func init() {
	var err error
	WarsawTimezone, err = time.LoadLocation("Europe/Warsaw")
	if err != nil {
		panic(err)
	}
}
