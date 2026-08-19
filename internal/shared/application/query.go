package application

import (
	"strings"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type ListQuery struct {
	shareddomain.PageQuery
}

func NewListQuery(values map[string][]string) (ListQuery, error) {
	q, err := shareddomain.ParsePageQuery(values)
	if err != nil {
		return ListQuery{}, err
	}
	return ListQuery{PageQuery: q}, nil
}

func AllowedSort(sort string, allowed map[string]bool, fallback string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return fallback
	}
	field := strings.TrimPrefix(sort, "-")
	if !allowed[field] {
		return fallback
	}
	return sort
}
