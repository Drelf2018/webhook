package utils

import "time"

func NextTimeDuration(hour, min, sec int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, now.Location())
	switch {
	case now.Hour() > hour,
		now.Hour() == hour && now.Minute() > min,
		now.Hour() == hour && now.Minute() == min && now.Second() > sec:
		next = next.AddDate(0, 0, 1)
	}
	return time.Until(next)
}
