package application

func normalizeActions(actions []string) []string {
	if len(actions) == 0 {
		return []string{"log"}
	}
	out := actions[:0]
	for _, action := range actions {
		if action == "log" || action == "webhook" || action == "maintenance" {
			out = append(out, action)
		}
	}
	if len(out) == 0 {
		return []string{"log"}
	}
	return out
}
