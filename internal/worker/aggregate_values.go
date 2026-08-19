package worker

import "math"

func clampAggregateValue(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func sanitizeAggregateValue(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func roundAggregateValue(value float64) float64 {
	return math.Round(value*100) / 100
}
