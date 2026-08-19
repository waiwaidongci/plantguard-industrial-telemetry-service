package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 200
)

type Page struct {
	Items      []any `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
}

type PageQuery struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
}

func ParsePageQuery(values url.Values) (PageQuery, error) {
	q := PageQuery{Page: 1, PageSize: DefaultPageSize, Filters: map[string]string{}}
	if raw := values.Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return q, New("invalid_pagination", "page must be a positive integer", 400)
		}
		q.Page = n
	}
	if raw := values.Get("page_size"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > MaxPageSize {
			return q, New("invalid_pagination", fmt.Sprintf("page_size must be between 1 and %d", MaxPageSize), 400)
		}
		q.PageSize = n
	}
	q.Sort = strings.TrimSpace(values.Get("sort"))
	for key, vals := range values {
		if strings.HasPrefix(key, "filter_") && len(vals) > 0 {
			q.Filters[strings.TrimPrefix(key, "filter_")] = vals[0]
		}
	}
	return q, nil
}

func (q PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

func (q PageQuery) Limit() int {
	return q.PageSize
}

func TotalPages(total int64, size int) int {
	if size <= 0 {
		return 0
	}
	return int((total + int64(size) - 1) / int64(size))
}

func NewPage(items []any, query PageQuery, total int64) Page {
	pages := TotalPages(total, query.PageSize)
	return Page{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
		TotalPages: pages,
		HasNext:    query.Page < pages,
	}
}
