package utils

import "time"

// Returns a duration with an input of value and a unit
// unit: s = second, M = min, d = day, w = week, m = month, y = year
func GetDuration(value int64, unit string) (duration time.Duration) {
	if unit == "" || unit == "s" {
		duration = time.Duration(value) * time.Second
	} else if unit == "M" {
		duration = time.Duration(value) * time.Minute
	} else if unit == "d" {
		duration = time.Duration(value*1440) * time.Minute
	} else if unit == "w" {
		duration = time.Duration(value*1440*7) * time.Minute
	} else if unit == "m" {
		duration = time.Duration(value*1440*30) * time.Minute
	} else if unit == "y" {
		duration = time.Duration(value*1440*365) * time.Minute
	} else {
		duration = time.Duration(value)
	}
	return
}
