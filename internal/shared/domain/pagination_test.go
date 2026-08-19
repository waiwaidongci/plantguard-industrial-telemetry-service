package domain

import (
	"net/url"
	"testing"
)

func TestParsePageQuery(t *testing.T) {
	query, err := ParsePageQuery(url.Values{
		"page":        []string{"2"},
		"page_size":   []string{"25"},
		"filter_name": []string{"pump"},
		"sort":        []string{"-created_at"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if query.Page != 2 || query.PageSize != 25 || query.Filters["name"] != "pump" {
		t.Fatalf("unexpected query: %#v", query)
	}
	if query.Offset() != 25 || query.Limit() != 25 {
		t.Fatalf("unexpected offset/limit: %d %d", query.Offset(), query.Limit())
	}
}
