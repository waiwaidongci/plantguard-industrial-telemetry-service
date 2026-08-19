package domain

import "strings"

func RequiredString(value string) bool {
	return strings.TrimSpace(value) != ""
}

func NormalizeString(value string) string {
	return strings.TrimSpace(value)
}

func OneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
