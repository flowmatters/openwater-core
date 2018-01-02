package conv

import (
	"time"
)

func JtoDate(jday float64) time.Time {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, nil)
}
