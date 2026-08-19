package application

func normalizeActions(actions []string) []string {
	if len(actions) == 0 {
		return []string{"log"}
	}
	out := make([]string, 0, len(actions))
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
