package valid

import "time"

func IsValidTime(t time.Time) bool {
	hour, min, _ := t.Clock()
	return hour >= 0 && hour < 24 && min >= 0 && min < 60
}
