package domain

import "time"

func CalculateStatus(lastSeen time.Time, now time.Time, offlineAfter time.Duration) string {
	if lastSeen.IsZero() {
		return "unknown"
	}
	if now.Sub(lastSeen) <= offlineAfter {
		return "online"
	}
	if now.Sub(lastSeen) <= offlineAfter*3 {
		return "jitter"
	}
	return "offline"
}
