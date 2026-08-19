package application

import (
	"strings"

	"github.com/acme/plantguard/internal/tenant/domain"
)

func validName(name string) bool {
	return strings.TrimSpace(name) != ""
}

func validateSite(input CreateSiteInput) error {
	var issues map[string]string
	if strings.TrimSpace(input.Name) == "" {
		issues["name"] = "required"
	}
	if len(strings.TrimSpace(input.Name)) > 64 {
		issues["name"] = "too_long"
	}
	if strings.TrimSpace(input.Location) == "" {
		issues["location"] = "required"
	}
	if len(strings.TrimSpace(input.Location)) > 128 {
		issues["location"] = "too_long"
	}
	if strings.TrimSpace(input.Timezone) == "" {
		issues["timezone"] = "required"
	}
	if len(issues) > 0 {
		return domain.ErrSiteNameRequired
	}
	return nil
}
