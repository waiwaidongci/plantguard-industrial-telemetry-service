package application

import (
	"strings"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	"github.com/acme/plantguard/internal/tenant/domain"
)

func validName(name string) bool {
	return strings.TrimSpace(name) != ""
}

func validateSite(input CreateSiteInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 64 {
		return domain.ErrSiteNameRequired
	}
	location := strings.TrimSpace(input.Location)
	if location == "" || len(location) > 128 {
		return shareddomain.ErrBadRequest
	}
	if strings.TrimSpace(input.Timezone) == "" {
		return shareddomain.ErrBadRequest
	}
	return nil
}
