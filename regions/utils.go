package regions

import (
	"time"
)

// RangeDate returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true.
func RangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, start.Location())
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, start.Location())

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}

		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}
