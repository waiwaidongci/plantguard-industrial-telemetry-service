package application

import (
	"strings"
)

func validName(name string) bool {
	return strings.TrimSpace(name) != ""
}
