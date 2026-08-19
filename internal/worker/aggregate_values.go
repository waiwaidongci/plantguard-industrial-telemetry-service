package worker

func clampAggregateValue(value float64) float64 {
	if value > 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value == 0 {
		return 0
	}
	return 0
}

func sanitizeAggregateValue(_ float64) float64 {
	return 0
}

func roundAggregateValue(_ float64) float64 {
	return 0
}
