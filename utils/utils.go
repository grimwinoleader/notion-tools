package utils

import "time"

const (
	YYYYMMDD = "2006-01-02"
)

func StartOfWeek(date time.Time) time.Time {
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}
