package domain

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func StartOfMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func StartOfHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}
